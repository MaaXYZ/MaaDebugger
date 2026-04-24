import { ref } from "vue";
import type { Ref } from "vue";
import { defineStore } from "pinia";
import { StartPipelineCheck, StopPipelineCheck } from "@/api/http";

export type PipelineNotifyLevel = "NULL" | "ERROR" | "WARNING";

export const useDebugSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    const autoCollapseLeftTabsOnRunStart = ref(true);
    const leftTabsCollapsed = ref(false);
    const watchResourceChange = ref(true);
    const watchResourceChangeInterval = ref(1000);
    const checkPipeline = ref(true);
    const checkPipelineNotifyLevel: Ref<PipelineNotifyLevel> = ref("ERROR");
    const preventResourceLoaded = ref(true);

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
          const result = await StartPipelineCheck();
          if (!result.succeed) {
            throw new Error(result.msg || "Start pipeline checker failed");
          }
          return;
        }
        const result = await StopPipelineCheck();
        if (!result.succeed) {
          throw new Error(result.msg || "Stop pipeline checker failed");
        }
      } catch (err) {
        console.error("[PipelineChecker] Failed to sync running state:", err);
      }
    }

    function setCheckPipelineNotifyLevel(value: PipelineNotifyLevel) {
      checkPipelineNotifyLevel.value = value;
    }

    function setPreventResourceLoaded(value: boolean) {
      preventResourceLoaded.value = value;
    }

    function reset() {
      autoCollapseLeftTabsOnRunStart.value = true;
      leftTabsCollapsed.value = false;
      watchResourceChange.value = true;
      watchResourceChangeInterval.value = 1000;
      checkPipeline.value = true;
      checkPipelineNotifyLevel.value = "ERROR";
      preventResourceLoaded.value = true;
      void syncPipelineChecker(true);
    }

    return {
      autoCollapseLeftTabsOnRunStart,
      leftTabsCollapsed,
      watchResourceChange,
      watchResourceChangeInterval,
      checkPipeline,
      checkPipelineNotifyLevel,
      preventResourceLoaded,
      setAutoCollapseLeftTabsOnRunStart,
      setLeftTabsCollapsed,
      setWatchResourceChange,
      setWatchResourceChangeInterval,
      setCheckPipeline,
      syncPipelineChecker,
      setCheckPipelineNotifyLevel,
      setPreventResourceLoaded,
      reset,
    };
  },
  { persist: true },
);
