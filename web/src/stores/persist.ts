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
 *
 * 排除瞬态字段（不需要持久化的字段）：
 *
 * ```ts
 * export const useMyStore = defineStore("myStore", () => {
 *   // ...
 * }, { persist: true, persistExclude: ['connecting'] })
 * ```
 */

declare module "pinia" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  export interface DefineStoreOptionsBase<S extends StateTree, Store> {
    /**
     * 是否将此 store 持久化到服务器
     */
    persist?: boolean;
    /**
     * 不需要持久化的字段名列表（瞬态字段）
     */
    persistExclude?: string[];
  }

  export interface PiniaCustomProperties {
    /**
     * 持久化数据恢复后的回调钩子。
     * 在 store 中定义此方法，插件会在 $patch 恢复数据后自动调用。
     * 可用于：重置瞬态字段、同步自增 ID 等。
     */
    onRestore?: () => void;
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

/**
 * 从 state 中过滤掉排除字段
 */
function filterState(
  state: Record<string, unknown>,
  exclude: string[],
): Record<string, unknown> {
  if (exclude.length === 0) return state;
  const filtered = { ...state };
  for (const key of exclude) {
    delete filtered[key];
  }
  return filtered;
}

export const serverPersistPlugin: PiniaPlugin = ({ store, options }) => {
  if (!options.persist) return;

  const storeId = store.$id;
  const excludeKeys = options.persistExclude ?? [];

  // 从服务器加载持久化数据
  getStoreConfig(storeId)
    .then((savedState) => {
      if (savedState && typeof savedState === "object") {
        store.$patch(savedState as StateTree);
        // 恢复后调用 onRestore 钩子（如有），供各 store 处理恢复后清理
        if (typeof store.onRestore === "function") {
          store.onRestore();
        }
      }
      // 加载完成后保存一次完整 state（确保新增字段的默认值也被持久化）
      const fullState = filterState(
        JSON.parse(JSON.stringify(store.$state)),
        excludeKeys,
      );
      debouncedSave(storeId, fullState);
    })
    .catch((err) => {
      console.error(`[PersistPlugin] Failed to load store "${storeId}":`, err);
    });

  // 监听 state 变化并保存到服务器（防抖）
  watch(
    () => store.$state,
    (newState) => {
      const filtered = filterState(
        JSON.parse(JSON.stringify(newState)),
        excludeKeys,
      );
      debouncedSave(storeId, filtered);
    },
    { deep: true },
  );
};
