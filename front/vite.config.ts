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
});
