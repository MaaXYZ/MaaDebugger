<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task</span>
                <TaskStatusBadge :status="taskStatus" />
                <div class="flex-1"></div>

                <!-- FPS control -->
                <UPopover>
                    <UButton color="neutral" variant="ghost" size="sm" class="tabular-nums">
                        {{ currentFps }} FPS
                    </UButton>
                    <template #content>
                        <div class="p-3 flex flex-col gap-2 w-48">
                            <span class="text-xs text-muted">Frame Rate</span>
                            <USlider v-model="fpsSlider" :min="1" :max="60" :step="1" />
                            <div class="flex items-center justify-between">
                                <span class="text-xs text-muted tabular-nums">{{ fpsSlider }} FPS</span>
                                <UButton size="xs" @click="applyFps">Apply</UButton>
                            </div>
                        </div>
                    </template>
                </UPopover>

                <!-- Pause/Resume -->
                <UTooltip :text="isPaused ? 'Resume' : 'Pause'">
                    <UButton color="neutral" variant="ghost" :icon="isPaused ? 'i-lucide-play' : 'i-lucide-pause'"
                             size="sm" :disabled="!isStreaming" @click="togglePause" />
                </UTooltip>

                <!-- Start/Stop streaming -->
                <UTooltip :text="isStreaming ? 'Stop streaming' : 'Start streaming'">
                    <UButton :color="isStreaming ? 'error' : 'primary'" variant="ghost"
                             :icon="isStreaming ? 'i-lucide-video-off' : 'i-lucide-video'" size="sm"
                             @click="toggleStreaming" />
                </UTooltip>

                <!-- Aspect ratio toggle -->
                <UTooltip
                    :text="aspectMode === 'landscape' ? 'Switch to 9:16 portrait' : 'Switch to 16:9 landscape'">
                    <UButton color="neutral" variant="ghost" :icon="aspectMode === 'landscape'
                        ? 'i-lucide-monitor'
                        : 'i-lucide-smartphone'" size="sm" @click="toggleAspect" />
                </UTooltip>
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <!-- Task controls -->
                <div class="flex flex-row items-center gap-2">
                    <div class="flex-1 min-w-0">
                        <USelectMenu v-model="selectedEntry" :items="entrySelectItems"
                            placeholder="Select task entry..." :search-input="{
                                placeholder: 'Filter...',
                                icon: 'i-lucide-search'
                            }" :ui="{ base: 'w-full', content: 'w-auto min-w-(--reka-combobox-trigger-width)' }" class="w-full" size="xl" value-key="value"
                            :disabled="isRunning" arrow />
                    </div>
                    <UTooltip text="Edit task override">
                        <UButton color="neutral" variant="outline" icon="i-lucide-file-edit" size="xl"
                                 @click="onEditOverride" />
                    </UTooltip>
                    <UButton v-if="!isRunning" color="success" variant="soft" icon="i-lucide-play" size="xl"
                             :disabled="!canStart" @click="onStart">
                        <template v-if="startStopKeys.length" #trailing>
                            <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                        </template>
                    </UButton>
                    <UButton v-else color="error" variant="soft" icon="i-lucide-square" size="xl"
                             :loading="isStopping" :disabled="isStopping" @click="onStop">
                        <template v-if="startStopKeys.length" #trailing>
                            <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                        </template>
                    </UButton>
                </div>

                <!-- Screenshot area -->
                <div class="relative overflow-hidden rounded-md border border-default bg-muted"
                     :style="containerStyle" @wheel.prevent="onWheel">
                    <div v-if="imageUrl" class="absolute inset-0 flex items-center justify-center cursor-grab select-none"
                         :class="{ 'cursor-grabbing': isDragging }" @mousedown="onDragStart" @mousemove="onDragMove"
                         @mouseup="onDragEnd" @mouseleave="onDragEnd">
                        <img :src="imageUrl" alt="Screenshot" draggable="false"
                             class="pointer-events-none w-full h-full object-contain" :style="imageStyle" />
                    </div>
                    <div v-else-if="screenshotError" class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-error">
                        <UIcon name="i-lucide-circle-x" class="size-12" />
                        <span class="text-sm font-medium">Screenshot Failed</span>
                        <span class="text-xs text-dimmed max-w-xs text-center">{{ screenshotError }}</span>
                    </div>
                    <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                        <UIcon name="i-lucide-image" class="size-12" />
                        <span class="text-sm">No screenshot available</span>
                    </div>
                </div>
            </div>
        </template>

        <template #footer>
            <div class="flex flex-row items-center gap-2">
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

                <UTooltip text="Fit to view">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm" :disabled="!imageUrl"
                             @click="resetZoom" />
                </UTooltip>

                <div class="flex-1"></div>

                <UTooltip text="Fullscreen">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-fullscreen" size="sm" :disabled="!imageUrl"
                             @click="isFullscreen = true" />
                </UTooltip>

                <UTooltip text="Download">
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
                    <UTooltip text="Download">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm"
                                 @click="downloadImage" />
                    </UTooltip>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import TaskStatusBadge from './task/TaskStatusBadge.vue'
