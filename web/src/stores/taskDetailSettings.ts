import { ref } from "vue";
import { defineStore } from "pinia";

export const useTaskDetailSettingsStore = defineStore(
  "taskDetailSettings",
  () => {
    const showRecoId = ref(true);
    const showActionId = ref(true);

    function setShowRecoId(value: boolean) {
      showRecoId.value = value;
    }

    function setShowActionId(value: boolean) {
      showActionId.value = value;
    }

    function reset() {
      showRecoId.value = true;
      showActionId.value = true;
    }

    return {
      showRecoId,
      showActionId,
      setShowRecoId,
      setShowActionId,
      reset,
    };
  },
  { persist: true },
);
