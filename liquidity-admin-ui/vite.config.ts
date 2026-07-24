import { fileURLToPath, URL } from "node:url";
import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const proxyTarget = env.VITE_PROXY_TARGET || "http://127.0.0.1:9999";
  return {
    base: env.VITE_ROUTER_BASE || "/",
    plugins: [vue()],
    resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
    server: {
      host: "0.0.0.0",
      port: 5180,
      strictPort: false,
      open: true,
      cors: true,
      proxy: {
        "/admin/liquidity": {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
