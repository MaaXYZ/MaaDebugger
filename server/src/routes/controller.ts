import { Hono } from "hono";
import type {
  ApiResponse,
  AdbDeviceInfo,
  Win32WindowInfo,
} from "@shared/types/api";
import * as maaService from "../services/maa.js";

export const controllerRoutes = new Hono();

/**
 * GET /api/controller/detect/adb
 * 检测可用的 ADB 设备
 */
controllerRoutes.get("/detect/adb", async (c) => {
  const devices = await maaService.detectAdb();
  const response: ApiResponse<AdbDeviceInfo[]> = {
    success: true,
    data: devices,
  };
  return c.json(response);
});

/**
 * GET /api/controller/detect/desktop
 * 检测桌面窗口（Win32/Gamepad 共用）
 *
 * Query: ?class_regex=...&window_regex=...（可选过滤）
 */
controllerRoutes.get("/detect/desktop", async (c) => {
  let windows = await maaService.detectDesktop();

  // 可选的过滤参数
  const classRegex = c.req.query("class_regex");
  const windowRegex = c.req.query("window_regex");

  if (classRegex) {
    try {
      const re = new RegExp(classRegex);
      windows = windows.filter((w) => re.test(w.class_name));
    } catch {
      // 无效正则，忽略过滤
    }
  }

  if (windowRegex) {
    try {
      const re = new RegExp(windowRegex);
      windows = windows.filter((w) => re.test(w.window_name));
    } catch {
      // 无效正则，忽略过滤
    }
  }

  const response: ApiResponse<Win32WindowInfo[]> = {
    success: true,
    data: windows,
  };
  return c.json(response);
});

/**
 * POST /api/controller/connect
 * 连接 Controller
 *
 * Body: ConnectControllerRequest
 */
controllerRoutes.post("/connect", async (c) => {
  const body = await c.req.json();

  if (body.type === "adb") {
    // 默认值：maa-node AdbScreencapMethod.Default / AdbInputMethod.Default
    const result = await maaService.connectAdb(
      body.adb_path ?? "",
      body.adb_address ?? "",
      body.adb_screencap_method ?? maaService.DEFAULT_SCREENCAP_METHOD,
      body.adb_input_method ?? maaService.DEFAULT_INPUT_METHOD,
      body.adb_config ?? "",
    );
    const response: ApiResponse = {
      success: result.success,
      error: result.error,
    };
    return c.json(response, result.success ? 200 : 400);
  }

  if (body.type === "win32") {
    if (!body.hwnd) {
      return c.json(
        { success: false, error: "hwnd is required for Win32 controller" },
        400,
      );
    }
    const result = await maaService.connectWin32(
      body.hwnd,
      body.win32_screencap_method ?? "1", // GDI
      body.win32_mouse_method ?? "1", // Seize
      body.win32_keyboard_method ?? "1", // Seize
    );
    const response: ApiResponse = {
      success: result.success,
      error: result.error,
    };
    return c.json(response, result.success ? 200 : 400);
  }

  if (body.type === "gamepad") {
    if (!body.hwnd) {
      return c.json(
        { success: false, error: "hwnd is required for Gamepad controller" },
        400,
      );
    }
    const result = await maaService.connectGamepad(
      body.hwnd,
      body.gamepad_screencap_method ?? "1", // GDI
      body.gamepad_type ?? "0", // Xbox 360
    );
    const response: ApiResponse = {
      success: result.success,
      error: result.error,
    };
    return c.json(response, result.success ? 200 : 400);
  }

  const response: ApiResponse = {
    success: false,
    error: `Unsupported controller type: ${body.type}`,
  };
  return c.json(response, 400);
});

/**
 * POST /api/controller/disconnect
 * 断开 Controller
 */
controllerRoutes.post("/disconnect", async (c) => {
  await maaService.disconnectController();
  const response: ApiResponse = { success: true };
  return c.json(response);
});
