import { ref } from "vue";
import { defineStore } from "pinia";

const DEFAULT_NODE_PAGE_SIZE = 20;

export const useTaskDetailSettingsStore = defineStore(
  "taskDetailSettings",
  () => {
    const showRecoId = ref(true);
    const showActionId = ref(true);
    const reverseNodeOrder = ref(true);
    const nodePageSize = ref(DEFAULT_NODE_PAGE_SIZE);

    function setShowRecoId(value: boolean) {
      showRecoId.value = value;
    }

    function setShowActionId(value: boolean) {
      showActionId.value = value;
    }

    function setReverseNodeOrder(value: boolean) {
      reverseNodeOrder.value = value;
    }

    function setNodePageSize(value: number) {
      const normalized = Number.isFinite(value)
        ? Math.max(1, Math.floor(value))
        : DEFAULT_NODE_PAGE_SIZE;
      nodePageSize.value = normalized;
    }

    function reset() {
      showRecoId.value = true;
      showActionId.value = true;
      reverseNodeOrder.value = true;
      nodePageSize.value = DEFAULT_NODE_PAGE_SIZE;
    }

    return {
      showRecoId,
      showActionId,
      reverseNodeOrder,
      nodePageSize,
      setShowRecoId,
      setShowActionId,
      setReverseNodeOrder,
      setNodePageSize,
      reset,
    };
  },
  { persist: true },
);
