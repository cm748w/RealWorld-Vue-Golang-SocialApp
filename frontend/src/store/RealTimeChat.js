import { createReconnectingWebSocket } from '@/utils/wsClient.js'
import { logWarn } from '@/utils/log.js'

// WebSocket 实例保存在模块作用域，而不是 Vuex state：
// state 会被 Vue 用 reactive() 代理，代理对象上调用 WebSocket 的 close()/send()
// 会因接收者不是原始 WebSocket 而抛 "Illegal invocation"。
let chatSocket = null

/** 安全读取本地登录态（localStorage.profile），解析失败不抛异常 */
function readProfile() {
    try {
        return JSON.parse(localStorage.getItem('profile') || 'null') || null
    } catch (error) {
        logWarn('RealTimeChat', '解析本地 profile 失败', error)
        return null
    }
}

const RealTimeChat = {
    state: {
        privateMessages: [],
        onlineFriends: [],
        userId: '',
        NumberOfMessagesReal: 0
    },
    getters: {
        Getuserid: (state) => {
            return state.userId
        },
        GetPrivateMessages: (state) => {
            return state.privateMessages
        },
        GetRealTimeNumberMessages: (state) => {
            return state.NumberOfMessagesReal
        },
        GetOnlinefriends: (state) => {
            return state.onlineFriends
        },
    },
    mutations: {
        UpdateNumberOfMessages(state) {
            state.NumberOfMessagesReal = state.NumberOfMessagesReal + 1
        },
        setOnlineUsers(state, onlineFriends) {
            state.onlineFriends = onlineFriends
        },
        AddPrivateMessage(state, message) {
            state.privateMessages = message;
        },
        clearPrivateMessage(state) {
            state.privateMessages = []
        },
        setUserId(state) {
            const profile = readProfile()
            state.userId = profile?.result?._id || ''
        }
    },
    actions: {
        async createChatConnection(context) {
            context.commit('setUserId')

            const profile = readProfile()
            const userId = profile?.result?._id || context.state.userId
            if (!userId || !profile?.token) {
                // 未登录：不建连（也不做重连）
                return false
            }

            if (chatSocket) {
                // 已连接或正在重连：不重复建连
                // （旧实现用 state.ws == null 判断，而 state.ws 要等 onopen 才赋值，
                //   连续调用会建出多条连接，这里用模块作用域引用彻底避免）
                chatSocket.start()
                return true
            }

            chatSocket = createReconnectingWebSocket({
                scope: 'RealTimeChat',
                // 聊天通道用 watchdog：后端对客户端发来的任何帧都会当作聊天消息
                // （backend/realTimeChat/main.go:78-89 读到 {"sender","receiver","content"} 后
                //  交给 SendToReceiver），既不回 pong，还会在服务端打出
                // "Receiver  not found" 日志，因此这里不发应用层心跳，
                // 只做 readyState 巡检 + onclose/onerror 退避重连。
                heartbeat: 'watchdog',
                url: () => {
                    const current = readProfile()
                    const token = encodeURIComponent(current?.token || '')
                    const uid = current?.result?._id || context.state.userId
                    if (!token || !uid) {
                        // 已登出 / token 丢失：停止重连
                        return null
                    }
                    // 运行时根据当前页面地址动态拼接，nginx 将 /ws-chat/ 反代到聊天服务 8001
                    const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
                    // TODO(security): token 仍在 URL query 里（会进 nginx access log 与 Referer）。
                    // 正确做法是改用 Sec-WebSocket-Protocol 子协议携带 token，
                    // 但这需要后端同步支持（backend/realTimeChat/realtime/jwtverify.go 目前只读 query），
                    // 属跨端改动，本次前端不单独变更协议。
                    return `${wsProtocol}://${window.location.host}/ws-chat/${uid}?token=${token}`
                },
                onMessage: (message) => {
                    if (!message) return
                    if (message.onlineFriends) {
                        const uniqueUsers = Array.from(new Set(message.onlineFriends))
                        context.commit('setOnlineUsers', uniqueUsers)
                    } else {
                        context.commit('UpdateNumberOfMessages')
                        context.commit('AddPrivateMessage', message)
                    }
                },
            })

            chatSocket.start()
            return true
        },
        async SendPrivateMessage(context, message) {
            // 未连接时返回 false，调用方（Chat.vue）据此回退到 REST 持久化路径，
            // 避免实时通道不可用时消息被静默丢弃。
            if (!chatSocket) return false
            return chatSocket.send(message)
        },
        async StopConnectionToChat() {
            if (chatSocket) {
                // stop() 会同时关闭连接、清掉重连定时器与心跳定时器
                chatSocket.stop()
                chatSocket = null
            }
        },
    }
}

export default RealTimeChat
