<template>
  <q-layout view="lHh Lpr lFf">
    <nav-bar />
    <q-page-container class="bg-grey-1">
      <router-view />

    </q-page-container>
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
    ...mapGetters("auth", ["GetUserData"])
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
  methods: {
    ...mapMutations("auth", ["SetData"]),
    ...mapActions("RealTimeNotify", ["connectToNotify","StopConnectionToNotify"]),
    ...mapActions(["createChatConnection","StopConnectionToChat"]),
  },
  mounted() {
    this.SetData()
    this.connectToNotify()
    this.createChatConnection()
    this.wasAuthed = !!this.GetUserData
  },
  beforeUnmount(){
    this.StopConnectionToNotify()
    this.StopConnectionToChat()
  },
  components: { NavBar }
}
</script>

<style lang="scss">
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

nav {
  padding: 30px;

  a {
    font-weight: bold;
    color: #2c3e50;

    &.router-link-exact-active {
      color: #42b983;
    }
  }
}
</style>
