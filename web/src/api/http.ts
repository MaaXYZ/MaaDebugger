import type {
  AdbDeviceInfo,
  Win32WindowInfo,
  ConnectControllerRequest,
  StatusSnapshot,
  MethodItems,
} from "@/types/api";
import type { InterfaceParseResult } from "@/types/interface";

import type { NodeDataResponse } from "@/components/Index/taskDetail/types";

export interface ApiResponse<T = unknown> {
  succeed: boolean;
  msg: string;
  data?: T;
}

/**
 * API 基础 URL
 * 开发模式下 Vite proxy 会转发 /api → 后端
 * 生产模式下同源
 */
const BASE_URL = "/api";

const REQUEST_TIMEOUT_MS = 10000;

function isApiResponse<T>(value: unknown): value is ApiResponse<T> {
  if (!value || typeof value !== "object") return false;
  const obj = value as Record<string, unknown>;
  return typeof obj.succeed === "boolean" && typeof obj.msg === "string";
}

function extractMessage(payload: unknown, fallback: string): string {
  if (typeof payload === "string" && payload.trim()) {
    return payload;
  }
  if (payload && typeof payload === "object") {
    const maybeMsg = (payload as Record<string, unknown>).msg;
    if (typeof maybeMsg === "string" && maybeMsg.trim()) {
      return maybeMsg;
    }
  }
  return fallback;
}

/**
 * 通用 HTTP 请求封装
 */
async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<ApiResponse<T>> {
  const url = `${BASE_URL}${path}`;
  const timeoutController = new AbortController();
  const timeoutId = setTimeout(() => {
    timeoutController.abort(
      new DOMException("Request timeout", "TimeoutError"),
    );
  }, REQUEST_TIMEOUT_MS);

  let externalAbortHandler: (() => void) | null = null;
  if (options.signal) {
    if (options.signal.aborted) {
      timeoutController.abort(options.signal.reason);
    } else {
      externalAbortHandler = () =>
        timeoutController.abort(options.signal?.reason);
      options.signal.addEventListener("abort", externalAbortHandler, {
        once: true,
      });
    }
  }

  try {
    const response = await fetch(url, {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      ...options,
      signal: timeoutController.signal,
    });

    const contentType = response.headers.get("content-type") ?? "";
    const isJson = contentType.includes("application/json");

    if (!response.ok) {
      let payload: unknown = null;
      if (response.status !== 204) {
        payload = isJson ? await response.json() : await response.text();
      }
      return {
        succeed: false,
        msg: extractMessage(
          payload,
          `HTTP ${response.status} ${response.statusText}`,
        ),
        data: undefined,
      };
    }

    if (response.status === 204) {
      return {
        succeed: true,
        msg: "ok",
        data: undefined,
      };
    }

    if (!isJson) {
      const text = await response.text();
      return {
        succeed: false,
        msg: extractMessage(
          text,
          `Unexpected response content-type: ${contentType || "unknown"}`,
        ),
        data: undefined,
      };
    }

    const payload = (await response.json()) as unknown;
    if (isApiResponse<T>(payload)) {
      return payload;
    }

    return {
      succeed: true,
      msg: "ok",
      data: payload as T,
    };
  } catch (err) {
    const isAborted = err instanceof DOMException && err.name === "AbortError";
    const isTimeout =
      timeoutController.signal.aborted && !options.signal?.aborted;

    return {
      succeed: false,
      msg: isAborted
        ? isTimeout
          ? "Request timeout"
          : "Request aborted"
        : err instanceof Error
          ? err.message
          : "Network error",
      data: undefined,
    };
  } finally {
    clearTimeout(timeoutId);
    if (options.signal && externalAbortHandler) {
      options.signal.removeEventListener("abort", externalAbortHandler);
    }
  }
}

// ============================================================
// Controller API
// ============================================================

/**
 * 检测 ADB 设备
 */
export async function detectAdbDevices(): Promise<AdbDeviceInfo[]> {
  const result = await request<AdbDeviceInfo[]>("/controller/detect/adb");
  if (!result.succeed) {
    throw new Error(result.msg || "Detect ADB failed");
  }
  return Array.isArray(result.data) ? result.data : [];
}

/**
 * 检测桌面窗口（Win32/Gamepad 共用）
 */
export async function detectDesktopWindows(
  classRegex?: string,
  windowRegex?: string,
): Promise<Win32WindowInfo[]> {
  const params = new URLSearchParams();
  if (classRegex) params.set("class_regex", classRegex);
  if (windowRegex) params.set("window_regex", windowRegex);
  const query = params.toString();
  const path = `/controller/detect/desktop${query ? `?${query}` : ""}`;
  const result = await request<Win32WindowInfo[]>(path);
  return result.data ?? [];
}

