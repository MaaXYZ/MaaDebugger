#!/usr/bin/env node

const { execFile, spawn } = require("node:child_process");
const { existsSync, readdirSync } = require("node:fs");
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

const childEnv = {
  ...process.env,
  MAADBG_CHANNEL: "npm",
  MAADBG_CHANNEL_PATH: channelPath,
};

const child = isWindows
  ? execFile(exePath, process.argv.slice(2), {
      stdio: "inherit",
      env: childEnv,
      windowsHide: false,
    })
  : spawn(exePath, process.argv.slice(2), {
      stdio: "inherit",
      env: childEnv,
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

  const exeDir = path.dirname(exePath);
  console.debug(`[maa-debugger] Executable path: ${exePath}`);
  console.debug(`[maa-debugger] Executable path length: ${exePath.length}`);
  console.debug(
    `[maa-debugger] Executable exists before launch: ${existsSync(exePath)}`,
  );
  console.debug(`[maa-debugger] Platform package root: ${platformPackageRoot}`);
  console.debug(`[maa-debugger] Channel path: ${channelPath}`);
  try {
    console.debug(
      `[maa-debugger] Executable directory entries: ${readdirSync(exeDir).join(", ") || "(empty)"}`,
    );
  } catch (readDirError) {
    console.error(
      `[maa-debugger] Failed to inspect executable directory: ${readDirError}`,
    );
  }

  switch (err.code) {
    case "EACCES":
      console.error(
        `[maa-debugger] Permission denied when starting ${executableName}. The executable bit may be missing from the installed platform package.`,
      );
      console.error(
        `[maa-debugger] Try restoring execute permission manually, for example: chmod 755 "${exePath}"`,
      );
      break;
    case "ENOENT": {
      console.error(
        `[maa-debugger] Executable not found or cannot be resolved at runtime: ${exePath}`,
      );
      if (isWindows) {
        console.error(
          `[maa-debugger] On Windows, this can also be caused by CreateProcess failing on a deeply nested pnpm dlx cache path.`,
        );
      }
      break;
    }
  }

  process.exit(1);
});
