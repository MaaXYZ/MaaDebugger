import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import ui from "@nuxt/ui/vite";
import IconifyIcons from "./plugins/icon-loader";

import path from "path";

export default defineConfig({
  plugins: [
    vue(),
    ui(),
    IconifyIcons({
      // 白名单用于兜底动态场景（无法被静态扫描捕获）
      appendList: {
        "simple-icons": ["xbox", "playstation"],
      },
      // 仅 build 生效：白名单条目缺失时抛出错误，防止图标丢失
      whitelistCheck: {
        enabled: true,
        throwOnMissing: true,
      },
      production: {
        pruneIcons: true,
        keepAliases: true,
        dropMeta: true,
      },
    }),
  ],
  resolve: {
    alias: [
      {
        find: "@",
        replacement: path.resolve(__dirname, "src"),
      },
      {
        find: /^monaco-editor$/,
        replacement: path.resolve(
          __dirname,
          "node_modules/monaco-editor/esm/vs/editor/editor.api.js",
        ),
      },
      {
        find: "monaco-editor/esm/vs/editor/editor.main.js",
        replacement: path.resolve(
          __dirname,
          "node_modules/monaco-editor/esm/vs/editor/editor.api.js",
        ),
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