/**
 * 连接 Controller
 */
export async function connectController(
  params: ConnectControllerRequest,
): Promise<ApiResponse> {
  return request("/controller/connect", {
    method: "POST",
    body: JSON.stringify(params),
  });
}

/**
 * 断开 Controller
 */
export async function disconnectController(): Promise<ApiResponse> {
  return request("/controller/disconnect", {
    method: "POST",
  });
}

// ============================================================
// Agent API
// ============================================================

export interface AgentInfo {
  identifier: string;
  status: string;
  error?: string;
}

/**
 * 创建并连接 Agent（每次都重新创建以避免状态残留）
 */
export async function connectAgent(identifier: string): Promise<ApiResponse> {
  return request("/agent/connect", {
    method: "POST",
    body: JSON.stringify({ identifier }),
  });
}

/**
 * 断开 Agent
 */
export async function disconnectAgent(
  identifier: string,
): Promise<ApiResponse> {
  return request("/agent/disconnect", {
    method: "POST",
    body: JSON.stringify({ identifier }),
  });
}

/**
 * 获取 Agent 列表
 */
export async function getAgentList(): Promise<AgentInfo[]> {
  const result = await request<AgentInfo[]>("/agent/list");
  return Array.isArray(result.data) ? result.data : [];
}

// ============================================================
// Resource API
// ============================================================

/**
 * 校验路径是否存在
 */
export async function checkPathExists(
  path: string,
  pathType?: "file" | "dir",
): Promise<ApiResponse<{ exists: boolean }>> {
  return request<{ exists: boolean }>("/path/exists", {
    method: "POST",
    body: JSON.stringify({ path, type: pathType }),
  });
}

/**
 * 解析 interface 文件
 */
export async function parseInterface(
  path: string,
): Promise<ApiResponse<InterfaceParseResult>> {
  return request("/interface/parse", {
    method: "POST",
    body: JSON.stringify({ path }),
  });
}

/**
 * 加载资源
 */
export async function loadResource(paths: string[]): Promise<ApiResponse> {
  return request("/resource/load", {
    method: "POST",
    body: JSON.stringify({ paths }),
  });
}

// ============================================================
// Task API
// ============================================================

/**
 * 获取指定节点的最新详情（含 reco detail + action detail）
 */
export async function getNodeDetail(
  name: string,
): Promise<
  import("@/components/Index/taskDetail/types").NodeDetailResponse | null
> {
  const result = await request<
    import("@/components/Index/taskDetail/types").NodeDetailResponse
  >(`/task/node/${encodeURIComponent(name)}`);
  return result.data ?? null;
}

/**
 * 获取运行时节点原始定义
 */
export async function getNodeData(
  name: string,
  options: { recoId?: number | null; actionId?: number | null } = {},
): Promise<NodeDataResponse | null> {
  const params = new URLSearchParams();
  if (options.recoId != null) {
    params.set("reco_id", String(options.recoId));
  }
  if (options.actionId != null) {
    params.set("action_id", String(options.actionId));
  }
  const query = params.toString();
  const result = await request<
    import("@/components/Index/taskDetail/types").NodeDataResponse
  >(`/task/node-data/${encodeURIComponent(name)}${query ? `?${query}` : ""}`);
  return result.data ?? null;
}

/**
 * 通过 reco_id 获取识别详情
 */
export async function getRecoDetailById(
  recoId: number,
): Promise<
  import("@/components/Index/taskDetail/types").RecoDetailResponse | null
> {
  const result = await request<
    import("@/components/Index/taskDetail/types").RecoDetailResponse
  >(`/task/reco/${recoId}`);
  return result.data ?? null;
}

/**
 * 通过 action_id 获取动作详情
 */
export async function getActionDetailById(
  actionId: number,
): Promise<
  import("@/components/Index/taskDetail/types").ActionDetailResponse | null
> {
  const result = await request<
    import("@/components/Index/taskDetail/types").ActionDetailResponse
  >(`/task/action/${actionId}`);
  return result.data ?? null;
}

export function getTaskImageUrl(imageId: string): string {
  return `${BASE_URL}/task/image/${encodeURIComponent(imageId)}`;
}

/**
 * 获取可运行节点列表
 */
export async function getTaskNodes(): Promise<string[]> {
  const result = await request<string[]>("/task/nodes");
  return Array.isArray(result.data) ? result.data : [];
}

