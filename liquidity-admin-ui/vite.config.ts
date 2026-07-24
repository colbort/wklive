import { fileURLToPath, URL } from "node:url";
import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiBase = env.VITE_API_BASE_URL || "/liquidity/admin";
  const target = env.VITE_PROXY_TARGET || "http://127.0.0.1:8890";
  return {
    plugins: [vue()],
    resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
    server: {
      host: "0.0.0.0",
      port: 5180,
      proxy: { [apiBase]: { target, changeOrigin: true } },
    },
  };
});
