import { computed, ref } from "vue";
import type { Ref } from "vue";
import { useResourceStore } from "@/stores/resource";
import { useSignalStore } from "@/stores/signal";
import { useStatusStore } from "@/stores/status";
import { useDebugSettingsStore } from "@/stores/debugSettings";
import {
  DoPipelineCheck,
  getPipelineCheckResult,
  loadResource,
} from "@/api/http";
import type { CheckResponse } from "@/types/pipeline";
import { useRouter } from "vue-router";

interface LoadResourceOptions {
  manual: boolean;
  changedPath: string;
}

interface PipelineIssueGrouped {
  error: CheckResponse[];
  warning: CheckResponse[];
}

interface PipelinePrecheckOptions {
  toastId: string;
  blockedTitle: string;
  notifyTitle?: string;
  notify?: boolean;
}

export default function useResourceControl() {
  const router = useRouter();
  const toast = useToast();
  const resourceStore = useResourceStore();
  const enabledPaths = computed(() => resourceStore.getEnabledPaths());
  const signalStore = useSignalStore();
  const statusStore = useStatusStore();
  const debugWorkspaceSettingsStore = useDebugSettingsStore();

  const WATCH_RESOURCE_TOAST_ID = "watch-resource";
  const RESOURCE_TOAST_ID = "resource-toast";

  const pipelineErrors: Ref<CheckResponse[]> = ref([]);
  const pipelineWarns: Ref<CheckResponse[]> = ref([]);
  const pipelineIssueModalOpen = ref(false);

  async function precheckPipelineIssuesBeforeAction(
    options: PipelinePrecheckOptions,
  ): Promise<{ blocked: boolean; grouped: PipelineIssueGrouped }> {
    const emptyGrouped: PipelineIssueGrouped = { error: [], warning: [] };
    if (!debugWorkspaceSettingsStore.checkPipeline) {
      return { blocked: false, grouped: emptyGrouped };
    }

    let pipelineCheckResult: CheckResponse[] = [];
    const checkResponse = await DoPipelineCheck();
    if (checkResponse.succeed && Array.isArray(checkResponse.data)) {
      pipelineCheckResult = checkResponse.data;
    } else {
      toast.add({
        id: options.toastId,
        title: "Pipeline Check Failed",
        description: checkResponse.msg,
        icon: "i-lucide-circle-x",
        color: "error",
      });
      console.error(
        "[PipelineChecker] Check once failed, fallback to cached result:",
        checkResponse.msg,
      );
      pipelineCheckResult = await getPipelineCheckResult();
    }

    const grouped = pipelineCheckResult.reduce(
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

    const errorCount = grouped.error.length;
    const warningCount = grouped.warning.length;
    const issueCount = errorCount + warningCount;

    if (debugWorkspaceSettingsStore.preventResourceLoaded && issueCount > 0) {
      toast.add({
        id: options.toastId,
        title: options.blockedTitle,
        description: `Pipeline has ${errorCount} error(s) and ${warningCount} warning(s). Fix issues or disable Prevent Resource Loaded to continue.`,
        icon: "i-lucide-octagon-x",
        color: "error",
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
          {
            icon: "i-lucide-settings",
            label: "Go to Settings",
            color: "neutral",
            variant: "outline",
            onClick: () => {
              router.push("/settings#debug-preventResourceLoaded");
            },
          },
        ],
      });
      return { blocked: true, grouped };
    }

    if (options.notify ?? true) {
      const notifyLevel = debugWorkspaceSettingsStore.checkPipelineNotifyLevel;
      const shouldNotify =
        notifyLevel === "WARNING"
          ? issueCount > 0
          : notifyLevel === "ERROR"
            ? errorCount > 0
            : false;

      if (shouldNotify) {
        toast.add({
          id: options.toastId,
          title: options.notifyTitle || "Pipeline Issues Found",
          description: `Found ${errorCount} error(s) and ${warningCount} warning(s).`,
          icon: "i-lucide-alert-triangle",
          color: errorCount > 0 ? "error" : "warning",
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

    return { blocked: false, grouped };
  }

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
    const precheck = await precheckPipelineIssuesBeforeAction({
      toastId: options.manual ? RESOURCE_TOAST_ID : WATCH_RESOURCE_TOAST_ID,
      blockedTitle: options.manual
        ? "Resource Load Blocked"
        : "Resource Reload Blocked",
      notifyTitle: "Pipeline Issues Found",
    });
    if (precheck.blocked) {
      statusStore.setResourceStatus("failed");
      return;
    }

    const { success, msg } = await tryLoadResource();

    if (!success) {
      statusStore.setResourceStatus("failed");
    } else {
      statusStore.setResourceStatus("loaded");
    }

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
  }

  function openPipelineCheckDetails(grouped?: {
    error: CheckResponse[];
    warning: CheckResponse[];
  }) {
    if (!grouped) return;

    pipelineErrors.value = grouped.error;
    pipelineWarns.value = grouped.warning;
    pipelineIssueModalOpen.value = true;
  }

  return {
    enabledPaths,
    tryLoadResource,
    precheckPipelineIssuesBeforeAction,
    onLoadResource: LoadResource,
    openPipelineCheckDetails,
    pipelineIssueModalOpen,
    pipelineErrors,
    pipelineWarns,
  };
}
