import type {
  ApiResponse,
  AdbDeviceInfo,
  Win32WindowInfo,
  ConnectControllerRequest,
  StatusSnapshot,
} from "@shared/types/api";

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
      success: false,
      error: err instanceof Error ? err.message : "Network error",
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
  return result.data ?? [];
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
// Resource API
// ============================================================

/**
 * 加载资源
 */
export async function loadResource(
  paths: string[],
): Promise<ApiResponse> {
  return request("/resource/load", {
    method: "POST",
    body: JSON.stringify({ paths }),
  });
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
