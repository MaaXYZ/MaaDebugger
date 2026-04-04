import { defineStore } from "pinia";
import { ref } from "vue";

export const SIGNAL_TYPES = {
  NULL: "NULL",
  REFRESH_NODE: "REFRESH_NODE",
} as const;
export type SignalType = keyof typeof SIGNAL_TYPES;

export const useSignalStore = defineStore("signal", () => {
  const reloadResource = ref(0);
  const refreshNode = ref(0);

  function emitReloadResource() {
    reloadResource.value++;
  }
  function emitRefreshNode() {
    refreshNode.value++;
  }
  return {
    reloadResource,
    refreshNode,
    emitRefreshNode,
    emitReloadResource,
  };
});
