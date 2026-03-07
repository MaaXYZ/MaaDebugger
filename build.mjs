#!/usr/bin/env node

/**
 * Build script: compile Vue frontend → embed into Go binary
 *
 * Usage:
 *   node build.mjs                          # build for current OS/arch
 *   node build.mjs --os linux               # cross-compile for linux
 *   node build.mjs --os windows             # cross-compile for windows
 *   node build.mjs --os darwin              # cross-compile for macOS
 *   node build.mjs --arch arm64             # cross-compile for arm64
 *   node build.mjs --os linux --arch arm64  # cross-compile for linux/arm64
 *   node build.mjs --skip-frontend          # skip frontend build
 *   node build.mjs --skip-go               # skip Go build (frontend only)
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
const skipGo = args.includes("--skip-go");
const osIndex = args.indexOf("--os");
const targetOS = osIndex !== -1 ? args[osIndex + 1] : undefined;
const archIndex = args.indexOf("--arch");
const targetArch = archIndex !== -1 ? args[archIndex + 1] : undefined;

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

const buildFrontend = !skipFrontend;
const buildGo = !skipGo;

const totalSteps = (buildFrontend ? 2 : 0) + (buildGo ? 1 : 0);
let currentStep = 0;

if (buildFrontend) {
  // ── Step 1: Install frontend dependencies ──
  step(++currentStep, totalSteps, "Installing frontend dependencies...");
  run("pnpm install --frozen-lockfile", WEB_DIR);

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
if (buildGo) {
  step(++currentStep, totalSteps, "Building Go binary...");

  const goosMap = { windows: "windows", linux: "linux", darwin: "darwin" };
  const resolvedOS = targetOS
    ? goosMap[targetOS] || targetOS
    : process.platform === "win32"
      ? "windows"
      : process.platform;

  const version = process.env.VERSION || "dev";
  const commitSHA = process.env.COMMIT_SHA || "";
  const buildTime = process.env.BUILD_TIME || `${Date.now()}`;
  const ldPath = "github.com/MaaXYZ/MaaDebugger/internal/buildinfo";
  const ldFlags = `-s -w -X ${ldPath}.Version=${version} -X ${ldPath}.CommitSHA=${commitSHA} -X ${ldPath}.BuildTime=${buildTime}`;

  const ext = resolvedOS === "windows" ? ".exe" : "";
  const outputName = `MaaDebugger${ext}`;

  const goEnv = {
    CGO_ENABLED: "0",
  };

  if (targetOS) {
    goEnv.GOOS = goosMap[targetOS] || targetOS;
  }
  if (targetArch) {
    goEnv.GOARCH = targetArch;
  }

  if (targetOS || targetArch) {
    console.log(
      `  Cross-compiling for GOOS=${goEnv.GOOS || "(current)"} GOARCH=${goEnv.GOARCH || "(current)"}`,
    );
  }

  run(
    `go build -trimpath -ldflags "${ldFlags}" -o ${path.join(ROOT, outputName)} ./cmd/server`,
    GO_DIR,
    goEnv,
  );

  console.log(`
========================================
  ✅ Build complete!
  Binary: ./${outputName}
========================================
`);
}
