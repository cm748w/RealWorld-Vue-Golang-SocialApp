const RealTimeNotify = {
    namespaced: true,
    state: {
        ws: null,
        notifyideslistNumber: 0,
        notifyidData: null,
    },
    getters: {
        Getnotifyideslist: (state) => {
            return state.notifyideslistNumber
        },
    },
    mutations: {
        SET_WS(state, ws) {
            state.ws = ws
        },
        ADD_NOTIFICATION(state, notify) {
            state.notifyideslistNumber = state.notifyideslistNumber + 1
            state.notifyidData = notify
        },
    },
    actions: {
        async connectToNotify(context) {
            if (JSON.parse(localStorage.getItem('profile')) && context.state.ws == null) {
                const profile = JSON.parse(localStorage.getItem('profile'))
                const Userid = profile.result._id
                // 运行时根据当前页面地址动态拼接，nginx 将 /ws-notify/ 反代到通知服务 8088
                const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
                // 携带 token 鉴权（服务端校验 iss 与路径 userId 一致）
                const token = encodeURIComponent(profile.token || '')
                const ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws-notify/${Userid}?token=${token}`)

                ws.onopen = () => {
                    context.commit('SET_WS', ws)
                }

                ws.onmessage = (event) => {
                    const Notify = JSON.parse(event.data)
                    context.commit('ADD_NOTIFICATION', Notify)
                }
            }
        },

        async StopConnectionToNotify(context) {
            try {
                if (context.state.ws) {  // 添加这行
                    context.state.ws.close()
                }
                context.commit('SET_WS', null)
            } catch (error) {
                console.log('error', error)
            }
        },
    },
}

export default RealTimeNotify;