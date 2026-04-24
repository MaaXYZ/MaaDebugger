import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import ui from "@nuxt/ui/vue-plugin";

import App from "./App.vue";
import { serverPersistPlugin } from "./stores/persist";
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

const app = createApp(App);

app.use(ui);
app.use(router);
app.use(pinia);

app.mount("#app");
