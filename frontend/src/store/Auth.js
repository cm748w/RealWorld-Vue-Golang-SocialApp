// 导入 API 模块，用于调用后端接口
import * as api from '../api/index.js';
// 导入 JWT 解码库，用于验证 token 过期时间
import { jwtDecode } from 'jwt-decode';
import { emitProfileSync } from './profileSync.js';

/**
 * 应用登出操作
 * @param {Object} state - Vuex 状态
 * @description 清除本地存储并重置认证数据
 */
function applyLogout(state) {
  localStorage.clear();
  state.authData = null;
}

/**
 * 检查 token 是否过期
 * @param {Object} decodedToken - 解码后的 token 对象
 * @returns {boolean} - token 是否已过期
 * @description 检查 token 中的过期时间是否小于当前时间
 */
function isTokenExpired(decodedToken) {
  const now = Date.now();
  // 获取 token 中的过期时间（秒）并转换为毫秒
  const expMs =
    decodedToken.exp != null ? decodedToken.exp * 1000 : null;
  // 获取 token 中的另一种过期时间格式
  const expiresMs =
    decodedToken.expires != null ? decodedToken.expires * 1000 : null;
  // 检查是否有任一过期时间已过期
  return (
    (expMs != null && expMs < now) ||
    (expiresMs != null && expiresMs < now)
  );
}

// 认证模块
const Auth = {
  // 启用命名空间，避免模块间的命名冲突
  namespaced: true,
  state: {
    // 存储认证数据
    authData: null,
  },
  getters: {
    /**
     * 获取用户数据
     * @param {Object} state - Vuex 状态
     * @returns {Object|null} - 用户认证数据
     */
    GetUserData: (state) => state.authData,
  },
  mutations: {
    /**
     * 设置认证数据
     * @param {Object} state - Vuex 状态
     * @param {Object} payload - 认证数据
     * @description 将认证数据存储到本地存储和状态中
     */
    Auth(state, payload) {
      // 将认证数据存储到本地存储
      localStorage.setItem('profile', JSON.stringify({ ...payload }));
      // 更新状态中的认证数据
      state.authData = payload;
    },
    /**
     * 从本地存储加载认证数据
     * @param {Object} state - Vuex 状态
     * @description 从本地存储中获取认证数据并验证 token 是否过期
     */
    SetData(state) {
      let user = null;
      try {
        // 从本地存储获取认证数据
        const raw = localStorage.getItem('profile');
        if (raw) {
          user = JSON.parse(raw);
        }
      } catch {
        // 解析失败时执行登出操作
        applyLogout(state);
        return;
      }

      // 获取 token
      const token = user?.token;
      if (token) {
        try {
          // 解码 token
          const decodedToken = jwtDecode(token);
          // 检查 token 是否过期
          if (isTokenExpired(decodedToken)) {
            // token 过期时执行登出操作
            applyLogout(state);
            return;
          }
        } catch {
          // 解码失败时执行登出操作
          applyLogout(state);
          return;
        }
      }

      // 更新状态中的认证数据
      state.authData = user;
    },
    /**
     * 执行登出操作
     * @param {Object} state - Vuex 状态
     * @description 调用 applyLogout 函数执行登出
     */
    Logout(state) {
      applyLogout(state);
    },
  },
  actions: {
    /**
     * 用户登录
     * @param {Object} context - Vuex 上下文
     * @param {Object} formData - 登录表单数据
     * @returns {Promise<Object>} - 登录响应数据
     * @description 调用登录 API 并处理响应
     */
    async signin({ commit }, formData) {
      try {
        // 调用登录 API
        const { data } = await api.signIn(formData);
        // 提交认证数据到状态
        commit('Auth', data);
        emitProfileSync({
          type: 'auth-updated',
          payload: {
            authData: data,
          },
        });
        // 从本地存储加载认证数据
        commit('SetData');
        // 返回登录响应数据
        return data;
      } catch (error) {
        // 打印错误信息
        console.log(error);
        // 返回错误对象
        return error;
      }
    },
    /**
     * 用户注册
     * @param {Object} context - Vuex 上下文
     * @param {Object} formData - 注册表单数据
     * @returns {Promise<Object>} - 注册响应数据
     * @description 调用注册 API 并处理响应
     */
    async signup({ commit }, formData) {
      try {
        // 调用注册 API
        const { data } = await api.signUp(formData);
        // 提交认证数据到状态
        commit('Auth', data);
        emitProfileSync({
          type: 'auth-updated',
          payload: {
            authData: data,
          },
        });
        // 从本地存储加载认证数据
        commit('SetData');
        // 返回注册响应数据
        return data;
      } catch (error) {
        // 打印错误信息
        console.log(error);
        // 返回错误对象
        return error;
      }
    },
    /**
     * 用户登出
     * @param {Object} context - Vuex 上下文
     * @description 调用 Logout mutation 执行登出
     */
    logout({ commit }) {
      commit('Logout');
      emitProfileSync({
        type: 'logout',
      });
    },
  },
};

export default Auth;
