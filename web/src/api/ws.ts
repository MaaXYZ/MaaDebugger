import type { WSServerMessage } from "@shared/types/ws";
import type { StatusSnapshot } from "@shared/types/api";

export type WSEventHandler = {
  onStatusUpdate?: (status: StatusSnapshot) => void;
  onLog?: (level: string, message: string) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
};

/**
 * WebSocket 客户端管理器
 *
 * 自动重连、消息类型分发
 */
class WSClient {
  private ws: WebSocket | null = null;
  private handlers: WSEventHandler = {};
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = 3000;
  private shouldReconnect = true;
  private url: string = "";

  /**
   * 连接到 WebSocket 服务器
   */
  connect(handlers: WSEventHandler = {}): void {
    this.handlers = handlers;
    this.shouldReconnect = true;

    // 根据当前页面 URL 构建 WebSocket URL
    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    this.url = `${protocol}//${location.host}/ws`;

    this.doConnect();
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * 获取连接状态
   */
  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private doConnect(): void {
    if (this.ws) {
      this.ws.close();
    }

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log("[WS] Connected");
        this.handlers.onOpen?.();
      };

      this.ws.onclose = () => {
        console.log("[WS] Disconnected");
        this.handlers.onClose?.();
        this.scheduleReconnect();
      };

      this.ws.onerror = (event) => {
        console.error("[WS] Error:", event);
        this.handlers.onError?.(event);
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event);
      };
    } catch (err) {
      console.error("[WS] Connection failed:", err);
      this.scheduleReconnect();
    }
  }

  private handleMessage(event: MessageEvent): void {
    // 二进制数据（预留给截图帧流）
    if (event.data instanceof Blob || event.data instanceof ArrayBuffer) {
      // TODO: Phase 2 截图帧处理
      return;
    }

    // JSON 文本消息
    try {
      const message = JSON.parse(event.data) as WSServerMessage;

      switch (message.type) {
        case "status.update":
          this.handlers.onStatusUpdate?.(message.payload as StatusSnapshot);
          break;

        case "log":
          this.handlers.onLog?.(
            (message.payload as any).level,
            (message.payload as any).message,
          );
          break;

        default:
          console.warn("[WS] Unknown message type:", (message as any).type);
      }
    } catch (err) {
      console.error("[WS] Failed to parse message:", err);
    }
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) return;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    this.reconnectTimer = setTimeout(() => {
      console.log("[WS] Reconnecting...");
      this.doConnect();
    }, this.reconnectDelay);
  }
}

export const wsClient = new WSClient();
