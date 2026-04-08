import { ref } from "vue";
import { defineStore } from "pinia";

export const useDebugWorkspaceSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoCollapseLeftTabsOnRunStart = ref(true);
    const leftTabsCollapsed = ref(false);
    const watchResourceChange = ref(true);
    const watchResourceChangeInterval = ref(1000);
    const checkPipeline = ref(true);
    const checkPipelineNotifyLevel = ref("ERROR");
    const preventRunning = ref(true);

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

    function setCheckPipeline(value: boolean) {
      checkPipeline.value = value;
    }

    function setCheckPipelineNotifyLevel(value: string) {
      checkPipelineNotifyLevel.value = value;
    }

    function setPreventRunning(value: boolean) {
      preventRunning.value = value;
    }

    function reset() {
      autoCollapseLeftTabsOnRunStart.value = true;
      leftTabsCollapsed.value = false;
      watchResourceChange.value = true;
      watchResourceChangeInterval.value = 500;
      checkPipeline.value = true;
      checkPipelineNotifyLevel.value = "ERROR";
      preventRunning.value = true;
    }

    return {
      autoCollapseLeftTabsOnRunStart,
      leftTabsCollapsed,
      watchResourceChange,
      watchResourceChangeInterval,
      checkPipeline,
      checkPipelineNotifyLevel,
      preventRunning,
      setAutoCollapseLeftTabsOnRunStart,
      setLeftTabsCollapsed,
      setWatchResourceChange,
      setWatchResourceChangeInterval,
      setCheckPipeline,
      setCheckPipelineNotifyLevel,
      setPreventRunning,
      reset,
    };
  },
  { persist: true },
);
