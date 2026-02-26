import { computed, onUnmounted, ref, watch } from "vue";
import {
  getScreenshotStatus,
  pauseScreenshot,
  resumeScreenshot,
  setScreenshotFPS,
  startScreenshot,
  stopScreenshot,
} from "@/api/http";
import {
  latestFrame,
  screenshotError,
  screenshotFps,
  screenshotPaused,
  screenshotRunning,
} from "@/stores/screenshot";

export function useScreenshotStream() {
  const imageData = ref<ArrayBuffer | null>(null);
  const imageUrl = ref<string | null>(null);
  const fpsSlider = ref(30);

  let pendingFrame: ArrayBuffer | null = null;
  let rafId = 0;

  const isStreaming = computed(() => screenshotRunning.value);
  const isPaused = computed(() => screenshotPaused.value);
  const currentFps = computed(() => screenshotFps.value);

  async function toggleStreaming() {
    if (isStreaming.value) {
      await stopScreenshot();
      screenshotRunning.value = false;
    } else {
      screenshotError.value = "";
      await startScreenshot();
      screenshotRunning.value = true;
      screenshotPaused.value = false;
    }
  }

  async function togglePause() {
    if (isPaused.value) {
      await resumeScreenshot();
      screenshotPaused.value = false;
    } else {
      await pauseScreenshot();
      screenshotPaused.value = true;
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

  async function initStatus() {
    const status = await getScreenshotStatus();
    if (status) {
      screenshotRunning.value = status.running;
      screenshotPaused.value = status.paused;
      screenshotFps.value = status.fps;
      fpsSlider.value = status.fps;
    }
  }

  watch(latestFrame, (frame) => {
    if (!frame) {
      imageData.value = null;
      return;
    }

    pendingFrame = frame;
    if (!rafId) {
      rafId = requestAnimationFrame(flushFrame);
    }
  });

  watch(imageData, () => {
    updateImageUrl();
  });

  onUnmounted(() => {
    if (rafId) cancelAnimationFrame(rafId);
    if (imageUrl.value) URL.revokeObjectURL(imageUrl.value);
  });

  return {
    imageData,
    imageUrl,
    fpsSlider,
    isStreaming,
    isPaused,
    currentFps,
    screenshotError,
    toggleStreaming,
    togglePause,
    applyFps,
    downloadImage,
    initStatus,
  };
}
