import { createReconnectingWebSocket } from '@/utils/wsClient.js'
import { logWarn } from '@/utils/log.js'

// WebSocket 实例保存在模块作用域（原因同 RealTimeChat.js）：
// 放进 Vuex state 会被 reactive() 代理，close()/send() 会抛 "Illegal invocation"。
let notifySocket = null

/** 安全读取本地登录态（localStorage.profile），解析失败不抛异常 */
function readProfile() {
    try {
        return JSON.parse(localStorage.getItem('profile') || 'null') || null
    } catch (error) {
        logWarn('RealTimeNotify', '解析本地 profile 失败', error)
        return null
    }
}

/**
 * 判断一帧是否只是心跳回显，而不是真正的通知。
 * 后端通知通道的读循环会把客户端发来的帧原样回显
 * （backend/realTimeNotification/realtime/notify-realtime-ws.go:96-104），
 * 心跳帧回显回来时所有业务字段都为空；如果不滤掉，未读计数和通知列表会被污染。
 */
function isHeartbeatEcho(payload) {
    return !payload || (!payload._id && !payload.details && !payload.targetId)
}

const RealTimeNotify = {
    namespaced: true,
    state: {
        notifyideslistNumber: 0,
        notifyidData: null,
    },
    getters: {
        Getnotifyideslist: (state) => {
            return state.notifyideslistNumber
        },
    },
    mutations: {
        ADD_NOTIFICATION(state, notify) {
            state.notifyideslistNumber = state.notifyideslistNumber + 1
            state.notifyidData = notify
        },
    },
    actions: {
        async connectToNotify(context) {
            const profile = readProfile()
            const userId = profile?.result?._id
            if (!userId || !profile?.token) {
                // 未登录：不建连（也不做重连）
                return false
            }

            if (notifySocket) {
                // 已连接或正在重连：不重复建连
                notifySocket.start()
                return true
            }

            notifySocket = createReconnectingWebSocket({
                scope: 'RealTimeNotify',
                // 通知通道用 ping：后端读循环会把客户端帧回显，等价于 pong，
                // 因此可以探测到“TCP 已死但 readyState 仍是 OPEN”的黑洞连接。
                heartbeat: 'ping',
                url: () => {
                    const current = readProfile()
                    const token = encodeURIComponent(current?.token || '')
                    const uid = current?.result?._id
                    if (!token || !uid) {
                        // 已登出 / token 丢失：停止重连
                        return null
                    }
                    // 运行时根据当前页面地址动态拼接，nginx 将 /ws-notify/ 反代到通知服务 8088
                    const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
                    // TODO(security): token 仍在 URL query 里（会进 nginx access log 与 Referer）。
                    // 正确做法是改用 Sec-WebSocket-Protocol 子协议携带 token，
                    // 但这需要后端同步支持（backend/realTimeNotification/realtime/jwtverify.go
                    // 目前只读 query），属跨端改动，本次前端不单独变更协议。
                    return `${wsProtocol}://${window.location.host}/ws-notify/${uid}?token=${token}`
                },
                onMessage: (payload) => {
                    if (isHeartbeatEcho(payload)) return
                    context.commit('ADD_NOTIFICATION', payload)
                },
            })

            notifySocket.start()
            return true
        },

        async StopConnectionToNotify() {
            if (notifySocket) {
                // stop() 会同时关闭连接、清掉重连定时器与心跳定时器
                notifySocket.stop()
                notifySocket = null
            }
        },
    },
}

export default RealTimeNotify;
