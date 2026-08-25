import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

/**
 * Vite 配置
 * - dev server 端口 5173
 * - /api 与 /mcp 代理到本地 Go 后端（8080），
 *   开发期同源请求，彻底规避跨域问题（后端 CORS 也已放开，双保险）
 */
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)), // @ → src 目录别名
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/mcp': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
