import axios from "axios";

// 使用相对路径 /api/，由 nginx 反向代理到后端 GolangApiServer:5000，
// 这样通过任意 IP/域名访问（含局域网 IP）都能正确连上后端，不再依赖构建时写死的地址。
const API = axios.create({ baseURL: '/api/' });

// 请求拦截器：为每个请求附加 Bearer token（从 localStorage 读取）
API.interceptors.request.use((req) => {
    if (localStorage.getItem('profile')) {
        req.headers.Authorization = `Bearer ${JSON.parse(localStorage.getItem('profile')).token}`;
    }
    return req;
})

// ---------------------------------------------------------------------------
//  GET 请求缓存 + 并发去重
//  背景：前端在挂载 / watcher 触发时会对相同接口发出大量重复请求，导致后端 429。
//  这里做两层防护：
//   1) 并发去重：同一时间对相同 GET 只发一次，其余复用同一个 Promise。
//   2) 短 TTL 缓存：一段时间内再次请求直接用缓存，不再打后端。
//  写请求成功后清空 GET 缓存，保证下次拿到的是新数据。
// ---------------------------------------------------------------------------

// 缓存条目的默认存活时间 (ms)：列表/用户这类数据 8s 足够新鲜。
const GET_TTL = 8_000;

const GET_CACHE = new Map();    // key -> { response, expiresAt }
const GET_INFLIGHT = new Map(); // key -> Promise

function sortObject(value) {
  if (Array.isArray(value)) return value.map(sortObject);
  if (value && typeof value === "object") {
    return Object.keys(value)
      .sort()
      .reduce((acc, key) => {
        acc[key] = sortObject(value[key]);
        return acc;
      }, {});
  }
  return value;
}

function getCacheKey(method, url, config = {}) {
  // 取当前登录 token 作为 key 的一部分，避免不同用户间串缓存
  let auth = "";
  try {
    const profile = JSON.parse(localStorage.getItem("profile") || "null");
    auth = profile?.token || "";
  } catch (error) {
    auth = "";
  }
  const params = config.params ? JSON.stringify(sortObject(config.params)) : "";
  return `${method} ${url} ${auth} ${params}`;
}

function cachedGet(url, config = {}) {
  const method = "get";
  const key = getCacheKey(method, url, config);

  // 1) 命中未过期缓存
  const hit = GET_CACHE.get(key);
  if (hit && hit.expiresAt > Date.now()) {
    return Promise.resolve(hit.response);
  }

  // 2) 并发去重：相同请求正在发送，直接复用
  if (GET_INFLIGHT.has(key)) {
    return GET_INFLIGHT.get(key);
  }

  const promise = _get(url, config)
    .then((response) => {
      // 单个请求可用 cacheTTL 覆盖，0 表示不缓存
      const ttl = config.cacheTTL !== undefined ? config.cacheTTL : GET_TTL;
      if (ttl > 0) {
        GET_CACHE.set(key, { response, expiresAt: Date.now() + ttl });
      }
      return response;
    })
    .finally(() => {
      GET_INFLIGHT.delete(key);
    });

  GET_INFLIGHT.set(key, promise);
  return promise;
}

// 记录原始方法，避免递归
const _get = API.get.bind(API);

// 拦截所有 GET：走缓存/去重
API.get = (url, config = {}) => cachedGet(url, config);

// 写请求成功后清空 GET 缓存，确保下一次读取是最新数据
["post", "put", "patch", "delete"].forEach((method) => {
  const original = API[method].bind(API);
  API[method] = (url, data, config = {}) =>
    original(url, data, config).then((response) => {
      GET_CACHE.clear();
      return response;
    });
});

// ---------------------------------------------------------------------------
//  429 退避重试：读 Retry-After，没有则用固定退避，最多重试一次
// ---------------------------------------------------------------------------
API.interceptors.response.use(
  (response) => response,
  async (error) => {
    const config = error.config;
    if (error.response?.status === 429 && config && !config.__retried) {
      config.__retried = true;
      const retryAfter = error.response.headers["retry-after"];
      const delay = retryAfter ? Number(retryAfter) * 1000 : 1500;
      await new Promise((resolve) => setTimeout(resolve, delay));
      return API(config);
    }
    return Promise.reject(error);
  }
);

// user end points
export const signIn = (formData) => API.post('user/signin', formData);
export const signUp = (formData) => API.post('user/signup', formData);
export const fetchUserProfile = (id, query = {}) => {
    const params = {};
    const { page, limit } = query;
    if (page != null) params.page = page;
    if (limit != null) params.limit = limit;
    return API.get(`user/getUser/${id}`, { params });
};
export const getSugUser = () => API.get("user/getSug");
export const updateUser = (userData) => API.patch("user/update", userData);
export const following = (id) => API.patch(`user/${id}/following`);

// post methods
export const fetchPosts = (id, query = {}) => {
    const params = { id };
    const { page, limit } = query;
    if (page != null) params.page = page;
    if (limit != null) params.limit = limit;
    return API.get('posts', { params });
};
export const createPost = (postData) => API.post('posts', postData);
export const searchPosts = (searchQuery) => API.get('posts/search', { params: { searchQuery } });
export const fetchPost = (id) => API.get(`posts/${id}`);
export const deletePost = (id) => API.delete(`posts/${id}`);
export const updatePost = (id, postData) => API.patch(`posts/${id}`, postData);
export const commentPost = (id, commentData) => API.post(`posts/${id}/commentPost`, commentData);
export const likePost = (id) => API.patch(`posts/${id}/likePost`);


// Notification
export const GetNotificationForUser = (userId) => API.get(`notification/${userId}`, { cacheTTL: 3_000 });
export const MarkNotificationAsReaded = (userId) =>
    API.patch(`notification/mark-notification-as-readed/${userId}`);

// chat
export const fetchUnreadMessageSummary = () => API.get("chat/get-user-unreadmsg", { cacheTTL: 3_000 });
export const fetchConversationMessages = (from, firstuid, seconduid) =>
    API.get("chat/getmsgsbynums", {
        params: {
            from,
            firstuid,
            seconduid,
        },
    });
export const markConversationAsRead = (otheruid) =>
    API.patch("chat/read-msg", null, {
        params: {
            otheruid,
        },
    });
export const sendChatMessage = (message) => API.post("chat/sendmessage", message);
