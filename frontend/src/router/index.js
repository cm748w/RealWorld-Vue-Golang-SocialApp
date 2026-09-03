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
  history: createWebHistory(process.env.BASE_URL),
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
