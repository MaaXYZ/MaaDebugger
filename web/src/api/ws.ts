import type { StatusSnapshot } from "@/types/api";
import type { TaskEvent } from "@/components/Index/taskDetail/types";
import type { AgentInfo } from "@/api/http";

export type WSEventHandler = {
  onStatusUpdate?: (status: StatusSnapshot) => void;
  onTaskEvent?: (event: TaskEvent) => void;
  onTaskCompleted?: (result: {
    success: boolean;
    error?: string;
    stopped?: boolean;
    entry?: string;
  }) => void;
  onAgentUpdate?: (agents: AgentInfo[]) => void;
  onScreenshotFrame?: (data: ArrayBuffer) => void;
  onScreenshotError?: (reason: string) => void;
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
  private shouldReconnect = true;
  private url: string = "";

  private reconnectAttempt = 0;
  private readonly reconnectBaseDelayMs = 3000;
  private readonly reconnectMaxDelayMs = 30000;
  private readonly reconnectJitterRatio = 0.2;

  /**
   * 连接到 WebSocket 服务器
   */
  connect(handlers: WSEventHandler = {}): void {
    this.handlers = handlers;
    this.shouldReconnect = true;
    this.resetReconnectState();

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
    this.resetReconnectState();
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
        this.resetReconnectState();
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
    if (event.data instanceof Blob) {
      event.data.arrayBuffer().then((buf) => {
        this.handlers.onScreenshotFrame?.(buf);
      });
      return;
    }
    if (event.data instanceof ArrayBuffer) {
      this.handlers.onScreenshotFrame?.(event.data);
      return;
    }

    // JSON 文本消息
    try {
      const message = JSON.parse(event.data) as Record<string, unknown>;

      switch (message.type) {
        case "status.update":
          this.handlers.onStatusUpdate?.(message.payload as StatusSnapshot);
          break;

        case "task.event":
          this.handlers.onTaskEvent?.(message.payload as TaskEvent);
          break;

        case "task.completed":
          this.handlers.onTaskCompleted?.(
            message.payload as {
              success: boolean;
              error?: string;
              stopped?: boolean;
              entry?: string;
            },
          );
          break;

        case "agent.update":
          this.handlers.onAgentUpdate?.(message.payload as AgentInfo[]);
          break;

        case "screenshot.error": {
          const err = message.payload as { reason: string };
          this.handlers.onScreenshotError?.(err.reason);
          break;
        }

        case "log": {
          const log = message.payload as { level: string; message: string };
          this.handlers.onLog?.(log.level, log.message);
          break;
        }

        default:
          console.warn("[WS] Unknown message type:", message.type);
      }
    } catch (err) {
      console.error("[WS] Failed to parse message:", err);
    }
  }

  private resetReconnectState(): void {
    this.reconnectAttempt = 0;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private computeReconnectDelay(): number {
    const baseDelay = Math.min(
      this.reconnectBaseDelayMs * 2 ** this.reconnectAttempt,
      this.reconnectMaxDelayMs,
    );
    const jitterSpan = baseDelay * this.reconnectJitterRatio;
    const jitter = (Math.random() * 2 - 1) * jitterSpan;
    return Math.max(0, Math.floor(baseDelay + jitter));
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) return;

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    const attempt = this.reconnectAttempt + 1;
    const delay = this.computeReconnectDelay();
    this.reconnectAttempt += 1;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      console.log(
        `[WS] Reconnecting (attempt ${attempt}, delay ${delay}ms)...`,
      );
      this.doConnect();
    }, delay);
  }
}

export const wsClient = new WSClient();
