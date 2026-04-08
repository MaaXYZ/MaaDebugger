import { ref } from "vue";
import type { Ref } from "vue";
import { defineStore } from "pinia";
import { StartPipelineCheck, StopPipelineCheck } from "@/api/http";

type PipelineNotifyLevel = "NULL" | "ERROR" | "WARNING";

export const useDebugSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoCollapseLeftTabsOnRunStart = ref(true);
    const leftTabsCollapsed = ref(false);
    const watchResourceChange = ref(true);
    const watchResourceChangeInterval = ref(1000);
    const checkPipeline = ref(true);
    const checkPipelineNotifyLevel: Ref<PipelineNotifyLevel> = ref("ERROR");
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
      void syncPipelineChecker(value);
    }

    async function syncPipelineChecker(enabled: boolean = checkPipeline.value) {
      try {
        if (enabled) {
          await StartPipelineCheck();
          return;
        }
        await StopPipelineCheck();
      } catch (err) {
        console.error("[PipelineChecker] Failed to sync running state:", err);
      }
    }

    function setCheckPipelineNotifyLevel(value: PipelineNotifyLevel) {
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
      syncPipelineChecker,
      setCheckPipelineNotifyLevel,
      setPreventRunning,
      reset,
    };
  },
  { persist: true },
);
