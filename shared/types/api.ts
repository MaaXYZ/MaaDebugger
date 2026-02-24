/**
 * 统一 REST API 响应格式
 */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
}

/**
 * ADB 设备信息（从 maa-node 返回后映射为此结构）
 */
export interface AdbDeviceInfo {
  name: string;
  adb_path: string;
  address: string;
  screencap_methods: string;
  input_methods: string;
  config: string;
}

/**
 * Win32 窗口信息
 */
export interface Win32WindowInfo {
  hwnd: string;
  window_name: string;
  class_name: string;
}

/**
 * Controller 连接请求
 */
export interface ConnectControllerRequest {
  type: "adb" | "win32" | "gamepad";

  // ADB 参数
  adb_path?: string;
  adb_address?: string;
  adb_config?: string;
  adb_screencap_method?: string;
  adb_input_method?: string;

  // Win32 参数
  hwnd?: string;
  win32_screencap_method?: string;
  win32_mouse_method?: string;
  win32_keyboard_method?: string;

  // Gamepad 参数
  gamepad_type?: string;
  gamepad_screencap_method?: string;
}

/**
 * Controller 连接状态
 */
export type ControllerStatus = "disconnected" | "connecting" | "connected";

/**
 * Resource 加载状态
 */
export type ResourceStatus = "unloaded" | "loading" | "loaded";

/**
 * Task 运行状态
 */
export type TaskStatus = "idle" | "running" | "success" | "failed";

/**
 * Agent 状态
 */
export type AgentStatus = "disconnected" | "connecting" | "connected";

/**
 * 全局状态快照
 */
export interface StatusSnapshot {
  controller: ControllerStatus;
  resource: ResourceStatus;
  task: TaskStatus;
  agent: AgentStatus;
}
