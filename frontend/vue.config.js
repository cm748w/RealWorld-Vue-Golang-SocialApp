const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  // 前端整体挂在 /app/ 基路径下：构建产物中的静态资源（js/css/图片等）都会带 /app 前缀，
  // 同时 Vue CLI 会把 process.env.BASE_URL 置为 publicPath，Router 的 history base 自动变为 /app/，
  // 页面路由（如 /app/Auth、/app/profile/:id）由外层网关统一反代到本服务。
  // 注意：此改动不涉及 /api、/ws-* 等后端接口路径，它们仍保持站点根路径，由 nginx/网关反代。
  publicPath: '/app/',
  transpileDependencies: [
    'quasar'
  ],

  pluginOptions: {
    quasar: {
      importStrategy: 'kebab',
      rtlSupport: false
    }
  },

  // 本地开发代理：前端代码统一使用相对路径（/api、/ws-notify、/ws-chat），
  // 与 Docker 环境下 nginx 反代的路径保持一致；此处让 dev server 在本地开发时
  // 把这些路径转发到对应的后端服务，避免 localhost 写死在代码里。
  devServer: {
    // 仅在本地开发时生效：关闭“运行时错误”全屏遮罩。
    // ResizeObserver loop 等浏览器良性提示经常被误判为运行时错误而弹出遮罩；
    // 真正的编译错误 (errors) 仍会显示，真实运行时错误仍会记录到浏览器控制台。
    client: {
      overlay: {
        errors: true,
        warnings: false,
        runtimeErrors: false
      }
    },
    proxy: {
      '/api': {
        target: 'http://localhost:5000',
        pathRewrite: { '^/api': '' },
        changeOrigin: true
      },
      '/ws-notify': {
        target: 'ws://localhost:8088',
        ws: true,
        pathRewrite: { '^/ws-notify': '/ws' },
        changeOrigin: true
      },
      '/ws-chat': {
        target: 'ws://localhost:8001',
        ws: true,
        pathRewrite: { '^/ws-chat': '/ws' },
        changeOrigin: true
      }
    }
  }
})
