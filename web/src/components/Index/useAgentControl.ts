import {
  connectAgent as apiConnect,
  disconnectAgent as apiDisconnect,
} from "@/api/http";
import { type AgentItem } from "@/stores/agent";

export async function doConnect(agent: AgentItem) {
  const identifier = agent.identifier.trim();
  if (!identifier) return;

  agent.status = "connecting";
  agent.errorMsg = "";

  try {
    const result = await apiConnect(identifier);

    if (result.succeed) {
      agent.status = "connected";
      agent.errorMsg = "";
    } else {
      agent.status = "failed";
      agent.errorMsg = result.msg || "Connection failed";
    }
  } catch (error: unknown) {
    agent.status = "failed";
    agent.errorMsg = error instanceof Error ? error.message : "Unknown error";
  }
}

export async function doDisconnect(agent: AgentItem) {
  await apiDisconnect(agent.identifier);
  agent.status = "idle";
  agent.errorMsg = "";
}
