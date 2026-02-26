#!/usr/bin/env node

/**
 * Build script: compile Vue frontend → embed into Go binary
 *
 * Usage:
 *   node scripts/build.mjs              # build for current OS
 *   node scripts/build.mjs --os linux   # cross-compile for linux
 *   node scripts/build.mjs --os windows # cross-compile for windows
 *   node scripts/build.mjs --os darwin  # cross-compile for macOS
 *   node scripts/build.mjs --skip-frontend  # skip frontend build
 */

import { execSync } from "node:child_process";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = __dirname;
const WEB_DIR = path.join(ROOT, "web");
const GO_DIR = path.join(ROOT, "server");

// Parse args
const args = process.argv.slice(2);
const skipFrontend = args.includes("--skip-frontend");
const osIndex = args.indexOf("--os");
const targetOS = osIndex !== -1 ? args[osIndex + 1] : undefined;

function run(cmd, cwd, env) {
  console.log(`\n> ${cmd}`);
  execSync(cmd, {
    cwd,
    stdio: "inherit",
    env: { ...process.env, ...env },
  });
}

function step(n, total, msg) {
  console.log(
    `\n[${"=".repeat(n)}${" ".repeat(total - n)}] (${n}/${total}) ${msg}`,
  );
}

const totalSteps = skipFrontend ? 1 : 3;
let currentStep = 0;

// ── Step 1: Install frontend dependencies ──
if (!skipFrontend) {
  step(++currentStep, totalSteps, "Installing frontend dependencies...");
  run("pnpm install", WEB_DIR);

  // ── Step 2: Build frontend ──
  step(++currentStep, totalSteps, "Building frontend (vite build)...");
  run("pnpm run build", WEB_DIR);

  // Verify output
  const distIndex = path.join(GO_DIR, "frontend", "dist", "index.html");
  if (!existsSync(distIndex)) {
    console.error("\n❌ Frontend build output not found at:", distIndex);
    process.exit(1);
  }
  console.log("✅ Frontend build output verified.");
}

// ── Step 3: Build Go binary ──
step(++currentStep, totalSteps, "Building Go binary...");

const ext =
  (targetOS ?? process.platform) === "win32" || targetOS === "windows"
    ? ".exe"
    : "";
const outputName = `MaaDebugger${ext}`;

const goEnv = {};
if (targetOS) {
  const goosMap = { windows: "windows", linux: "linux", darwin: "darwin" };
  goEnv.GOOS = goosMap[targetOS] || targetOS;
  goEnv.GOARCH = "amd64";
  console.log(
    `  Cross-compiling for GOOS=${goEnv.GOOS} GOARCH=${goEnv.GOARCH}`,
  );
}

run(`go build -o ${path.join(ROOT, outputName)} ./cmd/server`, GO_DIR, goEnv);

console.log(`
========================================
  ✅ Build complete!
  Binary: ./${outputName}
========================================
`);
