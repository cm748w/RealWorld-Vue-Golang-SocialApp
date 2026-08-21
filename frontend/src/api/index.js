import axios from "axios";

// 使用相对路径 /api/，由 nginx 反向代理到后端 GolangApiServer:5000，
// 这样通过任意 IP/域名访问（含局域网 IP）都能正确连上后端，不再依赖构建时写死的地址。
const API = axios.create({ baseURL: '/api/' });

API.interceptors.request.use((req) => {
    if (localStorage.getItem('profile')) {
        req.headers.Authorization = `Bearer ${JSON.parse(localStorage.getItem('profile')).token}`;
    }
    return req;
})

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
export const GetNotificationForUser = (userId) => API.get(`notification/${userId}`);
export const MarkNotificationAsReaded = (userId) =>
    API.patch(`notification/mark-notification-as-readed/${userId}`);

// chat
export const fetchUnreadMessageSummary = () => API.get("chat/get-user-unreadmsg");
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
