import { ref } from "vue";
import { defineStore } from "pinia";

export const useUpdateSettingsStore = defineStore(
  "updateSettings",
  () => {
    const showPreRelease = ref(false);

    function setShowPreRelease(value: boolean) {
      showPreRelease.value = value;
    }

    function reset() {
      showPreRelease.value = false;
    }

    return {
      showPreRelease,
      setShowPreRelease,
      reset,
    };
  },
  { persist: true },
);
