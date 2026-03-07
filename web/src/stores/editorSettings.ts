import { computed, ref } from "vue";
import { defineStore } from "pinia";

const MIN_FONT_SIZE = 12;
const MAX_FONT_SIZE = 24;
const MIN_EDITOR_HEIGHT = 360;
const MAX_EDITOR_HEIGHT = 1200;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export const useEditorSettingsStore = defineStore(
  "editorSettings",
  () => {
    const fontSize = ref(15);
    const minHeight = ref(420);
    const maxHeight = ref(720);

    const normalizedFontSize = computed(() =>
      clamp(fontSize.value, MIN_FONT_SIZE, MAX_FONT_SIZE),
    );

    const normalizedMinHeight = computed(() =>
      clamp(minHeight.value, MIN_EDITOR_HEIGHT, MAX_EDITOR_HEIGHT),
    );

    const normalizedMaxHeight = computed(() => {
      const nextMaxHeight = clamp(
        maxHeight.value,
        MIN_EDITOR_HEIGHT,
        MAX_EDITOR_HEIGHT,
      );
      return Math.max(nextMaxHeight, normalizedMinHeight.value);
    });

    function setFontSize(value: number) {
      fontSize.value = clamp(value, MIN_FONT_SIZE, MAX_FONT_SIZE);
    }

    function setMinHeight(value: number) {
      minHeight.value = clamp(value, MIN_EDITOR_HEIGHT, MAX_EDITOR_HEIGHT);
      if (maxHeight.value < minHeight.value) {
        maxHeight.value = minHeight.value;
      }
    }

    function setMaxHeight(value: number) {
      maxHeight.value = clamp(value, MIN_EDITOR_HEIGHT, MAX_EDITOR_HEIGHT);
      if (maxHeight.value < minHeight.value) {
        minHeight.value = maxHeight.value;
      }
    }

    function reset() {
      fontSize.value = 15;
      minHeight.value = 420;
      maxHeight.value = 720;
    }

    return {
      fontSize,
      minHeight,
      maxHeight,
      normalizedFontSize,
      normalizedMinHeight,
      normalizedMaxHeight,
      setFontSize,
      setMinHeight,
      setMaxHeight,
      reset,
    };
  },
  { persist: true },
);
