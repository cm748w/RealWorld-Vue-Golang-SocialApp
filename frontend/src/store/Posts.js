import * as api from '../api/index.js'

/**
 * 帖子管理模块
 * 负责处理社交应用中帖子的相关操作
 * 包括：获取、创建、更新、删除、点赞、评论和搜索等功能
 */
const Posts = {
    state: {isLoading: true, post:[], posts:[], SearchResult:[]},
    getters: {
        /**
         * 获取单个帖子数据
         * @param {Object} state - 状态对象
         * @returns {Function} 返回帖子数据的函数
         */
        GetPost: (state) => () => {
            return {...state.post}
        },
        /**
         * 获取所有帖子数据
         * @param {Object} state - 状态对象
         * @returns {Function} 返回帖子列表数据的函数
         */
        GetAllPosts: (state) => () => {
            return {...state.posts}
        },
        /**
         * 获取搜索结果数据
         * @param {Object} state - 状态对象
         * @returns {Function} 返回搜索结果数据的函数
         */
        GetSearchData: (state) => () => {
            return {...state.SearchResult}
        },
    },
    mutations: {
        /**
         * 更新单个帖子数据
         * @param {Object} state - 状态对象
         * @param {Object} payload - 帖子数据
         */
        Post(state, payload){
            state.post = payload
        },
        /**
         * 更新帖子列表数据
         * @param {Object} state - 状态对象
         * @param {Array} payload - 帖子列表数据
         */
        Posts(state, payload){
            state.posts = payload
        },
        /**
         * 更新搜索结果数据
         * @param {Object} state - 状态对象
         * @param {Array} payload - 搜索结果数据
         */
        Search(state, payload){
            state.SearchResult = payload
        },
    },
    actions: {
        async GetPost(context, id){
            try{
                let {data} = await api.fetchPost(id)
                context.commit('Post', data)
                return data
            }catch(error){
                console.log(error)
                return null
            }
        },
        async GetAllPosts(context, page){
            try{
                const user = JSON.parse(localStorage.getItem('profile'))
                const userId = user?.result?._id

                if (userId){
                    const {data} = await api.fetchPosts(userId, { page })
                    context.commit('Posts', data)
                    return data;
                }
                return null
            }catch(error){
                console.log(error)
                return null
            }
        },
        async SearchPosts(context, page){
            try{
                const user = JSON.parse(localStorage.getItem('profile'))
                const userId = user?.result?._id
                if (userId){
                    const {data} = await api.fetchPosts(userId, { page })
                    context.commit('Posts', data)
                    return data;
                }
                return null
            }catch(error){
                console.log(error)
                return null
            }
        },
        async getPostsUsersBySearch(context, searchData){
            try {
                const {data} = await api.searchPosts(searchData)
                context.commit('Search', data)
                return data
            } catch (error) {
                console.log(error)
                return null
            }
        },
        async createPost(context, post){
            try {
                const postData = {
                    title: post.title,
                    message: post.message
                }

                if (post.selectedFile) {
                    postData.selectedFile = post.selectedFile
                }

                const {data} = await api.createPost(postData)
                context.commit('Post', data)
                return data
            } catch (error) {
                console.log('Create post error:', error)
                console.log('Create post response:', error?.response?.data)
                throw error
            }
        },
        async updatePost(context, Data){
            try {
                const user = JSON.parse(localStorage.getItem('profile'))
                const userId = user?.result?._id

                const PostData = {
                    "title": Data.title,
                    "message": Data.message,
                    "creator": userId,
                    "selectedFile": Data.selectedFile,
                }

                const post = await api.updatePost(Data.id, PostData)
                context.commit('Post', post)
                return post
            } catch (error) {
                console.log(error)
                return null
            }
        },
        async LikePostByUser(context, id){
            try {
                const {data} = await api.likePost(id)
                context.commit('Post', data)
                console.log('LikePost', data)
                return data
            } catch (error) {
                console.log(error)
                return null
            }
        },
        async commentPost(context, form){
            try {
                const { data } = await api.commentPost(form.id, { value: form.value })
                context.commit('Post', data)
                return data
            } catch (error) {
                console.log(error)
                return null
            }
        },
        async deletePost(context, id){
            try {
                await api.deletePost(id)
                return true
            } catch (error) {
                console.log(error)
                return false
            }
        }
    }
}
export default Posts;
