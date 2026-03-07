#!/usr/bin/env node

const { spawn } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");
const { exit } = require("node:process");

// 获取可执行文件路径
const rootDir = path.resolve(__dirname);
const isWindows = process.platform === "win32";
const executableName = `MaaDebugger${isWindows ? ".exe" : ""}`;
const exePath = path.join(rootDir, executableName);
if (!existsSync(exePath)) {
  console.error(`[maa-debugger] Missing executable: ${exePath}`);
  console.error(
    `[maa-debugger] Expected binary name for current platform: ${executableName}`,
  );
  process.exit(1);
}

// 获取 MaaFramework 动态库路径
const nodePath = path.resolve(require.resolve("@maaxyz/maa-node"), "../../../");
const channelPath = path.join(
  nodePath,
  `maa-node-${process.platform}-${process.arch}`,
);

// 启动进程
const child = spawn(exePath, process.argv.slice(2), {
  // cwd: rootDir,
  stdio: "inherit",
  env: {
    ...process.env,
    MAADBG_CHANNEL: "npm",
    MAADBG_CHANNEL_PATH: channelPath,
  },
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 0);
});

child.on("error", (err) => {
  console.error(`[maa-debugger] Failed to start ${executableName}:`, err);
  process.exit(1);
});
