import { ref } from "vue";
import { defineStore } from "pinia";

export const useDebugWorkspaceSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoCollapseLeftTabsOnRunStart = ref(true);
    const leftTabsCollapsed = ref(false);

    function setAutoCollapseLeftTabsOnRunStart(value: boolean) {
      autoCollapseLeftTabsOnRunStart.value = value;
    }

    function setLeftTabsCollapsed(value: boolean) {
      leftTabsCollapsed.value = value;
    }

    function reset() {
      autoCollapseLeftTabsOnRunStart.value = true;
      leftTabsCollapsed.value = false;
    }

    return {
      autoCollapseLeftTabsOnRunStart,
      leftTabsCollapsed,
      setAutoCollapseLeftTabsOnRunStart,
      setLeftTabsCollapsed,
      reset,
    };
  },
  { persist: true },
);
