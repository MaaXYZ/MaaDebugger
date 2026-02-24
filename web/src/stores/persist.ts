import type { PiniaPlugin, StateTree } from "pinia";
import { watch } from "vue";
import { getStoreConfig, saveStoreConfig } from "@/api/http";

/**
 * Pinia 服务器持久化插件
 *
 * 在 store 初始化时从服务器加载持久化数据，
 * 在 state 变化时自动将数据保存到服务器。
 *
 * 使用方式：在 store 的 options 中设置 `persist: true`
 *
 * ```ts
 * export const useMyStore = defineStore("myStore", () => {
 *   // ...
 * }, { persist: true })
 * ```
 */

declare module "pinia" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  export interface DefineStoreOptionsBase<S extends StateTree, Store> {
    /**
     * 是否将此 store 持久化到服务器
     */
    persist?: boolean;
  }
}

let saveTimer: ReturnType<typeof setTimeout> | null = null;
const pendingSaves = new Map<string, unknown>();

/**
 * 防抖批量保存，避免频繁请求服务器
 */
function debouncedSave(storeId: string, state: unknown) {
  pendingSaves.set(storeId, state);

  if (saveTimer) {
    clearTimeout(saveTimer);
  }

  saveTimer = setTimeout(async () => {
    const entries = Array.from(pendingSaves.entries());
    pendingSaves.clear();
    saveTimer = null;

    for (const [id, data] of entries) {
      try {
        await saveStoreConfig(id, data);
      } catch (err) {
        console.error(`[PersistPlugin] Failed to save store "${id}":`, err);
      }
    }
  }, 500);
}

export const serverPersistPlugin: PiniaPlugin = ({ store, options }) => {
  if (!options.persist) return;

  const storeId = store.$id;

  // 从服务器加载持久化数据
  getStoreConfig(storeId)
    .then((savedState) => {
      if (savedState && typeof savedState === "object") {
        store.$patch(savedState as StateTree);
        // 恢复后同步自增 ID，避免 ID 冲突
        if (typeof (store as any).syncIds === "function") {
          (store as any).syncIds();
        }
      }
    })
    .catch((err) => {
      console.error(
        `[PersistPlugin] Failed to load store "${storeId}":`,
        err,
      );
    });

  // 监听 state 变化并保存到服务器（防抖）
  watch(
    () => store.$state,
    (newState) => {
      debouncedSave(storeId, JSON.parse(JSON.stringify(newState)));
    },
    { deep: true },
  );
};
