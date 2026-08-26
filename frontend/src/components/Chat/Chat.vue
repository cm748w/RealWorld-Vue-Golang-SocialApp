<template>
    <q-page class="constrain q-pa-md">
        <div class="row">
            <div class="col-12">
                <div class="chat-shell surface">

                    <!-- Contact list -->
                    <div class="chat-contact" v-show="!isMobile || !selectedUser">
                        <q-toolbar class="bg-grey-9 text-white shadow-1">
                            <q-toolbar-title class="text-subtitle1">Messages</q-toolbar-title>
                        </q-toolbar>

                        <q-list separator class="q-pa-sm">
                            <q-item @click="selectUser(contact)" v-for="contact in contacts" :key="contact._id"
                                class="q-my-xs" clickable v-ripple>
                                <q-item-section avatar>
                                    <q-avatar v-if="!contact.imageUrl">
                                        <img :src="defaultAvatar" />
                                    </q-avatar>
                                    <q-avatar v-else>
                                        <img :src="contact?.imageUrl" />
                                    </q-avatar>
                                </q-item-section>
                                <q-item-section>
                                    <q-item-label>{{ contact.name }}</q-item-label>
                                </q-item-section>

                                <q-item-section side v-if="contact.isOnline">
                                    <q-badge color="positive" rounded />
                                </q-item-section>

                                <q-item-section side v-if="contact.unReadedmessage && contact.unReadedmessage > 0">
                                    <q-badge color="negative" rounded :label="contact?.unReadedmessage" />
                                </q-item-section>
                            </q-item>
                        </q-list>
                    </div>

                    <!-- Chat pane -->
                    <div class="chat-pane" v-if="selectedUser != null">
                        <div class="q-pa-sm row items-center q-gutter-sm chat-pane-header">
                            <q-btn v-if="isMobile" flat round icon="eva-arrow-back" color="grey-8" @click="selectedUser = null" />
                            <q-avatar>
                                <img :src="selectedUser.imageUrl || defaultAvatar" />
                            </q-avatar>
                            <div>
                                <div class="text-weight-bold text-subtitle2">{{ selectedUser.name }}</div>
                                <div class="muted text-caption">{{ selectedUser.isOnline ? 'Online' : 'Offline' }}</div>
                            </div>
                        </div>

                        <q-separator />

                        <div class="message-container" ref="messageContainer" @scroll="handleScroll">
                            <div v-for="msg in messageBetweenUsers" :key="msg._id" style="width: 100%;">
                                <q-chat-message
                                    :name="msg.sender === MainUserData._id ? MainUserData.name : selectedUser.name"
                                    :avatar="msg.sender === MainUserData._id ? (MainUserData.imageUrl || defaultAvatar) : (selectedUser.imageUrl || defaultAvatar)"
                                    :text="[msg.content]" :sent="msg.sender === MainUserData._id ? true : false" />
                            </div>
                        </div>

                        <q-separator spaced />
                        <q-input outlined v-model="messageToSend.text" @keyup.enter="handleSendMessage"
                            label="write message.." class="q-pa-sm">
                            <q-btn v-if="messageToSend.text != ''" @click="handleSendMessage" flat round color="accent"
                                icon="eva-arrow-right" />
                        </q-input>
                    </div>

                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { mapActions, mapGetters, mapState } from 'vuex';

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
    name: 'ChatComponent',
    data() {
        return {
            messageToSend: { text: '' },
            contacts: [],
            messageBetweenUsers: [],
            messagelistnum: 0,
            selectedUser: null,
            MainUserData: {},
            uniqueOnlineUsers: [],
            defaultAvatar: DEFAULT_AVATAR,
        }
    },
    computed: {
        ...mapGetters('users', ['GetUserFollowersFollowing', 'GetUser']),
        ...mapGetters('auth', ['GetUserData']),
        ...mapState(["RealTimeChat"]),
        isMobile() {
            return this.$q.screen.lt.md
        }
    },
    watch: {
        "RealTimeChat.onlineFriends": function (online) {
            const onlineFriendsArray = Object.values(online || {})
            console.log('Online friend changed new val', onlineFriendsArray)
            this.uniqueOnlineUsers = Array.from(new Set(onlineFriendsArray.map(u => String(u))))
            this.updateOnlineList()
        },
        "RealTimeChat.privateMessages": function (message) {
            if (this.contacts.length > 0) {
                if (this.selectedUser && this.selectedUser?._id == message.sender) {
                    this.messageBetweenUsers.push(message)
                    let unreadSnapshot = 0
                    this.contacts.forEach((contact) => {
                        if (String(contact._id) === String(message.sender)) {
                            unreadSnapshot = Number(contact.unReadedmessage) || 0
                            contact.unReadedmessage = 0
                        }
                    })
                    this.CallMarkMsgAsReaded(this.selectedUser, unreadSnapshot)
                    setTimeout(() => {
                        this.scrollDownFunction()
                    }, 100)
                } else {
                    this.contacts.forEach((contact) => {
                        if (contact._id == message.sender) {
                            contact.unReadedmessage++
                        }
                    })
                    this.syncGlobalUnreadCount()
                }

            }
        }
    },
    async mounted() {
        this.MainUserData = this.GetUserData?.result || {}
        await this.GetUsList()

        const online = (this.RealTimeChat && this.RealTimeChat.onlineFriends) || {}
        this.uniqueOnlineUsers = Array.from(new Set(Object.values(online).map(u => String(u))))
        this.updateOnlineList()
    },
    methods: {
        ...mapActions({
            GetUnreadedMessageNum: 'GetUnreadedMessageNum',
            GetChatMsgsBetweenTwoUsers: 'GetChatMsgsBetweenTwoUsers',
            sendMessageAction: 'SendMessage',
            MarkMsgsAsReaded: 'MarkMsgsAsReaded',
            FetchUserFollowersFollowing: ['users', 'FetchUserFollowersFollowing'],
        }),
        ...mapActions(['SendPrivateMessage']),
        updateOnlineList() {
            if (!Array.isArray(this.contacts)) return
            const onlineSet = new Set((this.uniqueOnlineUsers || []).map(u => String(u)))
            this.contacts.forEach(contact => {
                const isOnline = onlineSet.has(String(contact._id))
                if (contact && Object.prototype.hasOwnProperty.call(contact, 'isOnline')) {
                    contact.isOnline = isOnline
                } else if (this.$set) {
                    this.$set(contact, 'isOnline', isOnline)
                } else if (contact) {
                    contact.isOnline = isOnline
                }
            })
        },
        handleScroll() {
            const container = this.$refs.messageContainer
            if (!container) return
            if (container.scrollTop === 0) {
                this.GetOldestMessagesBetweenUsers()
            }
        },
        async GetOldestMessagesBetweenUsers() {
            this.messagelistnum = this.messagelistnum + 1
            var firstuid = this.MainUserData._id
            var seconduid = this.selectedUser?._id
            if (!firstuid || !seconduid) return
            var from = this.messagelistnum
            var ndata = { from, firstuid, seconduid }

            const result = await this.GetChatMsgsBetweenTwoUsers(ndata)
            const msgs = result?.msgs || []
            if (Array.isArray(msgs)) {
                this.messageBetweenUsers.unshift(...msgs)
            }
        },
        scrollDownFunction() {
            const container = this.$refs.messageContainer
            if (!container) return
            container.scrollTop = container.scrollHeight
        },
        async CallMarkMsgAsReaded(user, unreadSnapshot = 0) {
            var mainuid = this.MainUserData._id
            var otheruid = user._id

            const data = {
                mainuid,
                otheruid,
                GetunReadedmessage: unreadSnapshot,
            }
            var { isMarked } = await this.MarkMsgsAsReaded(data)

            if (isMarked) {
                this.contacts.forEach(user => {
                    if (String(otheruid) == String(user._id)) {
                        user.unReadedmessage = 0
                    }
                })
                const unreadPayload = await this.GetUnreadedMessageNum(this.MainUserData._id)
                const totalUnreadMessageCount = unreadPayload?.totalUnreadMessageCount
                if (typeof totalUnreadMessageCount === 'number') {
                    this.$store.commit('updateUnreadedMsg', totalUnreadMessageCount)
                }
            }
        },
        async GetUnreadedMsgList() {
            if (!this.MainUserData._id) return
            const unreadPayload = await this.GetUnreadedMessageNum(this.MainUserData._id)
            var messages = unreadPayload?.messages || []

            this.contacts.forEach(user => {
                user.unReadedmessage = 0
            })

            messages.forEach(msg => {
                this.contacts.forEach(user => {
                    if (String(msg.otherUserId) == String(user._id)) {
                        user.unReadedmessage = Number(msg.numOfUnreadMessages)
                    }
                })
            })

            this.syncGlobalUnreadCount()
        },
        syncGlobalUnreadCount() {
            const totalUnread = this.contacts.reduce((sum, contact) => {
                return sum + (Number(contact.unReadedmessage) || 0)
            }, 0)
            this.$store.commit('updateUnreadedMsg', totalUnread)
        },
        async GetUsList() {
            this.contacts = []
            await this.$store.dispatch('users/FetchUserFollowersFollowing')
            var glist = this.GetUserFollowersFollowing || []
            this.contacts = (glist || []).map(c => ({
                ...c,
                isOnline: !!c.isOnline,
                unReadedmessage: 0,
            }))
            if (this.contacts.length) {
                this.GetUnreadedMsgList()
            }
            this.updateOnlineList()
        },
        async selectUser(user) {
            if (!user?._id) return
            this.selectedUser = null
            this.messageBetweenUsers = []

            this.selectedUser = user
            this.messagelistnum = 0
            var firstuid = this.MainUserData._id
            var seconduid = user._id
            if (!firstuid) return
            var from = 0
            var ndata = { from, firstuid, seconduid }
            const result = await this.GetChatMsgsBetweenTwoUsers(ndata)
            const msgs = result?.msgs || []
            if (Array.isArray(msgs)) {
                this.messageBetweenUsers.push(...msgs)
            }

            let unreadSnapshot = 0
            this.contacts.forEach((contact) => {
                if (String(contact._id) === String(user._id)) {
                    unreadSnapshot = Number(contact.unReadedmessage) || 0
                }
            })

            setTimeout(() => {
                this.scrollDownFunction()
                this.CallMarkMsgAsReaded(user, unreadSnapshot)
                this.syncGlobalUnreadCount()
            }, 100)
        },
        async handleSendMessage() {
            var content = this.messageToSend.text
            var sender = this.MainUserData._id
            var receiver = String(this.selectedUser?._id)

            if (!content || !receiver || !sender) {
                return
            }

            var sdata = { content, sender, receiver }

            if (!this.uniqueOnlineUsers.includes(receiver)) {
                var savedMessage = await this.sendMessageAction(sdata)
                if (savedMessage) {
                    this.messageBetweenUsers.push(savedMessage)
                    this.messageToSend.text = ''
                    setTimeout(() => {
                        this.scrollDownFunction()
                    }, 100)
                }
            } else {
                const localMessage = { ...sdata, _id: `temp-${Date.now()}` }
                this.messageBetweenUsers.push(localMessage)
                this.messageToSend.text = ''
                this.SendPrivateMessage(sdata)
                setTimeout(() => {
                    this.scrollDownFunction()
                }, 100)
            }

        }

    },
}
</script>

<style lang="scss" scoped>
.chat-shell {
  display: flex;
  height: calc(100vh - 200px);
  min-height: 520px;
  overflow: hidden;
}

.chat-contact {
  width: 320px;
  flex: 0 0 320px;
  border-right: 1px solid rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.chat-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.chat-pane-header {
  background: #fff;
}

.message-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: #f8fafc;
}

// Mobile / tablet: stack list + conversation
@media (max-width: 1023px) {
  .chat-shell {
    flex-direction: column;
    height: calc(100vh - 160px);
    min-height: 480px;
  }

  .chat-contact {
    width: 100%;
    flex: 1 1 auto;
    border-right: none;
  }

  .chat-pane {
    width: 100%;
    height: 100%;
  }
}
</style>
