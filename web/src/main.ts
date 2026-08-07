import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import ui from "@nuxt/ui/vue-plugin";

import App from "./App.vue";
import { serverPersistPlugin } from "./stores/persist";
import { useLocaleStore } from "./stores/locale";
import { i18n, guessInitialLocale } from "./i18n";
import "./style.css";
import { AllIcons } from "virtual:load-iconify-icons";
AllIcons();

const router = createRouter({
  routes: [
    {
      path: "/",
      component: () => import("./router/Index.vue"),
    },
    {
      path: "/settings",
      component: () => import("./router/SettingsPage.vue"),
    },
  ],
  history: createWebHistory(),
  scrollBehavior(to, _from, _savedPosition) {
    if (to.hash) {
      return new Promise((resolve) => {
        setTimeout(() => {
          const targetEl = document.querySelector(
            to.hash,
          ) as HTMLElement | null;

          if (targetEl) {
            // 补全 CSS 动画：直接添加动画 class 或操作 style
            targetEl.style.transition = "none";
            targetEl.style.backgroundColor = "#b9af63";

            setTimeout(() => {
              targetEl.style.transition = "background-color 1.5s ease-out";
              targetEl.style.backgroundColor = "transparent";
            }, 50);

            resolve({
              el: to.hash,
              top: 96,
              behavior: "smooth",
            });
          } else {
            resolve({ top: 96 });
          }
        }, 50);
      });
    }
    return { top: 0 };
  },
});
const pinia = createPinia();
pinia.use(serverPersistPlugin);

async function bootstrap() {
  const app = createApp(App);

  app.use(ui);
  app.use(router);
  app.use(pinia);
  app.use(i18n);

  // 挂载前先解析界面语言：后端决定默认语言（用户持久化选择优先，
  // 否则按 Accept-Language 检测，失败回退英语）。先设置乐观初始语言避免首屏闪烁。
  i18n.global.locale.value = guessInitialLocale();
  await useLocaleStore().apply();

  app.mount("#app");
}

void bootstrap();
