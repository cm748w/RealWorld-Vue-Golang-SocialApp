const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
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
