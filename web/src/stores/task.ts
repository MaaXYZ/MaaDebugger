import { defineStore } from "pinia";
import { ref } from "vue";

export const useTaskStore = defineStore(
  "task",
  () => {
    const selectedEntry = ref("");

    return {
      selectedEntry,
    };
  },
  { persist: true },
);
