<template>
   <q-header class="q-header-app" bordered>
      <q-toolbar class="constrain q-pa-sm">

         <!-- Brand: desktop / tablet -->
         <q-btn flat to="/" class="gt-sm">
            <q-icon left size="2.2em" name="eva-camera-outline" class="brand-icon" />
            <q-toolbar-title class="brand-font brand-text text-bold" style="font-size:1.5rem;">Home</q-toolbar-title>
         </q-btn>

         <q-separator class="gt-sm" vertical spaced />

         <!-- Search: desktop / tablet -->
         <q-toolbar-title class="text-center gt-sm">
            <q-input dense outlined class="search-input" label="Search"
               @keyup.enter="GoSearch($event)">
            </q-input>
         </q-toolbar-title>

         <!-- Chat -->
         <q-btn round flat
            v-show="GetUserData?.result"
            @click="GoToChat"
            :icon="chatUnreadCount > 0 ? 'eva-message-square' : 'eva-message-square-outline'"
            :color="chatUnreadCount > 0 ? 'accent' : undefined"
            class="gt-sm"
            >
            <q-badge v-if="chatUnreadCount > 0" color="negative" floating rounded :label="chatUnreadCount"/>
         </q-btn>

         <!-- Notifications -->
         <q-btn round flat
            v-show="GetUserData?.result"
            @click="GoToNotification"
            :icon="notificationNum > 0 ? 'eva-bell' : 'eva-bell-outline'"
            :color="notificationNum > 0 ? 'accent' : undefined"
            class="gt-sm"
            >
            <q-badge v-if="notificationNum > 0" floating color="negative" rounded :label="notificationNum"/>
         </q-btn>

         <!-- Dark mode toggle: desktop / tablet -->
         <q-btn round flat class="gt-sm q-ml-sm" @click="toggleDark"
            :icon="isDark ? 'eva-sun-outline' : 'eva-moon-outline'"
            :color="isDark ? undefined : 'grey-8'"
            aria-label="Toggle dark mode" />

         <!-- Avatar menu: desktop / tablet -->
         <q-btn v-show="GetUserData?.result" round flat class="gt-sm">
            <q-avatar size="36px" v-if="GetUserData?.result?.imageUrl">
               <img :src="GetUserData?.result?.imageUrl">
            </q-avatar>
            <q-avatar size="36px" v-else>
               <img :src="defaultAvatar">
            </q-avatar>
            <q-menu>
               <q-list style="min-width: 170px;">
                  <q-item clickable v-close-popup>
                     <q-item-section @click="Profile"><q-icon left name="eva-person-outline" size="18px"/> Profile</q-item-section>
                  </q-item>
                  <q-separator />
                  <q-item clickable v-close-popup @click="toggleDark">
                     <q-item-section>
                        <q-icon left :name="isDark ? 'eva-sun-outline' : 'eva-moon-outline'" size="18px" />
                        {{ isDark ? 'Light Mode' : 'Dark Mode' }}
                     </q-item-section>
                  </q-item>
                  <q-separator />
                  <q-item clickable v-close-popup @click="LogUserOut">
                     <q-item-section><q-icon left name="eva-log-out-outline" size="18px"/> Logout</q-item-section>
                  </q-item>
               </q-list>
            </q-menu>
         </q-btn>

         <!-- Brand + hamburger: mobile only -->
         <q-btn flat round icon="eva-menu-outline" @click="mobileMenuOpen = !mobileMenuOpen" class="lt-md btn-inherit" />
         <q-btn flat to="/" class="lt-md">
            <q-icon left size="1.8em" name="eva-camera-outline" class="brand-icon" />
            <q-toolbar-title class="brand-font brand-text text-bold" style="font-size:1.25rem;">Home</q-toolbar-title>
         </q-btn>

         <q-space class="lt-md" />

         <!-- Dark mode toggle: mobile -->
         <q-btn round flat class="lt-md" @click="toggleDark"
            :icon="isDark ? 'eva-sun-outline' : 'eva-moon-outline'"
            :color="isDark ? undefined : 'grey-8'" />

         <!-- Avatar: mobile only -->
         <q-btn v-show="GetUserData?.result" round flat class="lt-md" @click="handleProfileClick">
            <q-avatar size="32px" v-if="GetUserData?.result?.imageUrl">
               <img :src="GetUserData?.result?.imageUrl">
            </q-avatar>
            <q-avatar size="32px" v-else>
               <img :src="defaultAvatar">
            </q-avatar>
         </q-btn>

      </q-toolbar>

      <!-- Mobile drawer -->
      <q-drawer
         v-model="mobileMenuOpen"
         overlay
         :width="250"
         bordered
         class="q-drawer-app"
      >
         <div class="q-pa-md">
            <q-item class="q-mb-sm" v-if="GetUserData?.result">
               <q-item-section avatar>
                  <q-avatar size="48px">
                     <img :src="GetUserData?.result?.imageUrl || defaultAvatar">
                  </q-avatar>
               </q-item-section>
               <q-item-section>
                  <q-item-label class="text-weight-bold text-subtitle1">{{ GetUserData?.result?.name }}</q-item-label>
               </q-item-section>
            </q-item>

            <q-item>
               <q-item-section>
                  <q-input
                     dense
                     outlined
                     label="Search"
                     @keyup.enter="GoSearch($event); mobileMenuOpen = false"
                  />
               </q-item-section>
            </q-item>

            <q-separator class="q-my-sm" />

            <q-item clickable v-ripple @click="handleProfileClick" v-if="GetUserData?.result">
               <q-item-section avatar><q-icon name="eva-person-outline" color=""/></q-item-section>
               <q-item-section>Profile</q-item-section>
            </q-item>

            <q-item clickable v-ripple @click="handleChatClick" v-if="GetUserData?.result">
               <q-item-section avatar><q-icon name="eva-message-square-outline" color=""/></q-item-section>
               <q-item-section>Messages</q-item-section>
            </q-item>

            <q-item clickable v-ripple @click="handleNotificationClick" v-if="GetUserData?.result">
               <q-item-section avatar><q-icon name="eva-bell-outline" color=""/></q-item-section>
               <q-item-section>Notifications</q-item-section>
            </q-item>

            <q-separator class="q-my-sm" />

            <q-item clickable v-ripple @click="toggleDark">
               <q-item-section avatar>
                  <q-icon :name="isDark ? 'eva-sun-outline' : 'eva-moon-outline'" color="" />
               </q-item-section>
               <q-item-section>{{ isDark ? 'Light Mode' : 'Dark Mode' }}</q-item-section>
            </q-item>

            <q-item clickable v-ripple @click="handleLogoutClick" v-if="GetUserData?.result">
               <q-item-section avatar><q-icon name="eva-log-out-outline" color="negative"/></q-item-section>
               <q-item-section>Logout</q-item-section>
            </q-item>

            <q-item clickable v-ripple to="/Auth" v-else>
               <q-item-section avatar><q-icon name="eva-log-in-outline" color=""/></q-item-section>
               <q-item-section>Login</q-item-section>
            </q-item>
         </div>
      </q-drawer>

   </q-header>
