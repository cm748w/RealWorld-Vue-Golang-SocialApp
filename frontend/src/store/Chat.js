import * as api from '../api/index.js'

const Chat = {
    state:{
        unReadedMsgsNUM: 0
    },
    getters:{
        getUnReadedMsg: (state)=> ()=>{
            return state.unReadedMsgsNUM
        }
    },
    mutations:{
        updateUnreadedMsg(state, payload){
            state.unReadedMsgsNUM = payload
        }
    },
    actions:{
        async GetUnreadedMessageNum(context){
            try {
                let {data} = await api.fetchUnreadMessageSummary()
                context.commit('updateUnreadedMsg', data.totalUnreadMessageCount)
                return data
            } catch (error) {
                console.log(error)
                context.commit('updateUnreadedMsg', 0)
                return { messages: [], totalUnreadMessageCount: 0 }
            }
        },
        async GetChatMsgsBetweenTwoUsers(context, ndata){
            try {
                let {data} = await api.fetchConversationMessages(ndata.from, ndata.firstuid, ndata.seconduid)
                return data
            } catch (error) {
                console.log(error)
            }
        },
        async SendMessage(context, sdata){
            try {
                const msg = 
                {
                    "content": sdata.content,
                    "sender": sdata.sender,
                    "receiver": sdata.receiver,
                }
                let {data} = await api.sendChatMessage(msg)
                // Backend returns { message, result }, where result is the saved message.
                console.log(data)
                return data.result
            } catch (error) {
                console.log(error)
                return null
            }
        },
        async MarkMsgsAsReaded(context, datau){
            try {
                let {data} = await api.markConversationAsRead(datau.otheruid)
                var olunreaded = context.state.unReadedMsgsNUM
                var unreaded = datau.GetunReadedmessage

                var finalnum = olunreaded - unreaded
                context.commit('updateUnreadedMsg', finalnum)

                // console.log()
                return data

            } catch (error) {
                console.log(error)
            }
        }
    },
}

export default Chat