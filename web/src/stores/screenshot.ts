import { ref, shallowRef } from "vue";

export const DEFAULT_SCREENSHOT_FPS = 15;

export type ScreenshotOverlayState =
  | "none"
  | "disconnected"
  | "paused"
  | "failed";

export const latestFrame = shallowRef<ArrayBuffer | null>(null);
export const screenshotRunning = ref(false);
export const screenshotPaused = ref(false);
export const screenshotFps = ref(DEFAULT_SCREENSHOT_FPS);
export const screenshotActualFps = ref(0);
export const screenshotError = ref("");
export const screenshotOverlayState = ref<ScreenshotOverlayState>("none");
export const screenshotOverlayMessage = ref("");
