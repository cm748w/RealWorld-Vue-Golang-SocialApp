import * as api from '../api/index.js'

const NotificationStore = {
	state: {
		unReadedNotification: 0
	},
	getters: {
		GetUnReadedNotification: (state) => () => {
			return state.unReadedNotification
		},
	},
	mutations: {
		updateUnReadedNotification(state, payload) {
			state.unReadedNotification = payload
		},
	},
	actions: {
		async GetUnReadedNotifyNum(context, id){
			try {
				let {data} = await api.GetNotificationForUser(id)
				let numofunreadednot = 0
				data.notifications.forEach(el => {
					if (!el.isRead){
						numofunreadednot++
					}
				})

				context.commit('updateUnReadedNotification', numofunreadednot)

				return data.notifications
			} catch (error) {
				console.log(error)
			}
		},
		async MarkNotifyAsReaded(context, id){
			try{

				let {data} = await api.MarkNotificationAsReaded(id)
				context.commit('updateUnReadedNotification', 0)

				return data

			} catch (error) {
				console.log(error)
			}
		},
	},
}



export default NotificationStore

