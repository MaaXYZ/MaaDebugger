import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import ui from "@nuxt/ui/vite";

import path from "path";

export default defineConfig({
  plugins: [vue(), ui()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  build: {
    outDir: path.resolve(__dirname, "../server/frontend/dist"),
    emptyOutDir: true,
    chunkSizeWarningLimit: 10240,
  },
  server: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8011",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://127.0.0.1:8011",
        ws: true,
      },
    },
  },
});
