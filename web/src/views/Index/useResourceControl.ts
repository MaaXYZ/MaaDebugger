import { computed, ref } from "vue";
import type { Ref } from "vue";
import { useResourceStore } from "@/stores/resource";
import { useSignalStore } from "@/stores/signal";
import { useDebugSettingsStore } from "@/stores/debugSettings";
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
  const debugWorkspaceSettingsStore = useDebugSettingsStore();

  const WATCH_RESOURCE_TOAST_ID = "watch-resource";
  const RESOURCE_TOAST_ID = "resource-toast";
  const PIPELINE_CHECKER_TOAST_ID = "pipeline-checker-toast";

  const pipelineErrors: Ref<CheckResponse[]> = ref([]);
  const pipelineWarns: Ref<CheckResponse[]> = ref([]);
  const pipelineIssueModalOpen = ref(false);

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
      toast.remove(WATCH_RESOURCE_TOAST_ID);
      if (!success) {
        console.error("[Resource] Load failed:", msg);
        toast.add({
          id: RESOURCE_TOAST_ID,
          title: "Resource Load Failed",
          description: msg || "Unknown error",
          icon: "i-lucide-circle-x",
          color: "error",
        });
      } else {
        toast.add({
          id: RESOURCE_TOAST_ID,
          title: "Resource Loaded",
          icon: "i-lucide-check-circle",
          color: "success",
        });
      }
    } else {
      toast.remove(RESOURCE_TOAST_ID);
      toast.add({
        id: WATCH_RESOURCE_TOAST_ID,
        title: "Resource Changed",
        description: `The Resource will reload as ${options.changedPath} changed.`,
        icon: "i-lucide-loader",
        color: "info",
      });
    }

    pipelineErrors.value = [];
    pipelineWarns.value = [];
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

        pipelineErrors.value = grouped.error;
        pipelineWarns.value = grouped.warning;

        const toastType = grouped.error.length > 0 ? "error" : "warning";
        toast.add({
          id: PIPELINE_CHECKER_TOAST_ID,
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
    pipelineErrors.value = grouped.error;
    pipelineWarns.value = grouped.warning;
    pipelineIssueModalOpen.value = true;
  }

  return {
    enabledPaths,
    tryLoadResource,
    onLoadResource: LoadResource,
    pipelineIssueModalOpen,
    pipelineErrors,
    pipelineWarns,
  };
}
