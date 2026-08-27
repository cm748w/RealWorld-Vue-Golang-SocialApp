const RealTimeChat = {
    state: {
        ws: null,
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
        SET_WS(state, ws) {
            state.ws = ws
        },
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
            if (JSON.parse(localStorage.getItem('profile'))) {
                state.userId = JSON.parse(localStorage.getItem('profile'))?.result?._id
            }
        }
    },
    actions: {
        async createChatConnection(context) {
            try {
                context.commit('setUserId')
                if (context.state.userId && context.state.ws == null) {
                    // 运行时根据当前页面地址动态拼接，nginx 将 /ws-chat/ 反代到聊天服务 8001
                    const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
                    // 携带 token 鉴权（服务端校验 iss 与路径 userId 一致）
                    const token = encodeURIComponent(JSON.parse(localStorage.getItem('profile'))?.token || '')
                    const ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws-chat/${context.state.userId}?token=${token}`)

                    ws.onopen = () => {
                        context.commit('SET_WS', ws)
                    }

                    ws.onmessage = (event) => {
                        const message = JSON.parse(event.data)
                        if (!message.onlineFriends) {
                            context.commit('UpdateNumberOfMessages')
                            context.commit('AddPrivateMessage', message)
                        } else {
                            const uniqueUsers = Array.from(new Set(message.onlineFriends))
                            context.commit('setOnlineUsers', uniqueUsers)
                        }
                    }
                }
            } catch (error) {
                console.log('E', error)
            }
        },
        async SendPrivateMessage(context, message){
            if(context.state.ws){
                return context.state.ws.send(JSON.stringify(message))
            }
        },
        async StopConnectionToChat(context) {
            try {
                if (context.state.ws) {  // 添加这行
                    context.state.ws.close()
                }
                context.commit('SET_WS', null)
            } catch (error) {
                console.log('error', error)
            }
        },

    }


}


export default RealTimeChat