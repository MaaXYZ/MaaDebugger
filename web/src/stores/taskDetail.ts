import { ref } from "vue";
import type { PipelineNodeScope } from "@/views/Index/taskDetail/types";

export const activeTaskIndex = ref(0);
export const followLatestTask = ref(true);
export const selectedNodeId = ref<number | null>(null);
export const viewMode = ref<"live" | "history">("live");
export const livePage = ref(1);
export const historyPage = ref(1);
export const historySnapshotNodes = ref<PipelineNodeScope[]>([]);

export function resetTaskDetailState() {
  activeTaskIndex.value = 0;
  followLatestTask.value = true;
  selectedNodeId.value = null;
  viewMode.value = "live";
  livePage.value = 1;
  historyPage.value = 1;
  historySnapshotNodes.value = [];
}
