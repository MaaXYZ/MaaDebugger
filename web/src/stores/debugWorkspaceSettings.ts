import { ref } from "vue";
import { defineStore } from "pinia";

export const useDebugWorkspaceSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoCollapseLeftTabsOnRunStart = ref(true);
    const leftTabsCollapsed = ref(false);
    const watchResourceChange = ref(true);
    const watchResourceChangeInterval = ref(1000);

    function setAutoCollapseLeftTabsOnRunStart(value: boolean) {
      autoCollapseLeftTabsOnRunStart.value = value;
    }

    function setLeftTabsCollapsed(value: boolean) {
      leftTabsCollapsed.value = value;
    }

    function setWatchResourceChange(value: boolean) {
      watchResourceChange.value = value;
    }
    function setWatchResourceChangeInterval(value: number) {
      watchResourceChangeInterval.value = value;
    }

    function reset() {
      autoCollapseLeftTabsOnRunStart.value = true;
      leftTabsCollapsed.value = false;
      watchResourceChange.value = true;
      watchResourceChangeInterval.value = 500;
    }

    return {
      autoCollapseLeftTabsOnRunStart,
      leftTabsCollapsed,
      watchResourceChange,
      watchResourceChangeInterval,
      setAutoCollapseLeftTabsOnRunStart,
      setLeftTabsCollapsed,
      setWatchResourceChange,
      setWatchResourceChangeInterval,
      reset,
    };
  },
  { persist: true },
);