</template>

<script>
import { mapGetters, mapMutations, mapActions, mapState } from 'vuex';
import { debounce } from '@/utils/timing';

const DEFAULT_AVATAR = 'https://game-1255653016.file.myqcloud.com/manage/compress/custom_wzry_E1/312ff4442ddbe69154045e33b604ef56.jpg?imageMogr2/crop/512x512/gravity/center';

export default {
   name: 'NavBar',
   data() {
      return {
         mobileMenuOpen: false,
         defaultAvatar: DEFAULT_AVATAR
      }
   },
   computed: {
      ...mapGetters("auth", ["GetUserData"]),
      ...mapGetters(["getUnReadedMsg"]),
      ...mapGetters(["GetUnReadedNotification"]),
      chatUnreadCount() {
         return this.getUnReadedMsg || 0
      },
      notificationNum() {
         return this.GetUnReadedNotification || 0
      },
      isDark() {
         return this.$q.dark.isActive
      },
      ...mapState(['RealTimeNotify', 'RealTimeChat'])
   },
   watch: {
      GetUserData: {
         async handler(newVal, oldVal) {
            const nextUserId = newVal?.result?._id
            const prevUserId = oldVal?.result?._id

            if (String(nextUserId || '') === String(prevUserId || '')) {
               return
            }

            await this.bootstrapUnreadCounts()
         },
         immediate: true,
      },
      "RealTimeNotify.notifyideslistNumber": async function () {
         this.debouncedUnreadNotifyCount()
      },
   },
   methods: {
      ...mapMutations("auth", ["SetData"]),
      ...mapActions("auth", ["logout"]),
      ...mapActions(["GetUnReadedNotifyNum", "GetUnreadedMessageNum"]),
      ...mapActions("RealTimeNotify", ["StopConnectionToNotify"]),
      ...mapActions(["StopConnectionToChat"]),
      handleProfileClick() {
         this.mobileMenuOpen = false
         this.Profile()
      },
      handleChatClick(){
         this.mobileMenuOpen = false
         this.GoToChat()
      },
      handleNotificationClick(){
         this.mobileMenuOpen = false
         this.GoToNotification()
      },
      handleLogoutClick(){
         this.mobileMenuOpen = false
         this.LogUserOut()
      },
      toggleDark() {
         const next = !this.$q.dark.isActive
         this.$q.dark.set(next)
         localStorage.setItem('dsh-dark', next ? '1' : '0')
      },
      GoSearch(e) {
         console.log("go", e.target.value)
         this.$router.push({ path: `/Search`, query: { search: e.target.value } })
      },
      Profile() {
         let id = this.GetUserData?.result?._id;
         if (id) {
            this.$router.push(`/profile/${id}`);
         }
      },
      LogUserOut() {
         this.logout();
         this.StopConnectionToNotify()
         this.StopConnectionToChat()
         this.$router.push("/Auth");
      },
      GoToNotification() {
         this.$router.push('/Notification')
      },
      GoToChat() {
         this.$router.push('/Chat')
      },
      async bootstrapUnreadCounts() {
         await this.UNreadedNotifyCount()
         await this.unreadMessageCount()
      },
      async UNreadedNotifyCount() {
         const userId = this.GetUserData?.result?._id
         if (!userId) {
            this.$store.commit('updateUnReadedNotification', 0)
            return
         }
         await this.GetUnReadedNotifyNum(userId)
      },
      async unreadMessageCount() {
         const userId = this.GetUserData?.result?._id
         if (!userId) {
            this.$store.commit('updateUnreadedMsg', 0)
            return
         }
         const { totalUnreadMessageCount } = await this.GetUnreadedMessageNum(userId)
         this.$store.commit('updateUnreadedMsg', totalUnreadMessageCount || 0)
      },
   },
   async mounted() {
      this.SetData();
      // 初始化防抖：WebSocket 连续通知（连接重放 / 快速推送）合并成一次请求。
      // 首次加载与登录状态变化的刷新由 GetUserData 的 immediate watcher 触发，
      // 这里不再重复 bootstrap，避免同一批未读接口请求两遍。
      this.debouncedUnreadNotifyCount = debounce(this.UNreadedNotifyCount, 1000)
   },
}

</script>

<style lang="sass">
.search-input
   width: 100%
   max-width: 420px
   margin: 0 auto
   .q-field__control
      border-radius: 999px

.q-header-app
   background: #fff
   // !important：Quasar 的 marginal header 强制 color:#fff，
   // 不加的话未选中的聊天/通知/主题切换图标会继承纯白（亮色下看不见）
   color: #0F172A !important

.q-drawer-app
   background: #fff
   color: #0F172A

body.body--dark
   .q-header-app
      background: #0F172A
      color: #E2E8F0 !important
   .q-drawer-app
      background: #101828
      color: #E2E8F0
</style>
