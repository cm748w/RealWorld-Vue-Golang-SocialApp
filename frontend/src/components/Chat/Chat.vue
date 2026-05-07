<template>
    <q-page class="constrain q-pa-md">
        <div class="row q-col-gutter-lg">
            <div class="col-12 chat-container">
                <div class="user-list">
                    <div class="q-pa-md">
                        <q-toolbar class="bg-primary text-white shadow-1">
                            <q-toolbar-title>Following & Following</q-toolbar-title>
                        </q-toolbar>

                        <q-list bordered>
                            <q-item @click="selectUser(contact)" v-for="contact in contacts" :key="contact._id"
                                class="q-my-sm" clickable v-ripple>
                                <q-item-section avatar>
                                    <q-avatar v-if="!contact.imageUrl" color="primary" text-color="white">
                                        {{ contact.name[0] }}
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
                </div>

                <!-- chat box -->
                <div class="chat-messages" v-if="selectedUser != null" style="background: white;">
                    <div class="q-pa-md row justify-center" style="overflow-y: auto; max-height: 400px;"
                        ref="messageContainer" @scroll="handleScroll">
                        <div v-for="msg in messageBetweenUsers" :key="msg._id" style="width: 100%;">
                            <q-chat-message
                                :name="msg.sender === MainUserData._id ? MainUserData.name : selectedUser.name"
                                :avatar="msg.sender === MainUserData._id ? MainUserData.imageUrl : selectedUser.imageUrl"
                                :text="[msg.content]" :sent="msg.sender === MainUserData._id ? true : false" />
                        </div>
                    </div>

                    <q-separator spaced />
                    <q-input outlined v-model="messageToSend.text" @keyup.enter="handleSendMessage"
                        label="write message..">
                        <q-btn v-if="messageToSend.text != ''" @click="handleSendMessage" flat round color="primary"
                            icon="eva-arrow-right" />
                    </q-input>

                </div>

            </div>
        </div>
    </q-page>


</template>

<script>
import { mapActions, mapGetters, mapState } from 'vuex';
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
        }
    },
    computed: {
        ...mapGetters('users', ['GetUserFollowersFollowing', 'GetUser']),
        ...mapGetters('auth', ['GetUserData']),
        ...mapState(["RealTimeChat"]),
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
                    setTimeout(() => {
                        this.scrollDownFunction()
                    }, 100)
                } else {
                    this.contacts.forEach((contact) => {
                        if (contact._id == message.sender) {
                            contact.unReadedmessage++
                        }
                    })
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
                // scrolled to the top
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
        async CallMarkMsgAsReaded(user) {
            var mainuid = this.MainUserData._id
            var otheruid = user._id
            var GetunReadedmessage = 0

            this.contacts.forEach(
                user => {
                    if (String(otheruid) == String(user._id)) {
                        GetunReadedmessage = user.unReadedmessage
                    }
                }
            )

            var data = { mainuid, otheruid, GetunReadedmessage }
            var { isMarked } = await this.MarkMsgsAsReaded(data)

            if (isMarked) {
                this.contacts.forEach(user => {
                    if (String(otheruid) == String(user._id)) {
                        user.unReadedmessage = 0
                    }
                })
            }
        },
        async GetUnreadedMsgList() {
            if (!this.MainUserData._id) return
            const unreadPayload = await this.GetUnreadedMessageNum(this.MainUserData._id)
            var messages = unreadPayload?.messages || []
            this.contacts.forEach(user => {
                messages.forEach(msg => {
                    if (String(msg.otherUserId) == String(user._id)) {
                        user.unReadedmessage = Number(msg.numOfUnreadMessages)
                    }
                })
            })
        },
        async GetUsList() {
            this.contacts = []
            // Call the FetchUserFollowersFollowing action to fetch and cache the data
            await this.$store.dispatch('users/FetchUserFollowersFollowing')
            // Now read from the getter
            var glist = this.GetUserFollowersFollowing || []
            // ensure fields exist so Vue reactivity works (Vue2 compatibility)
            this.contacts = (glist || []).map(c => ({
                ...c,
                isOnline: !!c.isOnline,
                unReadedmessage: c.unReadedmessage ? Number(c.unReadedmessage) : 0,
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
            setTimeout(() => {
                this.scrollDownFunction()
                this.CallMarkMsgAsReaded(user)
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

<style scoped>
.chat-container {
    display: flex;
}

.chat-messages {
    flex: 1;
    padding: 10px;
}
</style>