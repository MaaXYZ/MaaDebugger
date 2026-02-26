import { ref } from "vue";
import type { AgentInfo } from "@/api/http";

/**
 * Reactive ref that holds the latest agent.update payload from WebSocket.
 * AgentCard watches this to sync server-side state.
 */
export const latestAgentUpdate = ref<AgentInfo[]>([]);