import type { TaskStatus } from './task/types'
import { useShortcutsStore, formatShortcut } from '@/stores/shortcuts'
import { useStatusStore } from '@/stores/status'
import {
    getTaskNodes, runTask, stopTask,
    startScreenshot, stopScreenshot,
    pauseScreenshot, resumeScreenshot,
    setScreenshotFPS, getScreenshotStatus,
} from '@/api/http'
import {
    latestFrame,
    screenshotRunning,
    screenshotPaused,
    screenshotFps,
    screenshotError,
} from '@/stores/screenshot'

// ===================== Task =====================

interface TaskEntry {
    label: string
    value: string
}

const toast = useToast()
const shortcutsStore = useShortcutsStore()
const statusStore = useStatusStore()
const entries = ref<TaskEntry[]>([])
const selectedEntry = ref<string>('')
const taskStatus = computed<TaskStatus>(() => statusStore.taskStatus)

const entrySelectItems = computed(() =>
    entries.value.map(e => ({ label: e.label, value: e.value }))
)

const isRunning = computed(() => taskStatus.value === 'running')

const canStart = computed(() =>
    statusStore.controllerStatus === 'connected'
    && statusStore.resourceStatus === 'loaded'
    && !!selectedEntry.value
)

const isStopping = ref(false)
const startStopKeys = computed(() => formatShortcut(shortcutsStore.getBinding('task.startStop')))

watch(() => statusStore.resourceStatus, (newStatus, oldStatus) => {
    if (oldStatus === 'loading' && newStatus === 'loaded') {
        refreshNodes()
    }
})

async function onStart() {
    if (!canStart.value) return
    const result = await runTask(selectedEntry.value, {})
    if (!result.succeed) {
        toast.add({ id: 'task-toast', title: 'Task Run Failed', description: result.msg, icon: 'i-lucide-circle-x', color: 'error' })
    } else {
        toast.add({ id: 'task-toast', title: 'Task Started', icon: 'i-lucide-play', color: 'success' })
    }
}

async function onStop() {
    if (isStopping.value) return
    isStopping.value = true
    try {
        const result = await stopTask()
        if (!result.succeed) {
            toast.add({ id: 'task-toast', title: 'Task Stop Failed', description: result.msg, icon: 'i-lucide-circle-x', color: 'error' })
        } else {
            toast.add({ id: 'task-toast', title: 'Task Stop Requested', icon: 'i-lucide-circle-stop', color: 'warning' })
        }
    } finally {
        isStopping.value = false
    }
}

function onEditOverride() {
    // TODO: Open editor for task override
}

async function refreshNodes() {
    const nodes = await getTaskNodes()
    entries.value = nodes.map((n) => ({ label: n, value: n }))
    if (!selectedEntry.value && entries.value.length > 0) {
        selectedEntry.value = entries.value[0]!.value
    }
}

function onKeydown(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
    if ((e.target as HTMLElement)?.isContentEditable) return
    if (shortcutsStore.matches(e, 'task.startStop')) {
        e.preventDefault()
        if (isRunning.value) onStop()
        else if (canStart.value) onStart()
    }
}

// ===================== Screenshot =====================

const MIN_ZOOM = 0.1
const MAX_ZOOM = 5
const ZOOM_STEP = 0.15

const imageData = ref<ArrayBuffer | null>(null)
const imageUrl = ref<string | null>(null)
let pendingFrame: ArrayBuffer | null = null
let rafId = 0
const aspectMode = ref<'landscape' | 'portrait'>('landscape')
const zoomLevel = ref(1)
const isFullscreen = ref(false)

const isStreaming = computed(() => screenshotRunning.value)
const isPaused = computed(() => screenshotPaused.value)
const currentFps = computed(() => screenshotFps.value)
const fpsSlider = ref(30)

const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const panOffset = ref({ x: 0, y: 0 })

