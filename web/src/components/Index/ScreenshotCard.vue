<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Screenshot</span>
                <div class="flex-1" />
                <!-- Aspect ratio toggle -->
                <UTooltip :text="aspectMode === 'landscape' ? 'Switch to 9:16 portrait' : 'Switch to 16:9 landscape'">
                    <UButton color="neutral" variant="ghost" :icon="aspectMode === 'landscape'
                        ? 'i-lucide-monitor'
                        : 'i-lucide-smartphone'" size="sm" @click="toggleAspect" />
                </UTooltip>
            </div>
        </template>

        <template #default>
            <!-- Image display area -->
            <div ref="containerRef" class="relative overflow-hidden rounded-md border border-default bg-muted"
                :style="containerStyle" @wheel.prevent="onWheel">
                <!-- Has image -->
                <div v-if="imageUrl" class="absolute inset-0 flex items-center justify-center cursor-grab select-none"
                    :class="{ 'cursor-grabbing': isDragging }" @mousedown="onDragStart" @mousemove="onDragMove"
                    @mouseup="onDragEnd" @mouseleave="onDragEnd">
                    <img ref="imgRef" :src="imageUrl" alt="Screenshot" draggable="false"
                        class="pointer-events-none max-w-none" :style="imageStyle" />
                </div>
                <!-- No image placeholder -->
                <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                    <UIcon name="i-lucide-image" class="size-12" />
                    <span class="text-sm">No screenshot available</span>
                </div>
            </div>
        </template>

        <template #footer>
            <div class="flex flex-row items-center gap-2">
                <!-- Zoom controls -->
                <div class="flex items-center gap-1">
                    <UTooltip text="Zoom out">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="sm"
                            :disabled="!imageUrl || zoomLevel <= MIN_ZOOM" @click="zoomOut" />
                    </UTooltip>
                    <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                        {{ zoomPercentage }}%
                    </span>
                    <UTooltip text="Zoom in">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="sm"
                            :disabled="!imageUrl || zoomLevel >= MAX_ZOOM" @click="zoomIn" />
                    </UTooltip>
                </div>

                <USeparator orientation="vertical" class="h-5" />

                <!-- Fit to container -->
                <UTooltip text="Fit to view">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm" :disabled="!imageUrl"
                        @click="resetZoom" />
                </UTooltip>

                <div class="flex-1" />

                <!-- Fullscreen -->
                <UTooltip text="Fullscreen">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-fullscreen" size="sm" :disabled="!imageUrl"
                        @click="isFullscreen = true" />
                </UTooltip>

                <!-- Download -->
                <UTooltip text="Download PNG">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm" :disabled="!imageUrl"
                        @click="downloadImage" />
                </UTooltip>
            </div>
        </template>
    </UCard>

    <!-- Fullscreen modal -->
    <UModal v-model:open="isFullscreen" title="Screenshot" fullscreen>
        <template #body>
            <div class="relative w-full h-full flex items-center justify-center overflow-hidden bg-muted"
                @wheel.prevent="onFullscreenWheel">
                <div class="flex items-center justify-center cursor-grab select-none"
                    :class="{ 'cursor-grabbing': isFullscreenDragging }" @mousedown="onFullscreenDragStart"
                    @mousemove="onFullscreenDragMove" @mouseup="onFullscreenDragEnd" @mouseleave="onFullscreenDragEnd">
                    <img v-if="imageUrl" :src="imageUrl" alt="Screenshot" draggable="false"
                        class="pointer-events-none max-w-none" :style="fullscreenImageStyle" />
                </div>
                <!-- Fullscreen toolbar -->
                <div
                    class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-elevated/90 backdrop-blur-sm rounded-lg px-3 py-2 border border-default shadow-lg">
                    <UTooltip text="Zoom out">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="sm"
                            :disabled="fullscreenZoom <= MIN_ZOOM" @click="fullscreenZoomOut" />
                    </UTooltip>
                    <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                        {{ fullscreenZoomPercentage }}%
                    </span>
                    <UTooltip text="Zoom in">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="sm"
                            :disabled="fullscreenZoom >= MAX_ZOOM" @click="fullscreenZoomIn" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Fit to view">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm"
                            @click="resetFullscreenZoom" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Download PNG">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm"
                            @click="downloadImage" />
                    </UTooltip>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

// --- Constants ---
const MIN_ZOOM = 0.1
const MAX_ZOOM = 5
const ZOOM_STEP = 0.15

// --- State ---
const imageData = ref<Uint8Array | ArrayBuffer | null>(null)
const imageUrl = ref<string | null>(null)
const aspectMode = ref<'landscape' | 'portrait'>('landscape')
const zoomLevel = ref(1)
const isFullscreen = ref(false)

