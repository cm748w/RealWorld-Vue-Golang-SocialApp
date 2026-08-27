// 导入 API 模块，用于调用后端接口
import * as api from '@/api/index.js'
import { emitProfileSync } from './profileSync.js'

// 关注列表缓存：避免每次进入 Chat 都把 followers+following 逐个重查（N+1）
const FOLLOWERS_TTL = 30_000

function normalizeUserResponse(data) {
    const payload = data?.user || data?.result || data?.data || data || {}
    const user = payload?.user || payload?.result || payload
    const posts = data?.posts || payload?.posts || []

    return {
        user,
        posts,
        raw: data,
    }
}

/**
 * 用户模块
 * 管理用户相关的状态和操作
 */
const Users = {
    // 启用命名空间，避免模块间的命名冲突
    namespaced: true,
    state: {
        // 按 id 缓存用户数据，供多个界面共享
        usersById: {},
        // 推荐用户列表
        recommendedUsers: [],
        // 用户的 followers 和 following 列表
        userFollowersFollowing: [],
        // 关注列表最后一次拉取时间（用于短 TTL 缓存）
        followersFetchedAt: 0
    },
    getters: {
        /**
         * 获取当前用户数据
         * @param {Object} state - Vuex 状态
         * @returns {Function} - 返回获取用户数据的函数
         */
        GetUser: (state) => (id) => {
            return id ? state.usersById[id] : null
        },
        /**
         * 获取推荐用户列表
         * @param {Object} state - Vuex 状态
         * @returns {Array} - 推荐用户列表
         */
        GetRecommendedUsers: (state) => {
            return state.recommendedUsers
        },
        /**
         * 获取用户的 followers 和 following 列表
         * @param {Object} state - Vuex 状态
         * @returns {Array} - followers 和 following 合并的用户列表
         */
        GetUserFollowersFollowing: (state) => {
            return state.userFollowersFollowing
        }
    },
    mutations: {
        /**
         * 设置用户数据
         * @param {Object} state - Vuex 状态
         * @param {Object} payload - 用户数据
         */
        SetUser(state, payload) {
            if (!payload?._id) {
                return
            }

            state.usersById = {
                ...state.usersById,
                [payload._id]: payload,
            }
        },
        /**
         * 设置推荐用户列表
         * @param {Object} state - Vuex 状态
         * @param {Array} payload - 推荐用户列表
         */
        SetRecommendedUsers(state, payload) {
            state.recommendedUsers = payload
        },
        /**
         * 更新用户关注状态
         * @param {Object} state - Vuex 状态
         * @param {Array} following - 关注列表
         */
        UpdateUserFollowing(state, payload) {
            const userId = payload?.id
            if (userId) {
                state.usersById = {
                    ...state.usersById,
                    [userId]: {
                        ...state.usersById[userId],
                        following: payload.following,
                    },
                }
            }
        },
        /**
         * 设置用户的 followers 和 following 列表
         * @param {Object} state - Vuex 状态
         * @param {Array} payload - 合并后的用户列表
         */
        SetUserFollowersFollowing(state, payload) {
            state.userFollowersFollowing = payload || []
        },
        SetFollowersFetchedAt(state, ts) {
            state.followersFetchedAt = ts
        }
    },
    actions: {
        /**
         * 根据 ID 获取用户信息
         * @param {Object} context - Vuex 上下文
         * @param {string} id - 用户 ID
         * @returns {Promise<Object>} - 用户信息
         */
        async GetUserById({ commit }, id) {
            try {
                const { data } = await api.fetchUserProfile(id)
                const normalized = normalizeUserResponse(data)
                commit('SetUser', normalized.user)
                return normalized
            } catch (error) {
                console.log(error)
                return error
            }
        },
        /**
         * 更新用户数据
         * @param {Object} context - Vuex 上下文
         * @param {Object} userData - 用户数据
         * @returns {Promise<Object>} - 更新后的用户信息
         */
        async UpdateUserData({ commit, rootState }, userData) {
            try {
                const { data } = await api.updateUser(userData)
                const normalized = normalizeUserResponse(data)
                commit('SetUser', normalized.user)
                
                // 如果更新的是当前登录用户，同时更新 Auth store 的数据
                const currentUserId = rootState.auth.authData?.result?._id
                const updatedUserId = normalized.user?._id
                if (String(currentUserId) === String(updatedUserId)) {
                    // 更新 Auth store 中的用户数据
                    const authData = {
                        ...rootState.auth.authData,
                        result: normalized.user
                    }
                    commit('auth/Auth', authData, { root: true })

                    emitProfileSync({
                        type: 'profile-updated',
                        payload: {
                            users: [normalized.user],
                            authData,
                        },
                    })
                } else {
                    emitProfileSync({
                        type: 'profile-updated',
                        payload: {
                            users: [normalized.user],
                        },
                    })
                }
                
                return normalized
            } catch (error) {
                console.log(error)
                return error
            }
        },
        /**
         * 关注用户
         * @param {Object} context - Vuex 上下文
         * @param {string} ProfileId - 要关注的用户 ID
         * @returns {Promise<Object>} - 关注操作的结果
         */
        async FollowUser({ commit, rootState }, ProfileId) {
            try {
                await api.following(ProfileId)
                // 关注状态变了，使关注列表缓存失效，下次进入 Chat 重新拉取
                commit('SetFollowersFetchedAt', 0)

                const refreshedProfile = await api.fetchUserProfile(ProfileId)
                const normalizedProfile = normalizeUserResponse(refreshedProfile.data)
                commit('SetUser', normalizedProfile.user)

                const currentUserId = rootState.auth.authData?.result?._id
                let normalizedCurrentUser = null
                if (currentUserId && String(currentUserId) !== String(ProfileId)) {
                    const refreshedCurrentUser = await api.fetchUserProfile(currentUserId)
                    normalizedCurrentUser = normalizeUserResponse(refreshedCurrentUser.data)
                    commit('SetUser', normalizedCurrentUser.user)

                    if (rootState.auth.authData) {
                        const authData = {
                            ...rootState.auth.authData,
                            result: normalizedCurrentUser.user
                        }
                        commit('auth/Auth', authData, { root: true })

                        emitProfileSync({
                            type: 'profile-updated',
                            payload: {
                                users: [normalizedProfile.user, normalizedCurrentUser.user],
                                authData,
                            },
                        })
                    }
                } else {
                    emitProfileSync({
                        type: 'profile-updated',
                        payload: {
                            users: [normalizedProfile.user],
                        },
                    })
                }

                return normalizedProfile
            } catch (error) {
                console.log(error)
                return error
            }
        },
        /**
         * 获取推荐用户
         * @param {Object} context - Vuex 上下文
         * @returns {Promise<Array>} - 推荐用户列表
         */
        async GetRecommendUsers({ commit }) {
            try {
                const { data } = await api.getSugUser()
                commit('SetRecommendedUsers', data)
                return data
            } catch (error) {
                console.log(error)
                return error
            }
        },
        /**
         * 获取用户的 followers 和 following 列表
         * @param {Object} context - Vuex 上下文
         * @returns {Promise<Array>} - followers 和 following 合并后的用户列表
         */
        async FetchUserFollowersFollowing({ commit, rootState, state }) {
            try {
                const localProfile = JSON.parse(localStorage.getItem('profile') || 'null')
                const currentUserId = rootState?.auth?.authData?.result?._id || localProfile?.result?._id
                if (!currentUserId) {
                    console.error('User profile not found in localStorage')
                    return []
                }

                // 短 TTL 缓存：30s 内再次进入 Chat 直接复用，避免 N+1 逐个重查
                if (state.userFollowersFollowing.length && Date.now() - (state.followersFetchedAt || 0) < FOLLOWERS_TTL) {
                    return state.userFollowersFollowing
                }

                // Always refresh current user from backend so followers/following are up-to-date.
                const { data: meData } = await api.fetchUserProfile(currentUserId)
                const meNormalized = normalizeUserResponse(meData)
                const latestUser = meNormalized?.user || {}

                // Keep auth/profile cache in sync for other pages that still read localStorage profile.
                const baseAuthData = rootState?.auth?.authData || localProfile || {}
                if (baseAuthData?.token) {
                    const refreshedAuthData = {
                        ...baseAuthData,
                        result: latestUser,
                    }
                    commit('auth/Auth', refreshedAuthData, { root: true })
                }

                var followers = latestUser.followers || []
                var following = latestUser.following || []

                const combineArray = [...followers, ...following]
                const uniqueArray = Array.from(new Set(combineArray))

                var userdata = []
                for(const uid of uniqueArray){
                    try {
                        const { data } = await api.fetchUserProfile(uid)
                        var user = {
                            "_id": data.user._id,
                            "name": data.user.name,
                            "imageUrl": data.user.imageUrl
                        }
                        userdata.push(user)
                    } catch (error) {
                        console.error(`Failed to fetch profile for user ${uid}:`, error)
                    }
                }
                
                commit('SetUserFollowersFollowing', userdata)
                commit('SetFollowersFetchedAt', Date.now())
                return userdata
            } catch (error) {
                console.error('Error fetching followers/following:', error)
                return []
            }
        }
    }
}

export default Users