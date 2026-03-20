import { ref } from "vue";
import { defineStore } from "pinia";

export const useDebugWorkspaceSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoHideLeftTabsWhenRunning = ref(true);
    const leftTabsCollapsed = ref(false);

    function setAutoHideLeftTabsWhenRunning(value: boolean) {
      autoHideLeftTabsWhenRunning.value = value;
    }

    function setLeftTabsCollapsed(value: boolean) {
      leftTabsCollapsed.value = value;
    }

    function reset() {
      autoHideLeftTabsWhenRunning.value = true;
      leftTabsCollapsed.value = false;
    }

    return {
      autoHideLeftTabsWhenRunning,
      leftTabsCollapsed,
      setAutoHideLeftTabsWhenRunning,
      setLeftTabsCollapsed,
      reset,
    };
  },
  { persist: true },
);
