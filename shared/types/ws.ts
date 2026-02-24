/**
 * WebSocket 消息基础结构
 */
export interface WSMessage<T = unknown> {
  type: string;
  id?: string;
  payload: T;
}

/**
 * 状态更新推送
 */
export interface StatusUpdateMessage extends WSMessage {
  type: "status.update";
  payload: {
    controller: "disconnected" | "connecting" | "connected";
    resource: "unloaded" | "loading" | "loaded";
    task: "idle" | "running" | "success" | "failed";
    agent: "disconnected" | "connecting" | "connected";
  };
}

/**
 * 日志推送
 */
export interface LogMessage extends WSMessage {
  type: "log";
  payload: {
    level: "info" | "warn" | "error";
    message: string;
    timestamp: number;
  };
}

/**
 * 客户端 -> 服务端：订阅状态
 */
export interface StatusSubscribeMessage extends WSMessage {
  type: "status.subscribe";
  payload: Record<string, never>;
}

/**
 * 所有 WebSocket 消息类型联合
 */
export type WSServerMessage = StatusUpdateMessage | LogMessage;
export type WSClientMessage = StatusSubscribeMessage;
