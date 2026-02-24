import { Hono } from "hono";
import type { ApiResponse, AdbDeviceInfo } from "@shared/types/api";
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

  // TODO: Win32 / Gamepad 连接
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
