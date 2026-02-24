import * as fs from "node:fs";
import * as path from "node:path";
import { fileURLToPath } from "node:url";

/**
 * 配置持久化服务
 *
 * 将前端 Pinia store 的状态持久化到服务器端的 JSON 文件中。
 * 每个 store 对应一个独立的 key，所有数据存储在一个 JSON 文件中。
 */

// 使用固定项目根目录路径，避免因启动 cwd 不同导致“持久化文件写到不同目录”
const PROJECT_ROOT = fileURLToPath(new URL("../../../", import.meta.url));
const CONFIG_DIR = path.join(PROJECT_ROOT, ".maa-debugger");
const CONFIG_FILE = path.join(CONFIG_DIR, "config.json");

// 模块加载时确保配置目录存在（只需一次）
if (!fs.existsSync(CONFIG_DIR)) {
  fs.mkdirSync(CONFIG_DIR, { recursive: true });
}

/**
 * 读取整个配置文件
 */
function readConfigFile(): Record<string, unknown> {
  if (!fs.existsSync(CONFIG_FILE)) {
    return {};
  }
  try {
    const content = fs.readFileSync(CONFIG_FILE, "utf-8");
    return JSON.parse(content) as Record<string, unknown>;
  } catch {
    console.warn("[ConfigService] Failed to parse config file, resetting.");
    return {};
  }
}

/**
 * 写入整个配置文件
 */
function writeConfigFile(data: Record<string, unknown>): void {
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(data, null, 2), "utf-8");
}

/**
 * 获取所有配置
 */
export function getAllConfig(): Record<string, unknown> {
  return readConfigFile();
}

/**
 * 获取某个 store 的持久化数据
 */
export function getConfig(key: string): unknown {
  const config = readConfigFile();
  return config[key] ?? null;
}

/**
 * 设置某个 store 的持久化数据
 */
export function setConfig(key: string, value: unknown): void {
  const config = readConfigFile();
  config[key] = value;
  writeConfigFile(config);
}

/**
 * 批量设置多个 store 的持久化数据
 */
export function setConfigs(data: Record<string, unknown>): void {
  const config = readConfigFile();
  Object.assign(config, data);
  writeConfigFile(config);
}
