import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  return {
    plugins: [react()],
    server: {
      port: 5173,
      proxy: {
        "/read-api": {
          target: env.VITE_READ_API_PROXY_TARGET || "http://127.0.0.1:8080",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/read-api/, ""),
        },
        "/auth-api": {
          target: env.VITE_AUTH_API_PROXY_TARGET || "http://127.0.0.1:8081",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/auth-api/, ""),
        },
        "/exchange-core-api": {
          target:
            env.VITE_EXCHANGE_CORE_API_PROXY_TARGET || "http://127.0.0.1:8082",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/exchange-core-api/, ""),
        },
      },
    },
  };
});
