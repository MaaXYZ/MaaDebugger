import { ref, shallowRef } from "vue";

export const latestFrame = shallowRef<ArrayBuffer | null>(null);
export const screenshotRunning = ref(false);
export const screenshotPaused = ref(false);
export const screenshotFps = ref(15);
export const screenshotActualFps = ref(0);
export const screenshotError = ref("");
