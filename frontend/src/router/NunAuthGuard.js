import store from '@/store'

/**
 * 未认证用户守卫
 * 防止已登录用户访问认证页面（登录/注册）
 * 如果用户已登录，重定向到首页
 */
export default function NunAuthGuard(to, from, next) {
    store.commit('auth/SetData')

    if (store.state.auth.authData) {
        // 用户已登录，不能访问认证页，重定向到首页
        next('/')
    } else {
        // 用户未登录，允许访问认证页
        next()
    }
}
