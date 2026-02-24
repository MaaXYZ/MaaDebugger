import type {
  AdbDeviceInfo,
  Win32WindowInfo,
  ControllerStatus,
  ResourceStatus,
  TaskStatus,
  AgentStatus,
  StatusSnapshot,
} from "@shared/types/api";
import { broadcaster } from "../ws/broadcaster.js";

/**
 * MaaFW 服务层
 *
 * 封装 @maaxyz/maa-node 调用，管理 Controller/Resource/Tasker 实例。
 * 使用 createRequire 延迟加载 maa-node，因为 native addon
 * 在 ESM 环境下需要通过 CJS require 来加载。
 */

import { createRequire } from "node:module";
const esmRequire = createRequire(import.meta.url);

let maaModule: typeof globalThis.maa | null = null;

function getMaa(): typeof globalThis.maa {
  const m = maaModule ?? (maaModule = esmRequire("@maaxyz/maa-node"));
  return m;
}

// --- 常量 ---
/** maa-node AdbScreencapMethod.Default（Uint64 字符串） */
export const DEFAULT_SCREENCAP_METHOD = "18446744073709551559";
/** maa-node AdbInputMethod.Default（Uint64 字符串） */
export const DEFAULT_INPUT_METHOD = "18446744073709551607";

// --- 状态 ---
let controllerStatus: ControllerStatus = "disconnected";
let resourceStatus: ResourceStatus = "unloaded";
let taskStatus: TaskStatus = "idle";
let agentStatus: AgentStatus = "disconnected";

// --- 实例 ---
let controller: maa.Controller | null = null;
let resource: maa.Resource | null = null;
let tasker: maa.Tasker | null = null;

/**
 * 获取当前状态快照
 */
export function getStatus(): StatusSnapshot {
  return {
    controller: controllerStatus,
    resource: resourceStatus,
    task: taskStatus,
    agent: agentStatus,
  };
}

/**
 * 广播状态更新
 */
function broadcastStatus(): void {
  broadcaster.broadcast({
    type: "status.update",
    payload: getStatus(),
  });
}

/**
 * 更新 Controller 状态
 */
function setControllerStatus(status: ControllerStatus): void {
  controllerStatus = status;
  broadcastStatus();
}

/**
 * 更新 Resource 状态
 */
function setResourceStatus(status: ResourceStatus): void {
  resourceStatus = status;
  broadcastStatus();
}

/**
 * 更新 Task 状态
 */
function setTaskStatus(status: TaskStatus): void {
  taskStatus = status;
  broadcastStatus();
}

// ============================================================
// Controller 操作
// ============================================================

/**
 * 检测 ADB 设备
 */
export async function detectAdb(): Promise<AdbDeviceInfo[]> {
  try {
    const maa = getMaa();
    const devices = await maa.AdbController.find();
    if (!devices) return [];

    // find() 返回的是 AdbDevice 元组数组: [name, adb_path, address, screencap_methods, input_methods, config][]
    return devices.map((d: maa.AdbDevice) => ({
      name: d[0] ?? "",
      adb_path: d[1] ?? "",
      address: d[2] ?? "",
      screencap_methods: d[3] ?? "",
      input_methods: d[4] ?? "",
      config: d[5] ?? "",
    }));
  } catch (err) {
    console.error("[MaaService] detectAdb failed:", err);
    return [];
  }
}

/**
 * 连接 ADB 控制器
 */
export async function connectAdb(
  adbPath: string,
  address: string,
  screencapMethod: maa.ScreencapOrInputMethods,
  inputMethod: maa.ScreencapOrInputMethods,
  config: string = "",
): Promise<{ success: boolean; error?: string }> {
  try {
    const maa = getMaa();

    // 断开旧连接
    if (controller) {
      controller.destroy();
      controller = null;
    }

    setControllerStatus("connecting");

    controller = new maa.AdbController(
      adbPath,
      address,
      screencapMethod,
      inputMethod,
      config,
    );

    await controller.post_connection().wait();

    if (controller.connected) {
      setControllerStatus("connected");
      return { success: true };
    } else {
      controller.destroy();
      controller = null;
      setControllerStatus("disconnected");
      return { success: false, error: `Failed to connect ADB: ${address}` };
    }
  } catch (err) {
    controller = null;
    setControllerStatus("disconnected");
    const message = err instanceof Error ? err.message : String(err);
    console.error("[MaaService] connectAdb failed:", message);
    return { success: false, error: message };
  }
}

/**
 * 断开 Controller
 */
export async function disconnectController(): Promise<void> {
  if (controller) {
    controller.destroy();
    controller = null;
  }
  setControllerStatus("disconnected");
}

/**
 * 检测桌面窗口（Win32/Gamepad 共用）
 */
export async function detectDesktop(): Promise<Win32WindowInfo[]> {
  try {
    const maa = getMaa();
    const devices = await maa.Win32Controller.find();
    if (!devices) return [];

    // find() 返回 DesktopDevice 元组数组: [handle, class_name, window_name][]
    return devices.map((d: maa.DesktopDevice) => ({
      hwnd: String(d[0]),
      class_name: d[1] ?? "",
      window_name: d[2] ?? "",
    }));
  } catch (err) {
    console.error("[MaaService] detectDesktop failed:", err);
    return [];
  }
}

