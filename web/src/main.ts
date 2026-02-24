import { createApp } from "vue";
import { createPinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import ui from "@nuxt/ui/vue-plugin";
import App from "./App.vue";
import "./style.css";

const router = createRouter({
  routes: [
    {
      path: "/settings",
      component: () => import("./router/SettingsPage.vue"),
    },
  ],
  history: createWebHistory(),
});
const pinia = createPinia();

const app = createApp(App);

app.use(ui);
app.use(router);
app.use(pinia);

app.mount("#app");
