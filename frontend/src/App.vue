<template>
  <q-layout view="hHh LpR lFf" class="app-shell">
    <nav-bar />

    <q-page-container>
      <router-view />
    </q-page-container>

    <!-- Mobile / tablet bottom navigation -->
    <q-footer
      v-if="isAuthed && isMobileNav"
      class="bottom-nav q-py-xs"
    >
      <div class="row justify-around q-pa-xs">
        <q-btn
          flat round
          size="md"
          :to="'/'"
          :icon="ActiveRoute('/') ? 'eva-home' : 'eva-home-outline'"
          :color="ActiveRoute('/') ? 'accent' : undefined"
          aria-label="Home"
        />
        <q-btn
          flat round
          size="md"
          :to="profilePath"
          :icon="ActiveRoute('/profile') ? 'eva-person' : 'eva-person-outline'"
          :color="ActiveRoute('/profile') ? 'accent' : undefined"
          aria-label="Profile"
        />
        <q-btn
          flat round
          size="md"
          :to="'/Chat'"
          :icon="ActiveRoute('/Chat') ? 'eva-message-square' : 'eva-message-square-outline'"
          :color="ActiveRoute('/Chat') ? 'accent' : undefined"
          aria-label="Messages"
        />
        <q-btn
          flat round
          size="md"
          :to="'/Notification'"
          :icon="ActiveRoute('/Notification') ? 'eva-bell' : 'eva-bell-outline'"
          :color="ActiveRoute('/Notification') ? 'accent' : undefined"
          aria-label="Notifications"
        />
      </div>
    </q-footer>
  </q-layout>
</template>

<script>
import { mapActions, mapMutations, mapGetters } from "vuex"
import NavBar from './views/NavBar.vue';
export default {
  name: "MainLayout",
  data() {
    return {
      wasAuthed: false
    }
  },
  computed: {
    ...mapGetters("auth", ["GetUserData"]),
    isAuthed() {
      return !!this.GetUserData?.result
    },
    isMobileNav() {
      // Bottom navigation only below md (phone + tablet)
      return this.$q.screen.lt.md
    },
    profilePath() {
      const id = this.GetUserData?.result?._id
      return id ? `/profile/${id}` : '/'
    }
  },
  methods: {
    ...mapMutations("auth", ["SetData"]),
    ...mapActions("RealTimeNotify", ["connectToNotify","StopConnectionToNotify"]),
    ...mapActions(["createChatConnection","StopConnectionToChat"]),
    ActiveRoute(path) {
      if (path === '/') return this.$route.path === '/'
      return this.$route.path.startsWith(path)
    },
    applyTheme() {
      const dark = localStorage.getItem('dsh-dark') === '1'
      this.$q.dark.set(dark)
    }
  },
  watch: {
    GetUserData(newVal) {
      const isNowAuthed = !!newVal

      // 如果从已登录变成未登录（登出了），跳转到登录页
      if (this.wasAuthed && !isNowAuthed) {
        this.$router.push('/Auth')
      }

      // 如果从未登录变成已登录（登入成功）且当前在认证页，跳转到首页
      if (!this.wasAuthed && isNowAuthed && this.$route.path === '/Auth') {
        this.$router.push('/')
      }

      // 更新登录状态标记
      this.wasAuthed = isNowAuthed
    }
  },
  mounted() {
    this.SetData()
    this.connectToNotify()
    this.createChatConnection()
    this.wasAuthed = !!this.GetUserData
    this.applyTheme()
  },
  beforeUnmount(){
    this.StopConnectionToNotify()
    this.StopConnectionToChat()
  },
  components: { NavBar }
}
</script>

<style lang="scss">
.app-shell {
  background-color: #FFFFFF;

  .q-page-container {
    // Give room so the mobile bottom nav never hides content
    min-height: 100%;
  }
}

#app {
  font-family: 'Inter', 'Roboto', Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #1e293b;
}

// Dark mode shell overrides
body.body--dark {
  .app-shell {
    background-color: #0B1120;
  }

  #app {
    color: #CBD5E1;
  }
}

@media (max-width: 1023px) {
  .q-layout {
    // Compensate the bottom navigation height
    .q-page-container {
      padding-bottom: 64px;
    }
  }
}
</style>
