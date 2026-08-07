import {
  connectAgent as apiConnect,
  disconnectAgent as apiDisconnect,
  getAgentList,
} from "@/api/http";
import { useI18n } from "vue-i18n";
import { i18n } from "@/i18n";
import { useAgentStore, type AgentItem } from "@/stores/agent";
import useBackoffPolling from "@/api/useBackoffPolling";

interface ConnectResult {
  success: boolean;
  msg?: string;
}

function applyRemoteStatus(
  agent: AgentItem,
  remote?: { status: string; error?: string },
) {
  if (!remote) {
    if (agent.status === "connecting") {
      agent.status = "idle";
      agent.errorMsg = "";
    }
    return;
  }

  switch (remote.status) {
    case "connected":
      agent.status = "connected";
      agent.errorMsg = "";
      break;
    case "connecting":
      agent.status = "connecting";
      agent.errorMsg = "";
      break;
    default:
      agent.status = "failed";
      agent.errorMsg = remote.error || i18n.global.t("agent.connectFailed");
      break;
  }
}

export default function useAgentControl() {
  const { t } = useI18n();
  const toast = useToast();
  const agentStore = useAgentStore();

  // 仅当有 agent 连接时才进行请求
  useBackoffPolling(
    () => syncAgentsFromServer(),
    () => agentStore.getConnectedAgents().length > 0,
    1500,
  );

  async function syncAgentsFromServer() {
    try {
      const remoteAgents = await getAgentList();
      for (const agent of agentStore.agents) {
        const remote = remoteAgents.find(
          (item) => item.identifier === agent.identifier.trim(),
        );
        applyRemoteStatus(agent, remote);
      }
    } catch {
      // ignore sync failure, keep local state
    }
  }

  async function tryConnectAgent(agent: AgentItem): Promise<ConnectResult> {
    const identifier = agent.identifier.trim();
    if (!identifier) return { success: false, msg: i18n.global.t("agent.emptyIdentifier") };

    const result = await apiConnect(identifier);

    if (result.succeed) {
      agent.status = "connected";
      agent.errorMsg = "";
      return { success: true };
    }

    agent.status = "failed";
    agent.errorMsg = result.msg || t("agent.connectFailed");
    return { success: false, msg: agent.errorMsg };
  }

  async function doConnect(agent: AgentItem) {
    agent.status = "connecting";
    agent.errorMsg = "";

    const result = await tryConnectAgent(agent);

    if (result.success) {
      agent.status = "connected";
      agent.errorMsg = "";
    } else {
      agent.status = "failed";
      agent.errorMsg = result.msg || t("agent.connectFailed");
      toast.add({
        id: "agent-toast",
        title: t("agent.connectFailed"),
        description: agent.errorMsg,
        icon: "i-lucide-circle-x",
        color: "error",
      });
      agentStore.cardExpanded = true;
    }

    await syncAgentsFromServer();
  }

  async function doDisconnect(agent: AgentItem) {
    await apiDisconnect(agent.identifier);
    agent.status = "idle";
    agent.errorMsg = "";
    await syncAgentsFromServer();
  }

  async function connectAgents(): Promise<ConnectResult> {
    const enabledAgents = agentStore.getEnabledAgents();

    for (const agent of agentStore.agents) {
      if (!agent.enabled) {
        agent.status = "idle";
        agent.errorMsg = "";
      }
    }

    if (enabledAgents.length === 0) {
      await syncAgentsFromServer();
      return { success: true };
    }

    const failedAgents: string[] = [];

    for (const agent of enabledAgents) {
      agent.status = "connecting";
      agent.errorMsg = "";

      const result = await tryConnectAgent(agent);
      if (!result.success) {
        failedAgents.push(agent.name || agent.identifier || t("agent.unnamed"));
      }
    }

    await syncAgentsFromServer();

    if (failedAgents.length > 0) {
      toast.add({
        id: "agent-toast",
        title: t("agent.connectFailed"),
        description: t("agent.failedList", { list: failedAgents.join(", ") }),
        icon: "i-lucide-circle-x",
        color: "error",
      });
      agentStore.cardExpanded = true;
      return {
        success: false,
        msg: t("agent.failedToConnect", { count: failedAgents.length }),
      };
    }

    return { success: true };
  }

  return {
    doConnect,
    doDisconnect,
    connectAgents,
    syncAgentsFromServer,
  };
}