/**
 * 连接 Win32 控制器
 */
export async function connectWin32(
  hwnd: maa.DesktopHandle,
  screencapMethod: maa.ScreencapOrInputMethods,
  mouseMethod: maa.ScreencapOrInputMethods,
  keyboardMethod: maa.ScreencapOrInputMethods,
): Promise<{ success: boolean; error?: string }> {
  try {
    const maa = getMaa();

    // 断开旧连接
    if (controller) {
      controller.destroy();
      controller = null;
    }

    setControllerStatus("connecting");

    controller = new maa.Win32Controller(
      hwnd,
      screencapMethod,
      mouseMethod,
      keyboardMethod,
    );

    await controller.post_connection().wait();

    if (controller.connected) {
      setControllerStatus("connected");
      return { success: true };
    } else {
      controller.destroy();
      controller = null;
      setControllerStatus("disconnected");
      return { success: false, error: `Failed to connect Win32 hwnd: ${hwnd}` };
    }
  } catch (err) {
    controller = null;
    setControllerStatus("disconnected");
    const message = err instanceof Error ? err.message : String(err);
    console.error("[MaaService] connectWin32 failed:", message);
    return { success: false, error: message };
  }
}

/**
 * 连接 Gamepad 控制器
 */
export async function connectGamepad(
  hwnd: maa.DesktopHandle,
  screencapMethod: maa.ScreencapOrInputMethods,
  gamepadType: maa.ScreencapOrInputMethods,
): Promise<{ success: boolean; error?: string }> {
  try {
    const maa = getMaa();

    // 断开旧连接
    if (controller) {
      controller.destroy();
      controller = null;
    }

    setControllerStatus("connecting");

    controller = new maa.GamepadController(hwnd, screencapMethod, gamepadType);

    await controller.post_connection().wait();

    if (controller.connected) {
      setControllerStatus("connected");
      return { success: true };
    } else {
      controller.destroy();
      controller = null;
      setControllerStatus("disconnected");
      return {
        success: false,
        error: `Failed to connect Gamepad hwnd: ${hwnd}`,
      };
    }
  } catch (err) {
    controller = null;
    setControllerStatus("disconnected");
    const message = err instanceof Error ? err.message : String(err);
    console.error("[MaaService] connectGamepad failed:", message);
    return { success: false, error: message };
  }
}

// ============================================================
// Resource 操作
// ============================================================

/**
 * 加载资源路径列表
 */
export async function loadResource(
  paths: string[],
): Promise<{ success: boolean; error?: string }> {
  try {
    const maa = getMaa();

    if (!resource) {
      resource = new maa.Resource();
    }

    setResourceStatus("loading");

    for (const p of paths) {
      await resource.post_bundle(p).wait();
    }

    setResourceStatus("loaded");
    return { success: true };
  } catch (err) {
    setResourceStatus("unloaded");
    const message = err instanceof Error ? err.message : String(err);
    console.error("[MaaService] loadResource failed:", message);
    return { success: false, error: message };
  }
}

// ============================================================
// Task 操作
// ============================================================

/**
 * 获取可用节点列表
 */
export function getNodeList(): string[] {
  if (!resource) return [];
  try {
    return resource.node_list ?? [];
  } catch {
    return [];
  }
}

/**
 * 运行任务
 */
export async function runTask(
  entry: string,
  pipelineOverride: Record<string, unknown> = {},
): Promise<{ success: boolean; error?: string }> {
  try {
    const maa = getMaa();

    if (!controller) {
      return { success: false, error: "Controller is not connected." };
    }
    if (!resource) {
      return { success: false, error: "Resource is not loaded." };
    }

    if (!tasker) {
      tasker = new maa.Tasker();
    }

    tasker.controller = controller;
    tasker.resource = resource;

    if (!tasker.inited) {
      return { success: false, error: "Failed to initialize Tasker." };
    }

    setTaskStatus("running");

    const result = await tasker.post_task(entry, pipelineOverride).wait();
    const succeeded = result?.succeeded ?? false;

    setTaskStatus(succeeded ? "success" : "failed");
    return { success: succeeded };
  } catch (err) {
    setTaskStatus("failed");
    const message = err instanceof Error ? err.message : String(err);
    console.error("[MaaService] runTask failed:", message);
    return { success: false, error: message };
  }
}

/**
 * 停止当前任务
 */
export async function stopTask(): Promise<void> {
  if (tasker) {
    try {
      await tasker.post_stop().wait();
    } catch (err) {
      console.error("[MaaService] stopTask failed:", err);
    }
  }
  setTaskStatus("idle");
}

// ============================================================
// 信息查询
// ============================================================

/**
 * 获取 MaaFW 版本信息
 */
export function getVersion(): string {
  try {
    const maa = getMaa();
    return maa.Global.version ?? "unknown";
  } catch {
    return "unknown";
  }
}

/**
 * 初始化 MaaFW
 */
export function initMaa(): boolean {
  try {
    const maa = getMaa();
    maa.Global.debug_mode = true;
    console.log(`[MaaService] MaaFW version: ${maa.Global.version}`);
    return true;
  } catch (err) {
    console.warn(
      "[MaaService] MaaFW not available (maa-node not installed):",
      err instanceof Error ? err.message : err,
    );
    return false;
  }
}
