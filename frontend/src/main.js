import { createApp } from 'vue'
import App from './App.vue'
import { Quasar } from 'quasar'
import quasarUserOptions from './quasar-user-options'
import router from './router'
import store from './store'

// 抑制 Chrome 的良性 “ResizeObserver loop completed…” 报错噪音。
// 该提示通常在字体/图片加载引发布局变化时出现，不影响功能。
window.addEventListener('error', (event) => {
  if (event && typeof event.message === 'string' && event.message.includes('ResizeObserver loop')) {
    event.preventDefault()
    event.stopImmediatePropagation()
  }
}, true)

createApp(App).use(store).use(router).use(Quasar, quasarUserOptions).mount('#app')
