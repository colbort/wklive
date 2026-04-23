import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  const isCapacitor = env.VITE_APP_TARGET === 'capacitor'

  return {
    base: isCapacitor ? './' : env.VITE_ROUTER_BASE || '/',
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
      host: '0.0.0.0',
      port: 5174,
      strictPort: false,
      open: true,
      cors: true,
      proxy: {
        '/app': {
          target: env.VITE_API_BASE_URL || 'http://localhost:5555',
          changeOrigin: true,
          ws: true,
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
