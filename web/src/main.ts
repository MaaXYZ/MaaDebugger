import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import ui from "@nuxt/ui/vue-plugin";
import App from "./App.vue";
import { serverPersistPlugin } from "./stores/persist";
import "./style.css";

const router = createRouter({
  routes: [
    {
      path: "/",
      component: () => import("./router/Index.vue"),
    },
    {
      path: "/TaskDetail",
      component: () => import("./router/TaskDetailPage.vue"),
    },
    {
      path: "/settings",
      component: () => import("./router/SettingsPage.vue"),
    },
  ],
  history: createWebHistory(),
});
const pinia = createPinia();
pinia.use(serverPersistPlugin);

const app = createApp(App);

app.use(ui);
app.use(router);
app.use(pinia);

app.mount("#app");
