#!/usr/bin/env node

import { execSync, spawn } from "node:child_process";
import { existsSync, readdirSync, symlinkSync } from "node:fs";
import { dirname, join, resolve } from "node:path";

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
  platformPackageRoot = dirname(
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
const exePath = join(platformPackageRoot, executableName);
if (!existsSync(exePath)) {
  logError(`[maa-debugger] Missing executable: ${exePath}`);
  logError(
    `[maa-debugger] Expected binary name for current platform: ${executableName}`,
  );
  process.exit(1);
}

const nodePath = resolve(require.resolve("@maaxyz/maa-node"), "../../../");
const channelPath = join(
  nodePath,
  `maa-node-${process.platform}-${process.arch}`,
);

const MAX_PATH_THRESHOLD = 240;
const pendingJunctions = [];

/**
 * If `targetDir` is longer than the threshold, create a directory junction
 * inside the nearest `node_modules` ancestor so the resulting path stays
 * within pnpm's own cache tree (no user-space pollution).
 *
 * Returns the (possibly shortened) directory path.
 */
function ensureShortPath(targetDir, junctionName) {
  if (!isWindows || targetDir.length < MAX_PATH_THRESHOLD) {
    return targetDir;
  }

  // Place the junction right inside the first `node_modules` segment so
  // the alias sits inside the package manager's own cache directory.
  const nmIndex = targetDir.indexOf("node_modules");
  if (nmIndex === -1) {
    return targetDir;
  }
  const junctionBase = targetDir.substring(0, nmIndex + "node_modules".length);
  const junctionPath = join(junctionBase, junctionName);

  try {
    if (!existsSync(junctionPath)) {
      symlinkSync(targetDir, junctionPath, "junction");
    }
    pendingJunctions.push(junctionPath);
    logDebug(
      `[maa-debugger] Created junction for long path (${targetDir.length} chars): ${junctionPath} -> ${targetDir}`,
    );
    return junctionPath;
  } catch (junctionError) {
    logDebug(
      `[maa-debugger] Failed to create junction (falling back to original path): ${junctionError.message}`,
    );
    return targetDir;
  }
}

function cleanupJunctions() {
  for (const jp of pendingJunctions) {
    try {
      // `rmdir` removes only the junction reparse point, NOT the target
      // directory contents.
      execSync(`rmdir "${jp}"`, { stdio: "ignore", shell: true });
    } catch {
      // best-effort cleanup
    }
  }
}

// Register cleanup for all exit paths
process.on("exit", cleanupJunctions);

const shortExeRoot = ensureShortPath(platformPackageRoot, "_maadbg_exe");
const shortChannelPath = ensureShortPath(channelPath, "_maadbg_lib");
const launchPath = join(shortExeRoot, executableName);

const childEnv = {
  ...process.env,
  MAADBG_CHANNEL: "npm",
  MAADBG_CHANNEL_PATH: shortChannelPath,
};

const child = spawn(launchPath, process.argv.slice(2), {
  stdio: "inherit",
  env: childEnv,
  shell: false,
  windowsHide: false,
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

  const exeDir = dirname(launchPath);
  logDebug(`[maa-debugger] Executable path: ${launchPath}`);
  logDebug(`[maa-debugger] Executable path length: ${launchPath.length}`);
  logDebug(`[maa-debugger] Original executable path: ${exePath}`);
  logDebug(`[maa-debugger] Original path length: ${exePath.length}`);
  logDebug(
    `[maa-debugger] Executable exists before launch: ${existsSync(launchPath)}`,
  );
  logDebug(`[maa-debugger] Platform package root: ${platformPackageRoot}`);
  logDebug(`[maa-debugger] Channel path: ${shortChannelPath}`);
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
        `[maa-debugger] Executable not found or cannot be resolved at runtime: ${launchPath}`,
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

// ensure Ctrl+C can clear the junction
process.on("SIGINT", () => {
  child.kill("SIGINT");
  cleanupJunctions();
  process.exit(130);
});

process.on("SIGTERM", () => {
  child.kill("SIGTERM");
  cleanupJunctions();
  process.exit(143);
});
