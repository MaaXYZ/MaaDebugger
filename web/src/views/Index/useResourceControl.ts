import { computed, watch } from "vue";
import { useResourceStore } from "@/stores/resource";
import { useSignalStore } from "@/stores/signal";
import { useDebugWorkspaceSettingsStore } from "@/stores/debugWorkspaceSettings";
import { getPipelineCheckResult, loadResource } from "@/api/http";
import type { CheckResponse } from "@/types/pipeline";

interface LoadResourceOptions {
  manual: boolean;
  changedPath: string;
}

export default function useResourceControl() {
  const toast = useToast();
  const resourceStore = useResourceStore();
  const enabledPaths = computed(() => resourceStore.getEnabledPaths());
  const signalStore = useSignalStore();
  const debugWorkspaceSettingsStore = useDebugWorkspaceSettingsStore();

  const WATCH_RESOURCE_TOAST_ID = "watch-resource";
  const resourceToastID = "resource-toast";
  const pipelineCheckerToastID = "pipeline-checker-toast";

  async function tryLoadResource(): Promise<{
    success: boolean;
    msg?: string;
  }> {
    const paths = enabledPaths.value;
    if (paths.length === 0) return { success: false, msg: "No enabled paths" };

    try {
      const result = await loadResource(paths);
      if (!result.succeed) {
        console.error("[Resource] Load failed:", result.msg);
        return { success: false, msg: result.msg };
      }
      signalStore.emitRefreshNode();
      return { success: true };
    } catch (err) {
      console.error("[Resource] Load failed:", err);
      return { success: false, msg: String(err) };
    }
  }

  async function LoadResource(
    options: LoadResourceOptions = { manual: true, changedPath: "" },
  ) {
    const { success, msg } = await tryLoadResource();

    if (options.manual) {
      if (!success) {
        console.error("[Resource] Load failed:", msg);
        toast.add({
          id: resourceToastID,
          title: "Resource Load Failed",
          description: msg || "Unknown error",
          icon: "i-lucide-circle-x",
          color: "error",
        });
      } else {
        toast.add({
          id: resourceToastID,
          title: "Resource Loaded",
          icon: "i-lucide-check-circle",
          color: "success",
        });
      }
    } else {
      toast.add({
        id: WATCH_RESOURCE_TOAST_ID,
        title: "Resource Changed",
        description: `The Resource will reload as ${options.changedPath} changed.`,
        icon: "i-lucide-loader",
        color: "info",
      });
    }

    if (success && debugWorkspaceSettingsStore.checkPipeline) {
      const pipelineCheckResult = await getPipelineCheckResult();
      const notifyLevel = debugWorkspaceSettingsStore.checkPipelineNotifyLevel;
      const filteredResult = pipelineCheckResult.filter((item) => {
        if (notifyLevel === "NULL") {
          return false;
        }
        if (notifyLevel === "ERROR") {
          return item.level === "error";
        }
        if (notifyLevel === "WARNING") {
          return item.level === "error" || item.level === "warning";
        }
      });

      if (filteredResult.length > 0) {
        const grouped = filteredResult.reduce(
          (acc, item) => {
            if (item.level === "error" || item.level === "warning") {
              acc[item.level].push(item);
            }
            return acc;
          },
          { error: [] as CheckResponse[], warning: [] as CheckResponse[] },
        );
        const toastType = grouped.error.length > 0 ? "error" : "warning";

        toast.add({
          id: pipelineCheckerToastID,
          progress: false,
          title: "Pipeline Check Results",
          description: `Found ${filteredResult.length} issues.`,
          icon: "i-lucide-alert-triangle",
          color: toastType,
          actions: [
            {
              icon: "i-lucide-info",
              label: "View Details",
              color: "neutral",
              variant: "outline",
              onClick: () => {
                openPipelineCheckDetails(grouped);
              },
            },
          ],
        });
      }
    }
  }

  function openPipelineCheckDetails(grouped: {
    error: CheckResponse[];
    warning: CheckResponse[];
  }) {
    // TODO: replace with UModal details view in follow-up.
    console.log("[PipelineChecker] errors:", grouped.error);
    console.log("[PipelineChecker] warnings:", grouped.warning);
  }

  watch(
    () => signalStore.reloadResource,
    (status) => {
      if (status > 0)
        void LoadResource({
          manual: false,
          changedPath: signalStore.changedPath,
        });
    },
  );

  return {
    enabledPaths,
    tryLoadResource,
    onLoadResource: LoadResource,
  };
}
