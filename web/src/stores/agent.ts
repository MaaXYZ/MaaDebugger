import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { ConnectionStatus } from "@/components/Index/agent/types";

export interface AgentItem {
  identifier: string;
  name: string;
  enabled: boolean;
  status: ConnectionStatus;
  errorMsg: string;
}

export const useAgentStore = defineStore(
  "agent",
  () => {
    const agents = ref<AgentItem[]>([]);
    const cardExpanded = ref(false);

    const connectedCount = computed(
      () => agents.value.filter((a) => a.status === "connected").length,
    );
    const hasConnecting = computed(() =>
      agents.value.some((a) => a.status === "connecting"),
    );
    const hasError = computed(() =>
      agents.value.some((a) => a.status === "failed"),
    );

    function addAgent(): AgentItem {
      const item: AgentItem = {
        identifier: "",
        name: "",
        enabled: true,
        status: "idle",
        errorMsg: "",
      };
      agents.value.push(item);
      return item;
    }

    function removeAgent(identifier: string) {
      const idx = agents.value.findIndex((a) => a.identifier === identifier);
      if (idx !== -1) {
        agents.value.splice(idx, 1);
      }
    }

    function removeByIndex(index: number) {
      agents.value.splice(index, 1);
    }

    function getByIdentifier(identifier: string): AgentItem | undefined {
      return agents.value.find((a) => a.identifier === identifier);
    }

    function getConnectedAgents() {
      return agents.value
        .filter((a) => a.status === "connected")
        .map((a) => ({ identifier: a.identifier, name: a.name }));
    }

    function getEnabledAgents() {
      return agents.value.filter((a) => a.enabled);
    }

    function resetRuntimeState() {
      cardExpanded.value = false;

      for (const agent of agents.value) {
        if (typeof agent.enabled !== "boolean") {
          agent.enabled = true;
        }
        agent.status = "idle";
        agent.errorMsg = "";
      }
    }

    /**
     * 持久化恢复后的钩子：重置瞬态运行时字段
     */
    function onRestore() {
      resetRuntimeState();
    }

    return {
      agents,
      cardExpanded,
      connectedCount,
      hasConnecting,
      hasError,
      addAgent,
      removeAgent,
      removeByIndex,
      getByIdentifier,
      getConnectedAgents,
      getEnabledAgents,
      resetRuntimeState,
      onRestore,
    };
  },
  {
    persist: true,
  },
);
