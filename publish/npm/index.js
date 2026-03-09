#!/usr/bin/env node

const { execFile, spawn } = require("node:child_process");
const { existsSync, readdirSync } = require("node:fs");
const path = require("node:path");

const platformPackageName = `@weinibuliu/maa-debugger-${process.platform}-${process.arch}`;

const supportsColor = (() => {
  if (process.env.NO_COLOR) {
    return false;
  }
  if (process.env.FORCE_COLOR === "0") {
    return false;
  }
  if (process.env.FORCE_COLOR) {
    return true;
  }
  return Boolean(process.stderr?.isTTY);
})();

const ansi = {
  red: ["\u001B[31m", "\u001B[39m"],
  cyan: ["\u001B[36m", "\u001B[39m"],
};

function colorize(name, text) {
  if (!supportsColor) {
    return text;
  }
  const color = ansi[name];
  return color ? `${color[0]}${text}${color[1]}` : text;
}

function logError(message, ...rest) {
  console.error(colorize("red", message), ...rest);
}

function logDebug(message, ...rest) {
  console.debug(colorize("cyan", message), ...rest);
}

let platformPackageRoot;
try {
  platformPackageRoot = path.dirname(
    require.resolve(`${platformPackageName}/package.json`),
  );
} catch (error) {
  logError(`[maa-debugger] Missing platform package: ${platformPackageName}`);
  logError(
    `[maa-debugger] Reinstall the package on ${process.platform}-${process.arch} or choose a supported platform.`,
  );
  process.exit(1);
}

const isWindows = process.platform === "win32";
const executableName = `MaaDebugger${isWindows ? ".exe" : ""}`;
const exePath = path.join(platformPackageRoot, executableName);
if (!existsSync(exePath)) {
  logError(`[maa-debugger] Missing executable: ${exePath}`);
  logError(
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
  logError(`[maa-debugger] Failed to start ${executableName}:`, err);

  const exeDir = path.dirname(exePath);
  logDebug(`[maa-debugger] Executable path: ${exePath}`);
  logDebug(`[maa-debugger] Executable path length: ${exePath.length}`);
  logDebug(`[maa-debugger] Executable exists before launch: ${existsSync(exePath)}`);
  logDebug(`[maa-debugger] Platform package root: ${platformPackageRoot}`);
  logDebug(`[maa-debugger] Channel path: ${channelPath}`);
  try {
    logDebug(
      `[maa-debugger] Executable directory entries: ${readdirSync(exeDir).join(", ") || "(empty)"}`,
    );
  } catch (readDirError) {
    logError(
      `[maa-debugger] Failed to inspect executable directory: ${readDirError}`,
    );
  }

  switch (err.code) {
    case "EACCES":
      logError(
        `[maa-debugger] Permission denied when starting ${executableName}. The executable bit may be missing from the installed platform package.`,
      );
      logError(
        `[maa-debugger] Try restoring execute permission manually, for example: chmod 755 "${exePath}"`,
      );
      break;
    case "ENOENT": {
      logError(
        `[maa-debugger] Executable not found or cannot be resolved at runtime: ${exePath}`,
      );
      if (isWindows) {
        logError(
          `[maa-debugger] On Windows, this can also be caused by CreateProcess failing on a deeply nested pnpm dlx cache path.`,
        );
      }
      break;
    }
  }

  process.exit(1);
});
