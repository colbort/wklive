import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBaseUrl = env.VITE_API_BASE_URL || '/chat/admin'
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://127.0.0.1:6666'

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 5178,
      strictPort: false,
      open: true,
      cors: true,
      proxy: {
        [apiBaseUrl]: {
          target: proxyTarget,
          changeOrigin: true,
          ws: true,
        },
        '/avatars': {
          target: proxyTarget,
          changeOrigin: true,
        },
        '/chat_uploads': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
