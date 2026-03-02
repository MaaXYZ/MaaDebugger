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
                <UTooltip :text="aspectMode === 'landscape' ? 'Switch to 9:16 portrait' : 'Switch to 16:9 landscape'">
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
                        <UTooltip :text="selectedEntry">
                            <USelectMenu v-model="selectedEntry" v-model:search-term="entrySearchTerm" virtualize
                                         :items="entrySelectItems" ignore-filter placeholder="Select task entry..."
                                         :search-input="{
                                             placeholder: 'Filter...',
                                             icon: 'i-lucide-search'
                                         }"
                                         :ui="{ base: 'w-full', content: '!w-auto min-w-(--entry-content-min-w) max-w-[80vw]' }"
                                         class="w-full" size="xl" value-key="value" :disabled="isRunning" arrow />
                        </UTooltip>
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
                    <UButton v-else color="error" variant="soft" icon="i-lucide-square" size="xl" :loading="isStopping"
                             :disabled="isStopping" @click="onStop">
                        <template v-if="startStopKeys.length" #trailing>
                            <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                        </template>
                    </UButton>
                </div>

                <!-- Screenshot area -->
                <div class="relative overflow-hidden rounded-md border border-default bg-muted" :style="containerStyle"
                     @wheel.prevent="onImageWheel">
                    <div v-if="imageUrl"
                         class="absolute inset-0 flex items-center justify-center cursor-grab select-none"
                         :class="{ 'cursor-grabbing': isDragging }" @mousedown="onDragStart" @mousemove="onDragMove"
                         @mouseup="onDragEnd" @mouseleave="onDragEnd">
                        <img :src="imageUrl" alt="Screenshot" draggable="false"
                             class="pointer-events-none w-full h-full object-contain" :style="imageStyle" />
                    </div>
                    <div v-else-if="screenshotError"
                         class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-error">
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

    <!-- Override editor modal -->
    <JsonEditorModal v-model:open="overrideEditorOpen" v-model="taskStore.overrideJson" title="Pipeline Override"
                     description="Edit pipeline override (JSONC with schema validation)" :schema="pipelineSchema" :external-schemas="{
                         './custom.action.schema.json': customActionSchema,
                         './custom.recognition.schema.json': customRecognitionSchema,
                     }" />
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, watch, watchEffect } from 'vue'
import TaskStatusBadge from './task/TaskStatusBadge.vue'
import { useTaskControls } from './task/useTaskControls'
import { useScreenshotStream } from './task/useScreenshotStream'
import { MAX_ZOOM, MIN_ZOOM, usePanZoom } from './task/usePanZoom'
import { JsonEditorModal } from '@/components/MonacoEditor'
import pipelineSchema from '@/schema/pipeline.schema.json'
import customActionSchema from '@/schema/custom.action.schema.json'
import customRecognitionSchema from '@/schema/custom.recognition.schema.json'

const toast = useToast()

const {
    entries,
    selectedEntry,
    taskStatus,
    entrySearchTerm,
    entrySelectItems,
    entryContentMinWidth,
    isRunning,
    canStart,
    isStopping,
    startStopKeys,
    overrideEditorOpen,
    taskStore,
    onStart,
    onStop,
    onEditOverride,
    refreshNodes,
    mount,
    unmount,
} = useTaskControls(toast)

const {
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
} = useScreenshotStream()

const {
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
} = usePanZoom()

// Sync entry content min-width CSS variable to :root for portal-rendered dropdown
watchEffect(() => {
    document.documentElement.style.setProperty('--entry-content-min-w', entryContentMinWidth.value)
})

function onImageWheel(e: WheelEvent) {
    onWheel(e, !!imageUrl.value)
}

watch(isFullscreen, (val) => {
    handleFullscreenChange(val)
})

onMounted(async () => {
    mount()
    await initStatus()
})

onUnmounted(() => {
    unmount()
})

defineExpose({
    selectedEntry,
    taskStatus,
    entries,
    refreshNodes,
    imageData,
    aspectMode,
})
</script>
