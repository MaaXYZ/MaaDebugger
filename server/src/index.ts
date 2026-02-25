import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { cors } from "hono/cors";
import { logger } from "hono/logger";
import { WebSocketServer } from "ws";
import type { IncomingMessage } from "node:http";
import type { Duplex } from "node:stream";

import { controllerRoutes } from "./routes/controller.js";
import { handleWSConnection } from "./ws/handler.js";
import {
  initMaa,
  getStatus,
  getVersion,
  loadResource,
} from "./services/maa.js";
import {
  getAllConfig,
  setConfigs,
  getConfig,
  setConfig,
} from "./services/config.js";

const app = new Hono();

// --- 中间件 ---
app.use("*", logger());
app.use(
  "*",
  cors({
    origin: "*", // 开发环境允许所有来源
  }),
);

// --- 状态信息路由 ---
app.get("/api/info/version", (c) => {
  return c.json({ success: true, data: getVersion() });
});

app.get("/api/info/status", (c) => {
  return c.json({ success: true, data: getStatus() });
});

// --- Controller 路由 ---
app.route("/api/controller", controllerRoutes);

// --- Resource 路由 ---
app.post("/api/resource/load", async (c) => {
  const body = await c.req.json();
  const paths: string[] = body.paths ?? [];

  if (paths.length === 0) {
    return c.json({ success: false, error: "No resource paths provided" }, 400);
  }

  const result = await loadResource(paths);
  return c.json(
    { success: result.success, error: result.error },
    result.success ? 200 : 400,
  );
});

// --- Config 持久化路由 ---
app.get("/api/config", (c) => {
  return c.json({ success: true, data: getAllConfig() });
});

app.get("/api/config/:key", (c) => {
  const key = c.req.param("key");
  return c.json({ success: true, data: getConfig(key) });
});

app.put("/api/config/:key", async (c) => {
  const key = c.req.param("key");
  const body = await c.req.json();
  setConfig(key, body);
  return c.json({ success: true });
});

app.put("/api/config", async (c) => {
  const body = await c.req.json();
  setConfigs(body as Record<string, unknown>);
  return c.json({ success: true });
});

// --- 根路由 ---
app.get("/", (c) => {
  return c.json({
    name: "maa-debugger-server",
    version: "0.0.1",
  });
});

// --- 启动服务器 ---
const PORT = parseInt(process.env.PORT ?? "8011", 10);
const HOST = process.env.HOST ?? "127.0.0.1";

// 尝试初始化 MaaFW
const maaAvailable = initMaa();

console.log(`
╔══════════════════════════════════════════╗
║         MaaDebugger Server v0.0.1        ║
╠══════════════════════════════════════════╣
║  HTTP:  http://${HOST}:${PORT}            ║
║  WS:    ws://${HOST}:${PORT}/ws           ║
║  MaaFW: ${maaAvailable ? "✅ Available" : "❌ Not available"}                 ║
╚══════════════════════════════════════════╝
`);

const server = serve({
  fetch: app.fetch,
  hostname: HOST,
  port: PORT,
});

// 处理端口占用等启动错误
(server as any).on("error", (err: NodeJS.ErrnoException) => {
  if (err.code === "EADDRINUSE") {
    console.error(
      `\n❌ Port ${PORT} is already in use. Please stop the other process or use a different port:\n` +
        `   PORT=3001 pnpm dev\n`,
    );
  } else {
    console.error("[Server] Error:", err);
  }
  process.exit(1);
});

// --- WebSocket 服务器 ---
const wss = new WebSocketServer({ noServer: true });

wss.on("connection", (ws) => {
  handleWSConnection(ws);
});

// 将 HTTP upgrade 请求路由到 WebSocket
(server as any).on(
  "upgrade",
  (request: IncomingMessage, socket: Duplex, head: Buffer) => {
    const url = new URL(request.url ?? "/", `http://${request.headers.host}`);

    if (url.pathname === "/ws") {
      wss.handleUpgrade(request, socket, head, (ws) => {
        wss.emit("connection", ws, request);
      });
    } else {
      socket.destroy();
    }
  },
);

console.log(`[Server] Listening on ${HOST}:${PORT}`);
