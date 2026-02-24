import { defineStore } from "pinia";
import { ref } from "vue";

/** maa-node AdbScreencapMethod.Default（Uint64 字符串） */
export const DEFAULT_SCREENCAP_METHOD = "18446744073709551559";
/** maa-node AdbInputMethod.Default（Uint64 字符串） */
export const DEFAULT_INPUT_METHOD = "18446744073709551607";

/**
 * Controller Store — 持久化 ADB 连接配置
 */
export const useControllerStore = defineStore(
  "controller",
  () => {
    // ADB 配置
    const adbPath = ref("");
    const adbAddress = ref("");
    const screencapMethod = ref(DEFAULT_SCREENCAP_METHOD);
    const inputMethod = ref(DEFAULT_INPUT_METHOD);
    const adbConfig = ref("");
    const selectedAdbDevice = ref("");

    // 控制器类型
    const controllerType = ref("adb");

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

    return {
      adbPath,
      adbAddress,
      screencapMethod,
      inputMethod,
      adbConfig,
      selectedAdbDevice,
      controllerType,
      updateAdbConfig,
    };
  },
  { persist: true },
);
