import * as api from '../api/index.js'
import { logWarn } from '@/utils/log.js'

// 失败返回契约（本模块统一）：
//  - 调用方会解构对象的 action（如 MarkMsgsAsReaded）→ 失败时返回同形状的安全默认值；
//  - 调用方会判空/可选链兜底的 action → 失败时返回 null 或空列表，绝不返回 undefined。
const Chat = {
    state:{
        unReadedMsgsNUM: 0
    },
    getters:{
        getUnReadedMsg: (state)=> {
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
                logWarn('Chat.GetUnreadedMessageNum', error)
                context.commit('updateUnreadedMsg', 0)
                return { messages: [], totalUnreadMessageCount: 0 }
            }
        },
        async GetChatMsgsBetweenTwoUsers(context, ndata){
            try {
                let {data} = await api.fetchConversationMessages(ndata.from, ndata.firstuid, ndata.seconduid)
                return data
            } catch (error) {
                logWarn('Chat.GetChatMsgsBetweenTwoUsers', error)
                // 调用方按 result.msgs 取值，失败时给同形状的空结果
                return { msgs: [] }
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
                return data.result
            } catch (error) {
                logWarn('Chat.SendMessage', error)
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

                return data

            } catch (error) {
                // 修复：旧实现在 catch 里只打日志不返回，导致该 action 返回 undefined，
                // 而 Chat.vue 会直接解构 `const { isMarked } = await ...` → TypeError 崩溃。
                // 这里保持与成功分支一致的返回契约，失败时明确告知调用方“未标记成功”。
                logWarn('Chat.MarkMsgsAsReaded', error)
                return { isMarked: false }
            }
        }
    },
}

export default Chat
