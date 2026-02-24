import { WebSocket } from "ws";
import type { WSClientMessage } from "@shared/types/ws";
import { broadcaster } from "./broadcaster.js";
import { getStatus } from "../services/maa.js";

/**
 * 处理新的 WebSocket 连接
 */
export function handleWSConnection(ws: WebSocket): void {
  console.log("[WS] Client connected");

  // 注册到广播管理器
  broadcaster.add(ws);

  // 立即发送当前状态
  ws.send(
    JSON.stringify({
      type: "status.update",
      payload: getStatus(),
    }),
  );

  // 处理客户端消息
  ws.on("message", (data) => {
    try {
      const message = JSON.parse(data.toString()) as WSClientMessage;
      handleClientMessage(ws, message);
    } catch (err) {
      console.error("[WS] Failed to parse message:", err);
    }
  });

  ws.on("close", () => {
    console.log("[WS] Client disconnected");
  });

  ws.on("error", (err) => {
    console.error("[WS] Error:", err);
  });
}

/**
 * 处理客户端发来的消息
 */
function handleClientMessage(ws: WebSocket, message: WSClientMessage): void {
  switch (message.type) {
    case "status.subscribe":
      // 客户端订阅状态，发送当前状态
      ws.send(
        JSON.stringify({
          type: "status.update",
          payload: getStatus(),
        }),
      );
      break;

    default:
      console.warn("[WS] Unknown message type:", message.type);
  }
}
