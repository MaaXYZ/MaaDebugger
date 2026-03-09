#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";

function getRequiredEnv(name) {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

function getReleaseChannel() {
  return process.env.RELEASE_CHANNEL || "nightly";
}

function getPackageVersion(channel) {
  if (channel === "nightly") {
    const sha = process.env.GITHUB_SHA || "dev";
    return `0.1.0-nightly.${sha}`;
  }

  if (channel === "latest") {
    return getRequiredEnv("RELEASE_VERSION");
  }

  throw new Error(`Unsupported RELEASE_CHANNEL: ${channel}`);
}

function buildPublishConfig(channel) {
  return {
    access: "public",
    provenance: true,
    tag: channel,
  };
}

function main() {
  const packagePath = path.resolve(getRequiredEnv("PACKAGE_JSON_PATH"));
  const packageName = process.env.PACKAGE_NAME;
  const targetOS = process.env.TARGET_OS;
  const targetCPU = process.env.TARGET_CPU;
  const channel = getReleaseChannel();

  const pkg = JSON.parse(fs.readFileSync(packagePath, "utf8"));
  const version = getPackageVersion(channel);

  if (packageName) {
    pkg.name = packageName;
  }

  pkg.version = version;
  pkg.publishConfig = buildPublishConfig(channel);

  if (targetOS) {
    pkg.os = [targetOS];
  }

  if (targetCPU) {
    pkg.cpu = [targetCPU];
  }

  if (pkg.optionalDependencies) {
    for (const key of Object.keys(pkg.optionalDependencies)) {
      pkg.optionalDependencies[key] = version;
    }
  }

  fs.writeFileSync(packagePath, `${JSON.stringify(pkg, null, 2)}\n`);
}

main();
