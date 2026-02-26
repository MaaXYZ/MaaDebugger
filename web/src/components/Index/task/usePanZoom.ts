import { computed, ref } from "vue";

export const MIN_ZOOM = 0.1;
export const MAX_ZOOM = 5;
const ZOOM_STEP = 0.15;

export function usePanZoom() {
  const aspectMode = ref<"landscape" | "portrait">("landscape");
  const zoomLevel = ref(1);
  const isFullscreen = ref(false);

  const isDragging = ref(false);
  const dragStart = ref({ x: 0, y: 0 });
  const dragOffset = ref({ x: 0, y: 0 });
  const panOffset = ref({ x: 0, y: 0 });

  const fullscreenZoom = ref(1);
  const isFullscreenDragging = ref(false);
  const fullscreenDragStart = ref({ x: 0, y: 0 });
  const fullscreenDragOffset = ref({ x: 0, y: 0 });
  const fullscreenPanOffset = ref({ x: 0, y: 0 });

  const containerStyle = computed(() => {
    const ratio = aspectMode.value === "landscape" ? "16/9" : "9/16";
    return {
      aspectRatio: ratio,
      maxHeight: aspectMode.value === "portrait" ? "70vh" : undefined,
    };
  });

  const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100));
  const fullscreenZoomPercentage = computed(() =>
    Math.round(fullscreenZoom.value * 100),
  );

  const imageStyle = computed(() => ({
    transform: `translate(${panOffset.value.x}px, ${panOffset.value.y}px) scale(${zoomLevel.value})`,
    transformOrigin: "center center",
    transition: isDragging.value ? "none" : "transform 0.2s ease",
  }));

  const fullscreenImageStyle = computed(() => ({
    transform: `translate(${fullscreenPanOffset.value.x}px, ${fullscreenPanOffset.value.y}px) scale(${fullscreenZoom.value})`,
    transformOrigin: "center center",
    transition: isFullscreenDragging.value ? "none" : "transform 0.2s ease",
    maxWidth: "90vw",
    maxHeight: "85vh",
    objectFit: "contain" as const,
  }));

  function toggleAspect() {
    aspectMode.value =
      aspectMode.value === "landscape" ? "portrait" : "landscape";
    resetZoom();
  }

  function zoomIn() {
    zoomLevel.value = Math.min(MAX_ZOOM, zoomLevel.value + ZOOM_STEP);
  }

  function zoomOut() {
    zoomLevel.value = Math.max(MIN_ZOOM, zoomLevel.value - ZOOM_STEP);
    if (zoomLevel.value <= 1) panOffset.value = { x: 0, y: 0 };
  }

  function resetZoom() {
    zoomLevel.value = 1;
    panOffset.value = { x: 0, y: 0 };
  }

  function onWheel(e: WheelEvent, hasImage: boolean) {
    if (!hasImage) return;
    if (e.deltaY < 0) zoomIn();
    else zoomOut();
  }

  function onDragStart(e: MouseEvent) {
    if (zoomLevel.value <= 1) return;
    isDragging.value = true;
    dragStart.value = { x: e.clientX, y: e.clientY };
    dragOffset.value = { ...panOffset.value };
  }

  function onDragMove(e: MouseEvent) {
    if (!isDragging.value) return;
    panOffset.value = {
      x: dragOffset.value.x + (e.clientX - dragStart.value.x),
      y: dragOffset.value.y + (e.clientY - dragStart.value.y),
    };
  }

  function onDragEnd() {
    isDragging.value = false;
  }

  function fullscreenZoomIn() {
    fullscreenZoom.value = Math.min(MAX_ZOOM, fullscreenZoom.value + ZOOM_STEP);
  }

  function fullscreenZoomOut() {
    fullscreenZoom.value = Math.max(MIN_ZOOM, fullscreenZoom.value - ZOOM_STEP);
    if (fullscreenZoom.value <= 1) fullscreenPanOffset.value = { x: 0, y: 0 };
  }

  function resetFullscreenZoom() {
    fullscreenZoom.value = 1;
    fullscreenPanOffset.value = { x: 0, y: 0 };
  }

  function onFullscreenWheel(e: WheelEvent) {
    if (e.deltaY < 0) fullscreenZoomIn();
    else fullscreenZoomOut();
  }

  function onFullscreenDragStart(e: MouseEvent) {
    if (fullscreenZoom.value <= 1) return;
    isFullscreenDragging.value = true;
    fullscreenDragStart.value = { x: e.clientX, y: e.clientY };
    fullscreenDragOffset.value = { ...fullscreenPanOffset.value };
  }

  function onFullscreenDragMove(e: MouseEvent) {
    if (!isFullscreenDragging.value) return;
    fullscreenPanOffset.value = {
      x:
        fullscreenDragOffset.value.x +
        (e.clientX - fullscreenDragStart.value.x),
      y:
        fullscreenDragOffset.value.y +
        (e.clientY - fullscreenDragStart.value.y),
    };
  }

  function onFullscreenDragEnd() {
    isFullscreenDragging.value = false;
  }

  function handleFullscreenChange(open: boolean) {
    if (!open) {
      resetFullscreenZoom();
    }
  }

  return {
    aspectMode,
    zoomLevel,
    isFullscreen,
    isDragging,
    fullscreenZoom,
    isFullscreenDragging,
    containerStyle,
    zoomPercentage,
    fullscreenZoomPercentage,
    imageStyle,
    fullscreenImageStyle,
    toggleAspect,
    zoomIn,
    zoomOut,
    resetZoom,
    onWheel,
    onDragStart,
    onDragMove,
    onDragEnd,
    fullscreenZoomIn,
    fullscreenZoomOut,
    resetFullscreenZoom,
    onFullscreenWheel,
    onFullscreenDragStart,
    onFullscreenDragMove,
    onFullscreenDragEnd,
    handleFullscreenChange,
  };
}
