import { WebSocket } from "ws";
import type { WSServerMessage } from "@shared/types/ws";

/**
 * WebSocket 广播管理器
 * 管理所有连接的 WebSocket 客户端，提供统一的广播能力
 */
class Broadcaster {
  private clients: Set<WebSocket> = new Set();

  /**
   * 添加一个 WebSocket 客户端
   */
  add(ws: WebSocket): void {
    this.clients.add(ws);
    ws.on("close", () => {
      this.clients.delete(ws);
    });
    ws.on("error", () => {
      this.clients.delete(ws);
    });
  }

  /**
   * 向所有连接的客户端广播 JSON 消息
   */
  broadcast(message: WSServerMessage): void {
    const data = JSON.stringify(message);
    for (const client of this.clients) {
      if (client.readyState === WebSocket.OPEN) {
        client.send(data);
      }
    }
  }

  /**
   * 向所有连接的客户端广播二进制数据
   */
  broadcastBinary(data: Buffer | Uint8Array): void {
    for (const client of this.clients) {
      if (client.readyState === WebSocket.OPEN) {
        client.send(data);
      }
    }
  }

  /**
   * 获取当前连接的客户端数量
   */
  get size(): number {
    return this.clients.size;
  }
}

export const broadcaster = new Broadcaster();
