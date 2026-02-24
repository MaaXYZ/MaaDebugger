import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type {
  StatusSnapshot,
  ControllerStatus,
  ResourceStatus,
  TaskStatus,
  AgentStatus,
} from "@shared/types/api";

/**
 * 全局 MaaFW 状态 Store
 *
 * 通过 WebSocket 接收后端广播的状态更新，供各组件使用。
 */
export const useStatusStore = defineStore("status", () => {
  const status = ref<StatusSnapshot>({
    controller: "disconnected",
    resource: "unloaded",
    task: "idle",
    agent: "disconnected",
  });

  // --- Getters ---
  const controllerStatus = computed<ControllerStatus>(
    () => status.value.controller,
  );
  const resourceStatus = computed<ResourceStatus>(
    () => status.value.resource,
  );
  const taskStatus = computed<TaskStatus>(() => status.value.task);
  const agentStatus = computed<AgentStatus>(() => status.value.agent);

  // --- Actions ---
  function updateStatus(snapshot: StatusSnapshot) {
    status.value = snapshot;
  }

  return {
    status,
    controllerStatus,
    resourceStatus,
    taskStatus,
    agentStatus,
    updateStatus,
  };
});
