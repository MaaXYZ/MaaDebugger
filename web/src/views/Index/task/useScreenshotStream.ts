import { computed, onUnmounted, ref, watch } from "vue";
import {
  getScreenshotStatus,
  pauseScreenshot,
  resumeScreenshot,
  setScreenshotFPS,
} from "@/api/http";
import {
  DEFAULT_SCREENSHOT_FPS,
  latestFrame,
  screenshotActualFps,
  screenshotError,
  screenshotFps,
  screenshotOverlayMessage,
  screenshotOverlayState,
  screenshotPaused,
  screenshotRunning,
} from "@/stores/screenshot";

export function useScreenshotStream() {
  const imageData = ref<ArrayBuffer | null>(null);
  const imageUrl = ref<string | null>(null);
  const fpsSlider = ref(screenshotFps.value || DEFAULT_SCREENSHOT_FPS);

  let pendingFrame: ArrayBuffer | null = null;
  let rafId = 0;
  let fpsWindowStart = 0;
  let fpsFrameCount = 0;

  const isPaused = computed(() => screenshotPaused.value);
  const currentFps = computed(() => screenshotFps.value);
  const actualFps = computed(() => screenshotActualFps.value);

  async function togglePause() {
    if (isPaused.value) {
      screenshotPaused.value = false;
      screenshotOverlayState.value = "none";
      screenshotOverlayMessage.value = "";
      await resumeScreenshot();
    } else {
      screenshotPaused.value = true;
      screenshotOverlayState.value = "paused";
      screenshotOverlayMessage.value = "Screenshot paused";
      await pauseScreenshot();
    }
  }

  async function applyFps() {
    const result = await setScreenshotFPS(fpsSlider.value);
    if (result.succeed && result.data) {
      screenshotFps.value = result.data.fps;
      fpsSlider.value = result.data.fps;
    }
  }

  function flushFrame() {
    rafId = 0;
    if (pendingFrame) {
      imageData.value = pendingFrame;
      pendingFrame = null;
    }
  }

  function updateImageUrl() {
    if (imageUrl.value) {
      URL.revokeObjectURL(imageUrl.value);
      imageUrl.value = null;
    }
    if (!imageData.value) return;
    const blob = new Blob([new Uint8Array(imageData.value)], {
      type: "image/jpeg",
    });
    imageUrl.value = URL.createObjectURL(blob);
  }

  function downloadImage() {
    if (!imageData.value) return;
    const blob = new Blob([new Uint8Array(imageData.value)], {
      type: "image/jpeg",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    const now = new Date();
    const timestamp =
      now.getFullYear().toString() +
      String(now.getMonth() + 1).padStart(2, "0") +
      String(now.getDate()).padStart(2, "0") +
      "_" +
      String(now.getHours()).padStart(2, "0") +
      String(now.getMinutes()).padStart(2, "0") +
      String(now.getSeconds()).padStart(2, "0");

    a.href = url;
    a.download = `screenshot_${timestamp}.jpg`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  function recordFrame() {
    const now = performance.now();
    if (!fpsWindowStart) {
      fpsWindowStart = now;
      fpsFrameCount = 1;
      screenshotActualFps.value = 0;
      return;
    }

    fpsFrameCount += 1;
    const elapsed = now - fpsWindowStart;
    if (elapsed >= 1000) {
      screenshotActualFps.value = Number(
        ((fpsFrameCount * 1000) / elapsed).toFixed(1),
      );
      fpsWindowStart = now;
      fpsFrameCount = 0;
    }
  }

  async function initStatus() {
    const status = await getScreenshotStatus();
    if (status) {
      screenshotRunning.value = status.running;
      screenshotPaused.value = status.paused;
      screenshotFps.value = status.fps;
      screenshotOverlayState.value = status.overlay_state;
      screenshotOverlayMessage.value = status.overlay_message;
      fpsSlider.value = status.fps;
    }
    if (!status?.running) {
      screenshotActualFps.value = 0;
    }
  }

  watch(latestFrame, (frame) => {
    if (!frame) {
      pendingFrame = null;
      imageData.value = null;
      return;
    }

    recordFrame();
    pendingFrame = frame;
    if (!rafId) {
      rafId = requestAnimationFrame(flushFrame);
    }
  });

  watch(screenshotOverlayState, (state) => {
    if (state === "paused" || state === "disconnected" || state === "failed") {
      pendingFrame = null;
      imageData.value = null;
      screenshotActualFps.value = 0;
    }
  });

  watch(imageData, () => {
    updateImageUrl();
  });

  watch(screenshotFps, (fps, previousFps) => {
    if (fpsSlider.value === previousFps) {
      fpsSlider.value = fps;
    }
  });

  onUnmounted(() => {
    if (rafId) cancelAnimationFrame(rafId);
    if (imageUrl.value) URL.revokeObjectURL(imageUrl.value);
    screenshotActualFps.value = 0;
  });

  return {
    imageData,
    imageUrl,
    fpsSlider,
    isPaused,
    currentFps,
    actualFps,
    screenshotError,
    screenshotOverlayState,
    screenshotOverlayMessage,
    togglePause,
    applyFps,
    downloadImage,
    initStatus,
  };
}
