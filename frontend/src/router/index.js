import { createRouter, createWebHistory } from 'vue-router'
import store from '@/store'
import HomeView from '@/views/HomeView.vue'
import Auth from '@/views/Auth.vue'
import ProfilePage from '@/views/Profile.vue'
import NunAuthGuard from './NunAuthGuard'
import AuthRequiredGuard from './AuthRequiredGuard'
import PostDetails from '@/components/post/PostDetails.vue'
import Search from '../components/search/Search.vue'
import Notification from '../components/Notification/Notification.vue'
import Chat from '../components/Chat/Chat.vue'

// 应用基路径：Vue CLI 会在构建/运行时把 process.env.BASE_URL 替换为
// vue.config.js 里 publicPath 的值（当前为 '/app/'），因此这里不需要硬编码。
// history 使用该 base 后，路由会自动挂在 /app/ 之下，例如：
//   - router.push('/Auth')            -> 浏览器地址栏 /app/Auth
//   - 直接访问 /app/profile/:id       -> 正常解析为 profile 路由
// 组件内所有路由跳转仍写相对 base 的路径（如 '/Auth'），不要手工加 '/app' 前缀。
const routerBase = process.env.BASE_URL
const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView
  },
  {
    path: '/Auth',
    name: 'Auth',
    component: Auth,
    beforeEnter: [NunAuthGuard]
  },
  {
    path: '/PostDetails/:id',
    name: 'PostDetails',
    component: PostDetails,
  },
  {
    path:'/Search',
    name: 'Search',
    component: Search,
  },
  {
    path: '/profile/:id',
    name: 'profile',
    component: ProfilePage,
    beforeEnter: [AuthRequiredGuard]
  },
  {
    path:'/Notification',
    name:'Notification',
    component: Notification
  },
  {
    path:'/Chat',
    name:'chat',
    component: Chat,
  },
]

const router = createRouter({
  history: createWebHistory(routerBase),
  routes
})

/**
 * 全局登录守卫：未登录用户访问任何非 /Auth 页面，一律跳转 /Auth，
 * 并携带 redirect 参数，登录成功后自动跳回原目标页。
 */
router.beforeEach((to) => {
  store.commit('auth/SetData')

  if (to.name !== 'Auth' && !store.state.auth.authData) {
    return { name: 'Auth', query: { redirect: to.fullPath } }
  }

  return true
})

export default router
