import {
  connectAgent as apiConnect,
  disconnectAgent as apiDisconnect,
} from "@/api/http";
import { type AgentItem } from "@/stores/agent";

export default function useAgentControl() {
  async function tryConnectAgent(
    agent: AgentItem,
  ): Promise<{ success: boolean; msg?: string }> {
    const identifier = agent.identifier.trim();
    if (!identifier) return { success: false, msg: "Empty identifier" };

    const result = await apiConnect(identifier);

    if (result.succeed) {
      agent.status = "connected";
      agent.errorMsg = "";
      return { success: true };
    } else {
      agent.status = "failed";
      agent.errorMsg = result.msg || "Connection failed";
      return { success: false, msg: agent.errorMsg };
    }
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
      agent.errorMsg = result.msg || "Connection failed";
    }
  }

  async function doDisconnect(agent: AgentItem) {
    await apiDisconnect(agent.identifier);
    agent.status = "idle";
    agent.errorMsg = "";
  }

  async function connectAgents() {}

  return {
    doConnect,
    doDisconnect,
    connectAgents,
  };
}
