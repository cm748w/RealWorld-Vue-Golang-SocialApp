import { createStore } from 'vuex'
import auth from './Auth.js'
import users from './Users.js'
import posts from './Posts.js'
import { installProfileSync } from './profileSync.js'
import NotificationStore from './Notification.js'
import Chat from './Chat.js'
const store = createStore({
  modules: {
    auth,
    users,
    posts,
    NotificationStore,
    Chat,
  },
})

installProfileSync(store)

export default store