/**
 * 运行任务
 */
export async function runTask(
  entry: string,
  pipelineOverride: Record<string, unknown> = {},
): Promise<ApiResponse> {
  return request("/task/run", {
    method: "POST",
    body: JSON.stringify({
      entry,
      pipeline_override: pipelineOverride,
    }),
  });
}

/**
 * 停止任务
 */
export async function stopTask(): Promise<ApiResponse> {
  return request("/task/stop", {
    method: "POST",
  });
}
// ============================================================
// Clear API
// ============================================================

export async function clearCache(): Promise<ApiResponse> {
  return request("/clear/cache", { method: "POST" });
}

type ControllerMethodType =
  | "adb_screencap"
  | "adb_input"
  | "window_screencap"
  | "win32_input"
  | "gamepad_type";

export async function getControllerMethod(
  methodType: ControllerMethodType,
): Promise<ApiResponse<MethodItems[]>> {
  return request(`/controller/methods?method_type=${methodType}`, {
    method: "GET",
  });
}

// ============================================================
// Screenshot API
// ============================================================

export interface ScreenshotStatus {
  running: boolean;
  paused: boolean;
  fps: number;
  overlay_state: "none" | "disconnected" | "paused" | "failed";
  overlay_message: string;
}

export async function startScreenshot(): Promise<ApiResponse> {
  return request("/screenshot/start", { method: "POST" });
}

export async function stopScreenshot(): Promise<ApiResponse> {
  return request("/screenshot/stop", { method: "POST" });
}

export async function pauseScreenshot(): Promise<ApiResponse> {
  return request("/screenshot/pause", { method: "POST" });
}

export async function resumeScreenshot(): Promise<ApiResponse> {
  return request("/screenshot/resume", { method: "POST" });
}

export async function setScreenshotFPS(
  fps: number,
): Promise<ApiResponse<{ fps: number }>> {
  return request("/screenshot/fps", {
    method: "PUT",
    body: JSON.stringify({ fps }),
  });
}

export async function getScreenshotStatus(): Promise<ScreenshotStatus | null> {
  const result = await request<ScreenshotStatus>("/screenshot/status");
  return result.data ?? null;
}

// ============================================================
// Info API
// ============================================================

/**
 * 获取 MaaDebugger 版本
 */
export async function getMaaDebuggerInfos(): Promise<Record<string, string>> {
  const result = await request<Record<string, string>>("/info/all");
  return result.data ?? {};
}

/**
 * 获取 MaaFW 版本
 */
export async function getMaaFrameworkVersion(): Promise<string> {
  const result = await request<string>("/fw/version");
  return result.data ?? "unknown";
}
/**
 * 获取当前频道
 */
export async function getChannel(): Promise<string> {
  const result = await request<string>("/channel");
  return result.data ?? "github";
}

/**
 * 获取当前状态
 */
export async function getStatusSnapshot(): Promise<StatusSnapshot | null> {
  const result = await request<StatusSnapshot>("/info/status");
  return result.data ?? null;
}

/**
 * 获取 UAC 状态
 */
export async function getUACStatus(): Promise<boolean> {
  const result = await request<boolean>("/info/uac");
  return result.data ?? false;
}

// ============================================================
// Config 持久化 API
// ============================================================

/**
 * 获取所有配置
 */
export async function getAllConfig(): Promise<Record<string, unknown>> {
  const result = await request<Record<string, unknown>>("/config");
  return (result.data ?? {}) as Record<string, unknown>;
}

/**
 * 获取某个 store 的配置
 */
export async function getStoreConfig<T = unknown>(
  key: string,
): Promise<T | null> {
  const result = await request<T>(`/config/${key}`);
  return result.data ?? null;
}

/**
 * 保存某个 store 的配置
 */
export async function saveStoreConfig(
  key: string,
  value: unknown,
): Promise<void> {
  await request(`/config/${key}`, {
    method: "PUT",
    body: JSON.stringify(value),
  });
}

// ============================================================
// Update Check API
// ============================================================

export interface UpdateCheckResult {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  note?: string;
  nightly: boolean;
  track: string;
}

/**
 * Check for updates
 * Nightly detection is handled automatically by the server
 * (channel == npm and version is a commit hash)
 */
export async function checkForUpdates(
  showPreRelease = false,
): Promise<UpdateCheckResult | null> {
  const query = new URLSearchParams({
    showPre: String(showPreRelease),
  });
  const result = await request<UpdateCheckResult>(
    `/update/check?${query.toString()}`,
  );
  return result.data ?? null;
}
