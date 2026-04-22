import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())

  return {
    base: env.VITE_ROUTER_BASE || '/',
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    define: {
      'process.env': env,
    },
    server: {
      port: 5174,
      strictPort: false,
      open: true,
      cors: true,
      proxy: {
        '/app': {
          target: env.VITE_API_BASE_URL || 'http://localhost:5555',
          changeOrigin: true,
          rewrite: (requestPath) => requestPath.replace(/^\/app/, '/app'),
        },
      },
    },
    build: {
      target: 'ES2022',
      outDir: 'dist',
      assetsDir: 'assets',
    },
  }
})
