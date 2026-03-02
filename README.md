# MaaDebugger

由 MaaXYZ 团队开发与维护的 MaaFramework 调试器，拥有现代的 WebUI 和即时的任务反馈。

Official desktop debugger for MaaFramework, featuring a web-based UI and real-time task inspection.

## 安装方式

### 使用 npm (Node.js)

MaaDebugger 以 `maa-debugger` 的名称发布于 [npm](https://www.npmjs.com) 。你可以使用 npm 或其他包管理器(如 pnpm)安装、管理与使用。这里仅介绍 npm 用法。

```bash
# 全局安装
npm i -g maa-debugger
maa-debugger
# 非全局安装
npx maa-debugger
```

MaaDebugger 支持命令行参数，你可以通过 `--help` / `-H` 命令来获取帮助。

```bash
npx maa-debugger --help
```

### 使用 pip (Python)

MaaDebugger
将以 `MaaDebugger` 的名称发布于 [PyPI](https://pypi.org/project/MaaDebugger) 。你可以使用 pip 或其他包管理器(如 uv)安装、管理与使用。
>[!WARNING]
在早期开发阶段，我们**不会**发布至 PyPI，请自行使用其他渠道。

### 自行下载

我们也提供了 [Github Release](https://github.com/MaaXYZ/MaaDebugger/releases) 下载渠道，下载解压后即可使用。

## 开发模式

在开发阶段需要同时启动**后端**和**前端**两个终端：

### Terminal 1: 后端 (Go Service)

```bash
cd server
air # 支持热重载
```

### Terminal 2: 前端 (Web)

```bash
cd web
pnpm i   # 首次需要安装依赖
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

构建完成后，启动 `./MaaDebugger`（Windows 下为 `MaaDebugger.exe`），程序会自动选择可用端口（默认从 8011 开始）并打开浏览器。

> [!NOTE]
> 需要注意的是 `MaaDebugger` 将寻找 `./bin` 下的 MaaFramework 动态库。

## 环境变量

TODO

## 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + Nuxt UI 4 + Pinia + Vite 7 + TypeScript |
| 后端 | Go + net/http + gorilla/websocket + zerolog |
| MaaFW | maa-framework-go (CGO binding) |
| 通信 | REST API + WebSocket |
| 嵌入 | Go `embed` (前端编译产物嵌入到 Go 二进制) |
