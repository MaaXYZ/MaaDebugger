# MaaDebugger

MaaFramework 的可视化调试工具。

## 启动项目

需要同时启动**后端**和**前端**两个终端：

### Terminal 1: 后端 (server)

```bash
cd server
pnpm install    # 首次需要安装依赖
pnpm dev        # 启动后端 (http://127.0.0.1:3000)
```

### Terminal 2: 前端 (web)

```bash
cd web
pnpm install    # 首次需要安装依赖
pnpm dev        # 启动前端 (http://localhost:5173)
```

然后打开浏览器访问 **<http://localhost:5173>** 即可。

前端会自动将 `/api` 和 `/ws` 请求代理到后端 `http://127.0.0.1:3000`。

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + Nuxt UI 4 + Pinia + Vite 7 + TypeScript |
| 后端 | Hono + ws + Node.js + TypeScript |
| MaaFW | @maaxyz/maa-node (native addon) |
| 通信 | REST API + WebSocket |
