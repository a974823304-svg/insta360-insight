import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 对应架构文档：前端可视化交互层
// /api 代理到 Go 后端 (localhost:8080)
export default defineConfig({
  // 相对 base: 部署到任意子路径/子域名都能正确加载静态资源, 避免空白页
  base: './',
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/avatars': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
