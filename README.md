# MaaDebugger

由 [MaaXYZ](https://github.com/MaaXYZ) 团队开发与维护的 MaaFramework 调试器，拥有现代 Web 界面与强大的调试功能。

## 安装方式

### 使用 npm

MaaDebugger 暂时以 `@weinibuliu/maa-debugger` 的名称发布于 [npm](https://www.npmjs.com) 。

```bash
npm install -g @weinibuliu/maa-debugger@latest
```

MaaDebugger 支持命令行参数，你可以通过 `--help` / `-H` 命令来获取帮助。

```bash
maa-debugger --help
# or
maa-dbg --help
```

## 如何开发

在开发阶段需要同时启动**后端**和**前端**。

### 后端 (Go Service)

> [!NOTE]
> `MaaDebugger` 将寻找 `./bin` 下的 MaaFramework 动态库。
>
> 因此在开发阶段，需要手动下载 [MaaFramework](https://github.com/MaaXYZ/MaaFramework/releases) 将其 `/bin` 目录解压至项目根目录。

```bash
air # 支持热重载
```

### 前端 (Web)

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
node build.mjs --skip-go           # 跳过 Go 编译（仅构建前端）
```

构建完成后，启动 `./MaaDebugger` 。程序会从 8011 端口开始自动检测可用端口并打开浏览器。
