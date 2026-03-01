import { defineStore } from "pinia";
import { ref } from "vue";

export const DEFAULT_SCREENCAP_METHOD = "18446744073709551559";
export const DEFAULT_INPUT_METHOD = "18446744073709551607";

/** Win32/Gamepad 共享默认截图方法 */
export const DEFAULT_DESKTOP_SCREENCAP = "1"; // GDI

/** Win32 默认输入方法 */
export const DEFAULT_WIN32_MOUSE = "1"; // Seize
export const DEFAULT_WIN32_KEYBOARD = "1"; // Seize

/** Gamepad 默认类型 */
export const DEFAULT_GAMEPAD_TYPE = "0"; // Xbox 360

/**
 * Controller Store — 持久化 ADB / Win32 / Gamepad 连接配置
 *
 * Win32 和 Gamepad 共享以下字段：
 * - hwnd, className, windowName（选中窗口）
 * - screencapMethod（截图方法）
 *
 * 仅 input 部分不同：
 * - Win32: mouseMethod + keyboardMethod
 * - Gamepad: gamepadType
 */
export const useControllerStore = defineStore(
  "controller",
  () => {
    // 控制器类型
    const controllerType = ref("adb");

    // --- ADB 配置 ---
    const adbPath = ref("");
    const adbAddress = ref("");
    const screencapMethod = ref(DEFAULT_SCREENCAP_METHOD);
    const inputMethod = ref(DEFAULT_INPUT_METHOD);
    const adbConfig = ref("");
    const selectedAdbDevice = ref("");

    // --- 桌面共享配置（Win32 / Gamepad 共用） ---
    const desktopHwnd = ref("");
    const desktopClassName = ref("");
    const desktopWindowName = ref("");
    const desktopScreencapMethod = ref(DEFAULT_DESKTOP_SCREENCAP);
    const selectedDesktopWindow = ref(""); // 选中窗口的 hwnd
    const desktopClassFilter = ref(""); // 搜索窗口时的类名过滤
    const desktopWindowRegex = ref(""); // 搜索窗口时的窗口名正则

    // --- Win32 独有配置 ---
    const win32MouseMethod = ref(DEFAULT_WIN32_MOUSE);
    const win32KeyboardMethod = ref(DEFAULT_WIN32_KEYBOARD);

    // --- Gamepad 独有配置 ---
    const gamepadType = ref(DEFAULT_GAMEPAD_TYPE);

    // --- PlayCover 配置 ---
    const playcoverAddress = ref("");
    const playcoverUuid = ref("");

    // 连接中状态（瞬态，不持久化，但全局可访问）
    const connecting = ref(false);

    function updateAdbConfig(config: {
      adb_path: string;
      adb_address: string;
      screencap_method: string;
      input_method: string;
      adb_config?: string;
    }) {
      adbPath.value = config.adb_path;
      adbAddress.value = config.adb_address;
      screencapMethod.value = config.screencap_method;
      inputMethod.value = config.input_method;
      if (config.adb_config !== undefined) {
        adbConfig.value = config.adb_config;
      }
    }

    /**
     * 更新桌面共享配置（hwnd + 窗口信息 + 截图方法）
     */
    function updateDesktopConfig(config: {
      hwnd: string;
      class_name?: string;
      window_name?: string;
      screencap_method: string;
    }) {
      desktopHwnd.value = config.hwnd;
      desktopClassName.value = config.class_name ?? "";
      desktopWindowName.value = config.window_name ?? "";
      desktopScreencapMethod.value = config.screencap_method;
      selectedDesktopWindow.value = config.hwnd;
    }

    /**
     * 更新 Win32 独有配置
     */
    function updateWin32Input(config: {
      mouse_method: string;
      keyboard_method: string;
    }) {
      win32MouseMethod.value = config.mouse_method;
      win32KeyboardMethod.value = config.keyboard_method;
    }

    /**
     * 更新 Gamepad 独有配置
     */
    function updateGamepadInput(config: { gamepad_type: string }) {
      gamepadType.value = config.gamepad_type;
    }

    /**
     * 更新 PlayCover 配置
     */
    function updatePlayCoverConfig(config: { address: string; uuid: string }) {
      playcoverAddress.value = config.address;
      playcoverUuid.value = config.uuid;
    }

    return {
      controllerType,
      // ADB
      adbPath,
      adbAddress,
      screencapMethod,
      inputMethod,
      adbConfig,
      selectedAdbDevice,
      // 桌面共享（Win32/Gamepad）
      desktopHwnd,
      desktopClassName,
      desktopWindowName,
      desktopScreencapMethod,
      selectedDesktopWindow,
      desktopClassFilter,
      desktopWindowRegex,
      // Win32 独有
      win32MouseMethod,
      win32KeyboardMethod,
      // Gamepad 独有
      gamepadType,
      // PlayCover
      playcoverAddress,
      playcoverUuid,
      // Shared
      connecting,
      updateAdbConfig,
      updateDesktopConfig,
      updateWin32Input,
      updateGamepadInput,
      updatePlayCoverConfig,
    };
  },
  { persist: true, persistExclude: ["connecting"] },
);