// Drag state (card view)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const panOffset = ref({ x: 0, y: 0 })

// Fullscreen zoom & drag
const fullscreenZoom = ref(1)
const isFullscreenDragging = ref(false)
const fullscreenDragStart = ref({ x: 0, y: 0 })
const fullscreenDragOffset = ref({ x: 0, y: 0 })
const fullscreenPanOffset = ref({ x: 0, y: 0 })

// --- Computed ---
const containerStyle = computed(() => {
    const ratio = aspectMode.value === 'landscape' ? '16/9' : '9/16'
    return {
        aspectRatio: ratio,
        maxHeight: aspectMode.value === 'portrait' ? '70vh' : undefined,
    }
})

const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))
const fullscreenZoomPercentage = computed(() => Math.round(fullscreenZoom.value * 100))

const imageStyle = computed(() => ({
    transform: `translate(${panOffset.value.x}px, ${panOffset.value.y}px) scale(${zoomLevel.value})`,
    transformOrigin: 'center center',
    transition: isDragging.value ? 'none' : 'transform 0.2s ease',
}))

const fullscreenImageStyle = computed(() => ({
    transform: `translate(${fullscreenPanOffset.value.x}px, ${fullscreenPanOffset.value.y}px) scale(${fullscreenZoom.value})`,
    transformOrigin: 'center center',
    transition: isFullscreenDragging.value ? 'none' : 'transform 0.2s ease',
    maxWidth: '90vw',
    maxHeight: '85vh',
    objectFit: 'contain' as const,
}))

// --- Aspect toggle ---
function toggleAspect() {
    aspectMode.value = aspectMode.value === 'landscape' ? 'portrait' : 'landscape'
    resetZoom()
}

// --- Zoom (card view) ---
function zoomIn() {
    zoomLevel.value = Math.min(MAX_ZOOM, zoomLevel.value + ZOOM_STEP)
}

function zoomOut() {
    zoomLevel.value = Math.max(MIN_ZOOM, zoomLevel.value - ZOOM_STEP)
    // Reset pan if zoomed back to 1 or below
    if (zoomLevel.value <= 1) {
        panOffset.value = { x: 0, y: 0 }
    }
}

function resetZoom() {
    zoomLevel.value = 1
    panOffset.value = { x: 0, y: 0 }
}

function onWheel(e: WheelEvent) {
    if (!imageUrl.value) return
    if (e.deltaY < 0) {
        zoomIn()
    } else {
        zoomOut()
    }
}

// --- Drag (card view) ---
function onDragStart(e: MouseEvent) {
    if (zoomLevel.value <= 1) return
    isDragging.value = true
    dragStart.value = { x: e.clientX, y: e.clientY }
    dragOffset.value = { ...panOffset.value }
}

function onDragMove(e: MouseEvent) {
    if (!isDragging.value) return
    panOffset.value = {
        x: dragOffset.value.x + (e.clientX - dragStart.value.x),
        y: dragOffset.value.y + (e.clientY - dragStart.value.y),
    }
}

function onDragEnd() {
    isDragging.value = false
}

// --- Zoom (fullscreen) ---
function fullscreenZoomIn() {
    fullscreenZoom.value = Math.min(MAX_ZOOM, fullscreenZoom.value + ZOOM_STEP)
}

function fullscreenZoomOut() {
    fullscreenZoom.value = Math.max(MIN_ZOOM, fullscreenZoom.value - ZOOM_STEP)
    if (fullscreenZoom.value <= 1) {
        fullscreenPanOffset.value = { x: 0, y: 0 }
    }
}

function resetFullscreenZoom() {
    fullscreenZoom.value = 1
    fullscreenPanOffset.value = { x: 0, y: 0 }
}

function onFullscreenWheel(e: WheelEvent) {
    if (e.deltaY < 0) {
        fullscreenZoomIn()
    } else {
        fullscreenZoomOut()
    }
}

// --- Drag (fullscreen) ---
function onFullscreenDragStart(e: MouseEvent) {
    if (fullscreenZoom.value <= 1) return
    isFullscreenDragging.value = true
    fullscreenDragStart.value = { x: e.clientX, y: e.clientY }
    fullscreenDragOffset.value = { ...fullscreenPanOffset.value }
}

function onFullscreenDragMove(e: MouseEvent) {
    if (!isFullscreenDragging.value) return
    fullscreenPanOffset.value = {
        x: fullscreenDragOffset.value.x + (e.clientX - fullscreenDragStart.value.x),
        y: fullscreenDragOffset.value.y + (e.clientY - fullscreenDragStart.value.y),
    }
}

function onFullscreenDragEnd() {
    isFullscreenDragging.value = false
}

