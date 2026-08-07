import { ref } from "vue";
import type { Ref } from "vue";
import { defineStore } from "pinia";
import { getStoreConfig, StartPipelineCheck, StopPipelineCheck } from "@/api/http";

export type PipelineNotifyLevel = "NULL" | "ERROR" | "WARNING";

/**
 * 设置侧边栏的自动折叠模式（合并自原来的两个布尔开关，消除语义歧义）：
 * - "never"：从不自动折叠，侧边栏仅由手动切换控制；
 * - "task-running"：任务开始运行时自动折叠（折叠后保持，直到手动展开）；
 * - "always"：始终折叠，直到手动展开。
 */
export type SidebarAutoCollapseMode = "never" | "task-running" | "always";

export const useDebugSettingsStore = defineStore(
  "debugWorkspaceSettings",
  () => {
    /** 自动折叠模式（策略设置，唯一语义来源） */
    const sidebarAutoCollapse = ref<SidebarAutoCollapseMode>("task-running");
    /** 当前是否折叠（实时位置状态：手动切换或任务触发都会写入并持久化） */
    const leftTabsCollapsed = ref(false);
    const watchResourceChange = ref(true);
    const watchResourceChangeInterval = ref(1000);
    const showTaskFps = ref(false);
    const checkPipeline = ref(true);
    const checkPipelineNotifyLevel: Ref<PipelineNotifyLevel> = ref("ERROR");
    const preventResourceLoaded = ref(true);

    function setSidebarAutoCollapse(mode: SidebarAutoCollapseMode) {
      sidebarAutoCollapse.value = mode;
      // 模式切换时同步当前位置，避免模式与状态互相矛盾：
      // "always" 立即折叠；"never" / "task-running" 展开。
      if (mode === "always") {
        leftTabsCollapsed.value = true;
      } else {
        leftTabsCollapsed.value = false;
      }
    }

    function setLeftTabsCollapsed(value: boolean) {
      leftTabsCollapsed.value = value;
    }

    /**
     * 旧配置迁移：早期版本用两个布尔开关（autoCollapseLeftTabsOnRunStart + leftTabsCollapsed）
     * 表达折叠策略，合并为单一模式后映射到新模式，避免老用户行为被静默改变。
     */
    function onRestore() {
      void (async () => {
        try {
          const saved = await getStoreConfig<{
            sidebarAutoCollapse?: SidebarAutoCollapseMode;
            autoCollapseLeftTabsOnRunStart?: boolean;
            leftTabsCollapsed?: boolean;
          }>("debugWorkspaceSettings");
          if (!saved) return;
          // 已是新模式或旧字段不存在时无需迁移
          if (saved.sidebarAutoCollapse || typeof saved.autoCollapseLeftTabsOnRunStart !== "boolean") {
            return;
          }
          if (saved.leftTabsCollapsed) {
            setSidebarAutoCollapse("always");
          } else if (saved.autoCollapseLeftTabsOnRunStart) {
            setSidebarAutoCollapse("task-running");
          } else {
            setSidebarAutoCollapse("never");
          }
        } catch {
          // 迁移失败时保持默认模式
        }
      })();
    }

    function setWatchResourceChange(value: boolean) {
      watchResourceChange.value = value;
    }

    function setWatchResourceChangeInterval(value: number) {
      watchResourceChangeInterval.value = value;
    }

    function setShowTaskFps(value: boolean) {
      showTaskFps.value = value;
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
      sidebarAutoCollapse.value = "task-running";
      leftTabsCollapsed.value = false;
      watchResourceChange.value = true;
      watchResourceChangeInterval.value = 1000;
      showTaskFps.value = false;
      checkPipeline.value = true;
      checkPipelineNotifyLevel.value = "ERROR";
      preventResourceLoaded.value = true;
      void syncPipelineChecker(true);
    }

    return {
      sidebarAutoCollapse,
      leftTabsCollapsed,
      watchResourceChange,
      watchResourceChangeInterval,
      showTaskFps,
      checkPipeline,
      checkPipelineNotifyLevel,
      preventResourceLoaded,
      setSidebarAutoCollapse,
      setLeftTabsCollapsed,
      onRestore,
      setWatchResourceChange,
      setWatchResourceChangeInterval,
      setShowTaskFps,
      setCheckPipeline,
      syncPipelineChecker,
      setCheckPipelineNotifyLevel,
      setPreventResourceLoaded,
      reset,
    };
  },
  { persist: true },
);
