# MaaDebugger

MaaFramework 的可视化调试工具。

## 开发模式

需要同时启动**后端**和**前端**两个终端：

### Terminal 1: 后端 (Go Service)

```bash
cd server
go run ./cmd/server              # 启动后端，自动选择可用端口（默认从 8011 开始）
go run ./cmd/server --port 9090  # 指定端口
```

### Terminal 2: 前端 (Web)

```bash
cd web
pnpm install    # 首次需要安装依赖
pnpm dev        # 启动前端 (http://localhost:5173)
```

然后打开浏览器访问 **<http://localhost:5173>** 即可。

前端会自动将 `/api` 和 `/ws` 请求代理到后端 `http://127.0.0.1:8011`。

## 生产构建

一键构建前端并嵌入到 Go 二进制文件中，最终产物为单个可执行文件：

```bash
node build.mjs                     # 构建当前平台
node build.mjs --os linux          # 交叉编译 Linux
node build.mjs --os windows        # 交叉编译 Windows
node build.mjs --os darwin         # 交叉编译 macOS
node build.mjs --skip-frontend     # 跳过前端构建（仅编译 Go）
```

构建完成后，启动 `./MaaDebugger`（Windows 下为 `MaaDebugger.exe`），程序会自动选择可用端口（默认从 8011 开始）并打开浏览器。也可通过 `--port` 指定端口：

```bash
./MaaDebugger --port 9090
```
> [!NOTE]
> 需要注意的是 `MaaDebugger` 将寻找 `./bin` 下的 MaaFramework 动态库。当前阶段请先手动下载并解压，后续将优化这一问题。

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + Nuxt UI 4 + Pinia + Vite 7 + TypeScript |
| 后端 | Go + net/http + gorilla/websocket + zerolog |
| MaaFW | maa-framework-go (CGO binding) |
| 通信 | REST API + WebSocket |
| 嵌入 | Go `embed` (前端编译产物嵌入到 Go 二进制) |