// --- Download ---
function downloadImage() {
    if (!imageData.value) return

    const blob = new Blob(
        [imageData.value instanceof ArrayBuffer ? new Uint8Array(imageData.value as ArrayBuffer) : imageData.value as BlobPart],
        { type: 'image/png' }
    )
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const now = new Date()
    const timestamp = now.getFullYear().toString()
        + String(now.getMonth() + 1).padStart(2, '0')
        + String(now.getDate()).padStart(2, '0')
        + '_'
        + String(now.getHours()).padStart(2, '0')
        + String(now.getMinutes()).padStart(2, '0')
        + String(now.getSeconds()).padStart(2, '0')
    a.href = url
    a.download = `screenshot_${timestamp}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
}

// --- Image data management ---
function updateImageUrl() {
    // Revoke previous URL
    if (imageUrl.value) {
        URL.revokeObjectURL(imageUrl.value)
        imageUrl.value = null
    }

    if (!imageData.value) return

    const blob = new Blob(
        [imageData.value instanceof ArrayBuffer ? new Uint8Array(imageData.value as ArrayBuffer) : imageData.value as BlobPart],
        { type: 'image/png' }
    )
    imageUrl.value = URL.createObjectURL(blob)
}

watch(imageData, () => {
    updateImageUrl()
    resetZoom()
})

// Reset fullscreen zoom when modal closes
watch(isFullscreen, (val) => {
    if (!val) {
        resetFullscreenZoom()
    }
})

// --- Mock data generation ---
function generateMockImage(): Uint8Array {
    const width = aspectMode.value === 'landscape' ? 1080 : 720
    const height = aspectMode.value === 'landscape' ? 720 : 1080

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')!

    // Background gradient
    const gradient = ctx.createLinearGradient(0, 0, width, height)
    gradient.addColorStop(0, '#667eea')
    gradient.addColorStop(0.5, '#764ba2')
    gradient.addColorStop(1, '#f093fb')
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, width, height)

    // Grid lines
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)'
    ctx.lineWidth = 1
    const gridSize = 60
    for (let x = 0; x <= width; x += gridSize) {
        ctx.beginPath()
        ctx.moveTo(x, 0)
        ctx.lineTo(x, height)
        ctx.stroke()
    }
    for (let y = 0; y <= height; y += gridSize) {
        ctx.beginPath()
        ctx.moveTo(0, y)
        ctx.lineTo(width, y)
        ctx.stroke()
    }

    // Center text
    ctx.fillStyle = 'rgba(255, 255, 255, 0.9)'
    ctx.font = 'bold 32px system-ui, sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(`Mock Screenshot`, width / 2, height / 2 - 24)

    ctx.font = '20px system-ui, sans-serif'
    ctx.fillStyle = 'rgba(255, 255, 255, 0.7)'
    ctx.fillText(`${width} × ${height}`, width / 2, height / 2 + 16)

    // Corner markers
    const markerSize = 20
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.5)'
    ctx.lineWidth = 2
    // Top-left
    ctx.beginPath()
    ctx.moveTo(10, 10 + markerSize)
    ctx.lineTo(10, 10)
    ctx.lineTo(10 + markerSize, 10)
    ctx.stroke()
    // Top-right
    ctx.beginPath()
    ctx.moveTo(width - 10 - markerSize, 10)
    ctx.lineTo(width - 10, 10)
    ctx.lineTo(width - 10, 10 + markerSize)
    ctx.stroke()
    // Bottom-left
    ctx.beginPath()
    ctx.moveTo(10, height - 10 - markerSize)
    ctx.lineTo(10, height - 10)
    ctx.lineTo(10 + markerSize, height - 10)
    ctx.stroke()
    // Bottom-right
    ctx.beginPath()
    ctx.moveTo(width - 10 - markerSize, height - 10)
    ctx.lineTo(width - 10, height - 10)
    ctx.lineTo(width - 10, height - 10 - markerSize)
    ctx.stroke()

    // Convert canvas to PNG binary
    const dataUrl = canvas.toDataURL('image/png')
    const base64 = dataUrl.split(',')[1] ?? ''
    const binaryStr = atob(base64)
    const bytes = new Uint8Array(binaryStr.length)
    for (let i = 0; i < binaryStr.length; i++) {
        bytes[i] = binaryStr.charCodeAt(i)
    }
    return bytes
}

// --- Public API ---
function setImageData(data: Uint8Array | ArrayBuffer | null) {
    imageData.value = data
}

function clearImage() {
    imageData.value = null
}

defineExpose({
    setImageData,
    clearImage,
    imageData,
    aspectMode,
})

// --- Lifecycle ---
onMounted(() => {
    // Load mock data on mount for development preview
    imageData.value = generateMockImage()
})

onUnmounted(() => {
    if (imageUrl.value) {
        URL.revokeObjectURL(imageUrl.value)
    }
})
</script>
