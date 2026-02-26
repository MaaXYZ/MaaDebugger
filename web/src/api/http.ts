import type {
  AdbDeviceInfo,
  Win32WindowInfo,
  ConnectControllerRequest,
  StatusSnapshot,
} from "@/types/api";

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

/**
 * 通用 HTTP 请求封装
 */
async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<ApiResponse<T>> {
  const url = `${BASE_URL}${path}`;

  try {
    const response = await fetch(url, {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      ...options,
    });

    const data = (await response.json()) as ApiResponse<T>;
    return data;
  } catch (err) {
    return {
      succeed: false,
      msg: err instanceof Error ? err.message : "Network error",
      data: undefined,
    };
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
export async function connectAgent(
  identifier: string,
): Promise<ApiResponse> {
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
// Screenshot API
// ============================================================

export interface ScreenshotStatus {
  running: boolean;
  paused: boolean;
  fps: number;
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
 * 获取 MaaFW 版本
 */
export async function getVersion(): Promise<string> {
  const result = await request<string>("/info/version");
  return result.data ?? "unknown";
}

/**
 * 获取当前状态
 */
export async function getStatusSnapshot(): Promise<StatusSnapshot | null> {
  const result = await request<StatusSnapshot>("/info/status");
  return result.data ?? null;
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
