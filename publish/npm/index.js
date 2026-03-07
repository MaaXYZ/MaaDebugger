#!/usr/bin/env node

const { spawn } = require("node:child_process");
const { existsSync } = require("node:fs");
const path = require("node:path");

const platformPackageName = `@weinibuliu/maa-debugger-${process.platform}-${process.arch}`;

let platformPackageRoot;
try {
  platformPackageRoot = path.dirname(
    require.resolve(`${platformPackageName}/package.json`),
  );
} catch (error) {
  console.error(
    `[maa-debugger] Missing platform package: ${platformPackageName}`,
  );
  console.error(
    `[maa-debugger] Reinstall the package on ${process.platform}-${process.arch} or choose a supported platform.`,
  );
  process.exit(1);
}

const isWindows = process.platform === "win32";
const executableName = `MaaDebugger${isWindows ? ".exe" : ""}`;
const exePath = path.join(platformPackageRoot, executableName);
if (!existsSync(exePath)) {
  console.error(`[maa-debugger] Missing executable: ${exePath}`);
  console.error(
    `[maa-debugger] Expected binary name for current platform: ${executableName}`,
  );
  process.exit(1);
}

const nodePath = path.resolve(require.resolve("@maaxyz/maa-node"), "../../../");
const channelPath = path.join(
  nodePath,
  `maa-node-${process.platform}-${process.arch}`,
);

const child = spawn(exePath, process.argv.slice(2), {
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
