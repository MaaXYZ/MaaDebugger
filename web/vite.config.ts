import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import ui from "@nuxt/ui/vite";

import path from "path";

export default defineConfig({
  plugins: [vue(), ui()],
  resolve: {
    alias: [
      {
        find: "@",
        replacement: path.resolve(__dirname, "src"),
      },
      {
        find: /^monaco-editor$/,
        replacement: path.resolve(__dirname, "node_modules/monaco-editor/esm/vs/editor/editor.api.js"),
      },
      {
        find: "monaco-editor/esm/vs/editor/editor.main.js",
        replacement: path.resolve(__dirname, "node_modules/monaco-editor/esm/vs/editor/editor.api.js"),
      },
    ],
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
