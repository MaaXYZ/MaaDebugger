import { defineStore } from "pinia";
import { ref } from "vue";

export const useTaskStore = defineStore(
  "task",
  () => {
    const selectedEntry = ref("");
    /** Pipeline override JSON string (JSONC), persisted */
    const overrideJson = ref("{}");

    return {
      selectedEntry,
      overrideJson,
    };
  },
  { persist: true },
);
