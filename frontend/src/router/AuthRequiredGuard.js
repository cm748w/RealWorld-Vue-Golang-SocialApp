import store from '@/store'

/**
 * 需要登录才能访问的路由守卫
 * 如果用户未登录，重定向到认证页面
 */
export default function AuthRequiredGuard(to, from, next) {
    store.commit('auth/SetData')

    if (store.state.auth.authData) {
        // 用户已登录，允许访问
        next()
    } else {
        // 用户未登录，重定向到认证页
        next('/Auth')
    }
}
