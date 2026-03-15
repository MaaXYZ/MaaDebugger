/**
 * 前端类型定义 — 与 Go 后端 JSON 序列化保持一致
 *
 * Go 来源:
 *   - response.Envelope          → ApiResponse
 *   - httpapi.adbDeviceInfo      → AdbDeviceInfo
 *   - httpapi.desktopWindowInfo  → Win32WindowInfo
 *   - state.Snapshot             → StatusSnapshot
 *   - httpapi.handleControllerConnect payload → ConnectControllerRequest
 */

/**
 * Controller Methods Info
 * @see server/internal/maaservice/controller_methods.go
 */
export interface MethodItems {
  label: string;
  value: string;
  icon?: string;
}

/**
 * ADB 设备信息
 * @see server/internal/httpapi/router.go — adbDeviceInfo
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
 * @see server/internal/httpapi/router.go — desktopWindowInfo
 */
export interface Win32WindowInfo {
  hwnd: string;
  window_name: string;
  class_name: string;
}

/**
 * Controller 连接请求
 * @see server/internal/httpapi/router.go — handleControllerConnect
 */
export interface ConnectControllerRequest {
  type: "adb" | "win32" | "gamepad" | "playcover" | "wlroot" | "custom";

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

  // PlayCover 参数
  playcover_address?: string;
  playcover_uuid?: string;

  // WlRoot 参数
  wlroot_socket_path?: string;
}

/**
 * Controller 连接状态
 * @see server/internal/state/store.go — Snapshot.Controller
 */
export type ControllerStatus = "disconnected" | "connecting" | "connected";

/**
 * Resource 加载状态
 * @see server/internal/state/store.go — Snapshot.Resource
 */
export type ResourceStatus = "unloaded" | "loading" | "loaded" | "failed";

/**
 * Task 运行状态
 * @see server/internal/state/store.go — Snapshot.Task
 */
export type TaskStatus = "idle" | "running" | "success" | "failed" | "stopped";

/**
 * Agent 状态
 * @see server/internal/state/store.go — Snapshot.Agent
 */
export type AgentStatus = "disconnected" | "connecting" | "connected";

/**
 * 全局状态快照
 * @see server/internal/state/store.go — Snapshot
 */
export interface StatusSnapshot {
  controller: ControllerStatus;
  resource: ResourceStatus;
  task: TaskStatus;
  agent: AgentStatus;
}