const fullscreenZoom = ref(1)
const isFullscreenDragging = ref(false)
const fullscreenDragStart = ref({ x: 0, y: 0 })
const fullscreenDragOffset = ref({ x: 0, y: 0 })
const fullscreenPanOffset = ref({ x: 0, y: 0 })

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

async function toggleStreaming() {
    if (isStreaming.value) {
        await stopScreenshot()
        screenshotRunning.value = false
    } else {
        screenshotError.value = ''
        await startScreenshot()
        screenshotRunning.value = true
        screenshotPaused.value = false
    }
}

async function togglePause() {
    if (isPaused.value) {
        await resumeScreenshot()
        screenshotPaused.value = false
    } else {
        await pauseScreenshot()
        screenshotPaused.value = true
    }
}

async function applyFps() {
    const result = await setScreenshotFPS(fpsSlider.value)
    if (result.succeed && result.data) {
        screenshotFps.value = result.data.fps
        fpsSlider.value = result.data.fps
    }
}

function flushFrame() {
    rafId = 0
    if (pendingFrame) {
        imageData.value = pendingFrame
        pendingFrame = null
    }
}

watch(latestFrame, (frame) => {
    if (!frame) {
        imageData.value = null
        return
    }
    pendingFrame = frame
    if (!rafId) {
        rafId = requestAnimationFrame(flushFrame)
    }
})

function toggleAspect() {
    aspectMode.value = aspectMode.value === 'landscape' ? 'portrait' : 'landscape'
    resetZoom()
}

function zoomIn() { zoomLevel.value = Math.min(MAX_ZOOM, zoomLevel.value + ZOOM_STEP) }
function zoomOut() {
    zoomLevel.value = Math.max(MIN_ZOOM, zoomLevel.value - ZOOM_STEP)
    if (zoomLevel.value <= 1) panOffset.value = { x: 0, y: 0 }
}
function resetZoom() { zoomLevel.value = 1; panOffset.value = { x: 0, y: 0 } }
function onWheel(e: WheelEvent) {
    if (!imageUrl.value) return
    if (e.deltaY < 0) zoomIn(); else zoomOut()
}

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
function onDragEnd() { isDragging.value = false }

function fullscreenZoomIn() { fullscreenZoom.value = Math.min(MAX_ZOOM, fullscreenZoom.value + ZOOM_STEP) }
function fullscreenZoomOut() {
    fullscreenZoom.value = Math.max(MIN_ZOOM, fullscreenZoom.value - ZOOM_STEP)
    if (fullscreenZoom.value <= 1) fullscreenPanOffset.value = { x: 0, y: 0 }
}
function resetFullscreenZoom() { fullscreenZoom.value = 1; fullscreenPanOffset.value = { x: 0, y: 0 } }
function onFullscreenWheel(e: WheelEvent) {
    if (e.deltaY < 0) fullscreenZoomIn(); else fullscreenZoomOut()
}

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
function onFullscreenDragEnd() { isFullscreenDragging.value = false }

function downloadImage() {
    if (!imageData.value) return
    const blob = new Blob([new Uint8Array(imageData.value)], { type: 'image/jpeg' })
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
    a.download = `screenshot_${timestamp}.jpg`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
}

function updateImageUrl() {
    if (imageUrl.value) {
        URL.revokeObjectURL(imageUrl.value)
        imageUrl.value = null
    }
    if (!imageData.value) return
    const blob = new Blob([new Uint8Array(imageData.value)], { type: 'image/jpeg' })
    imageUrl.value = URL.createObjectURL(blob)
}

watch(imageData, () => { updateImageUrl() })
watch(isFullscreen, (val) => { if (!val) resetFullscreenZoom() })

// ===================== Lifecycle =====================

onMounted(async () => {
    window.addEventListener('keydown', onKeydown)
    refreshNodes()
    const status = await getScreenshotStatus()
    if (status) {
        screenshotRunning.value = status.running
        screenshotPaused.value = status.paused
        screenshotFps.value = status.fps
        fpsSlider.value = status.fps
    }
})

onUnmounted(() => {
    window.removeEventListener('keydown', onKeydown)
    if (rafId) cancelAnimationFrame(rafId)
    if (imageUrl.value) URL.revokeObjectURL(imageUrl.value)
})

// ===================== Public API =====================

defineExpose({
    selectedEntry,
    taskStatus,
    entries,
    refreshNodes,
    imageData,
    aspectMode,
})
</script>
