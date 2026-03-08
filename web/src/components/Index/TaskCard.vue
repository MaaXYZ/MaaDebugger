<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task</span>
                <TaskStatusBadge :status="taskStatus" />
                <div class="flex-1"></div>

                <UPopover v-if="hasInterfaceTasks">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-list-tree" size="sm">
                        Interface Task
                    </UButton>
                    <template #content>
                        <div class="w-[28rem] max-w-[90vw] p-3 flex flex-col gap-3">
                            <div class="text-sm font-medium">Task Preset</div>
                            <USelectMenu :model-value="taskStore.selectedInterfaceTaskName" :items="interfaceTaskItems"
                                value-key="value" @update:model-value="onInterfaceTaskSelected" />

                            <div v-if="selectedInterfaceTask"
                                class="rounded-lg border border-default bg-elevated/50 p-3 space-y-3">
                                <div>
                                    <div class="text-sm font-medium text-default">
                                        {{ selectedInterfaceTask.name }}
                                    </div>
                                    <div class="text-xs text-dimmed break-all">
                                        Entry: {{ selectedInterfaceTask.entry || 'n/a' }}
                                    </div>
                                    <div v-if="selectedInterfaceTask.description"
                                        class="mt-1 text-xs text-dimmed whitespace-pre-wrap">
                                        {{ selectedInterfaceTask.description }}
                                    </div>
                                </div>

                                <div v-if="selectedTaskOptionDefs.length" class="space-y-3">
                                    <div class="text-xs font-medium text-default">Options</div>
                                    <div v-for="optionDef in selectedTaskOptionDefs" :key="optionDef.name"
                                        class="space-y-2">
                                        <div class="flex items-start justify-between gap-3">
                                            <div class="min-w-0">
                                                <div class="text-xs font-medium text-default break-all">
                                                    {{ optionDef.name }}
                                                </div>
                                                <div v-if="optionDef.description"
                                                    class="text-xs text-dimmed whitespace-pre-wrap break-all">
                                                    {{ optionDef.description }}
                                                </div>
                                            </div>
                                            <UBadge color="neutral" variant="subtle" size="sm">
                                                {{ optionDef.type || 'option' }}
                                            </UBadge>
                                        </div>
                                        <USelectMenu :model-value="selectedCaseMap[optionDef.name]"
                                            :items="buildOptionCaseItems(optionDef)" value-key="value"
                                            @update:model-value="(value) => onOptionCaseSelected(optionDef.name, value as string)" />
                                    </div>
                                </div>

                                <div class="space-y-2">
                                    <div class="text-xs font-medium text-default">Generated Override</div>
                                    <pre
                                        class="max-h-48 overflow-auto rounded-md border border-default bg-default p-2 text-[11px] leading-5 text-toned">{{ derivedOverridePreview }}</pre>
                                </div>
                            </div>
                        </div>
                    </template>
                </UPopover>

                <UPopover>
                    <UButton color="neutral" variant="ghost" size="sm" class="tabular-nums">
                        {{ currentFps }} /
                        <span class="font-medium" :class="actualFpsTone">{{ actualFpsLabel }}</span>
                        FPS
                    </UButton>
                    <template #content>
                        <div class="p-3 flex flex-col gap-3 w-56">
                            <div class="flex items-center justify-between gap-3">
                                <span class="text-xs text-muted">Target FPS</span>
                                <span class="text-xs tabular-nums">{{ currentFps }}</span>
                            </div>
                            <div class="flex items-center justify-between gap-3">
                                <span class="text-xs text-muted">Actual FPS</span>
                                <span class="text-xs tabular-nums font-medium" :class="actualFpsTone">{{ actualFpsLabel
                                }}</span>
                            </div>
                            <USeparator />
                            <div class="flex flex-col gap-2">
                                <span class="text-xs text-muted">Frame Rate</span>
                                <USlider v-model="fpsSlider" :min="1" :max="60" :step="1" />
                                <div class="flex items-center justify-between">
                                    <span class="text-xs text-muted tabular-nums">{{ fpsSlider }} FPS</span>
                                    <UButton size="xs" @click="applyFps">Apply</UButton>
                                </div>
                            </div>
                        </div>
                    </template>
                </UPopover>

                <UTooltip :text="isPaused ? 'Resume' : 'Pause'">
                    <UButton color="neutral" variant="ghost" :icon="isPaused ? 'i-lucide-play' : 'i-lucide-pause'"
                        size="sm" :disabled="!isStreaming" @click="togglePause" />
                </UTooltip>

                <UTooltip :text="isStreaming ? 'Stop streaming' : 'Start streaming'">
                    <UButton :color="isStreaming ? 'error' : 'primary'" variant="ghost"
                        :icon="isStreaming ? 'i-lucide-video-off' : 'i-lucide-video'" size="sm"
                        @click="toggleStreaming" />
                </UTooltip>

                <UTooltip :text="aspectMode === 'landscape' ? 'Switch to 9:16 portrait' : 'Switch to 16:9 landscape'">
                    <UButton color="neutral" variant="ghost" :icon="aspectMode === 'landscape'
                        ? 'i-lucide-monitor'
                        : 'i-lucide-smartphone'" size="sm" @click="toggleAspect" />
                </UTooltip>
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
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
                            :loading="isPreparingOverrideEditor" :disabled="isPreparingOverrideEditor"
                            @click="onEditOverride">
                            <template v-if="hasInterfaceTasks" #trailing>
                                <UBadge color="neutral" variant="subtle" size="sm">
                                    {{ selectedTaskOptionSelections.length }}
                                </UBadge>
                            </template>
                        </UButton>
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

                <div v-if="hasInterfaceTasks && selectedInterfaceTask"
                    class="rounded-lg border border-default bg-elevated/40 p-3 text-sm">
                    <div class="flex items-start justify-between gap-3">
                        <div class="min-w-0">
                            <div class="font-medium text-default break-all">
                                {{ selectedInterfaceTask.name }}
                            </div>
                            <div class="text-xs text-dimmed break-all">
                                Entry: {{ selectedInterfaceTask.entry || selectedEntry }}
                            </div>
                        </div>
                        <UBadge color="primary" variant="subtle">
                            {{ selectedTaskOptionSelections.length }} option(s)
                        </UBadge>
                    </div>
                </div>

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

    <component :is="jsonEditorModalComponent" v-if="jsonEditorModalComponent" v-model:open="overrideEditorOpen"
        :model-value="taskStore.manualOverrideJson" title="Pipeline Override"
        description="Edit manual override patch (merged on top of selected interface task options)"
        :schema="editorSchema" :external-schemas="editorExternalSchemas" @update:model-value="setManualOverrideJson" />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, shallowRef, watch, watchEffect } from 'vue'
import type { Component } from 'vue'
import TaskStatusBadge from './task/TaskStatusBadge.vue'
import useTaskControls from './task/useTaskControls'
import { useScreenshotStream } from './task/useScreenshotStream'
import { MAX_ZOOM, MIN_ZOOM, usePanZoom } from './task/usePanZoom'
import { warmupMonacoJsonWorker } from '@/components/MonacoEditor'
import type { InterfaceTaskOptionDefinition } from '@/types/interface'

const toast = useToast()
const jsonEditorModalComponent = shallowRef<Component | null>(null)
const editorSchema = shallowRef<Record<string, unknown> | undefined>()
const editorExternalSchemas = shallowRef<Record<string, Record<string, unknown>> | undefined>()
const isPreparingOverrideEditor = ref(false)
let editorAssetsLoaded = false

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
    hasInterfaceTasks,
    interfaceTaskItems,
    selectedInterfaceTask,
    selectedTaskOptionDefs,
    selectedTaskOptionSelections,
    selectInterfaceTask,
    setInterfaceOptionCase,
    setManualOverrideJson,
    onStart,
    onStop,
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
    actualFps,
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

const actualFpsLabel = computed(() => actualFps.value.toFixed(1))
const actualFpsTone = computed(() => {
    if (actualFps.value >= currentFps.value * 0.9) return 'text-success'
    if (actualFps.value >= currentFps.value * 0.6) return 'text-warning'
    return 'text-error'
})
const selectedCaseMap = computed<Record<string, string>>(() =>
    Object.fromEntries(selectedTaskOptionSelections.value.map((item) => [item.optionName, item.caseName])),
)
const derivedOverridePreview = computed(() =>
    JSON.stringify(taskStore.derivedInterfaceOverride, null, 2),
)

watchEffect(() => {
    document.documentElement.style.setProperty('--entry-content-min-w', entryContentMinWidth.value)
})

async function ensureEditorAssetsLoaded() {
    if (editorAssetsLoaded) return

    const [modalModule, pipelineSchemaModule, customActionSchemaModule, customRecognitionSchemaModule] = await Promise.all([
        import('@/components/MonacoEditor/JsonEditorModal.vue'),
        import('@/schema/pipeline.schema.json'),
        import('@/schema/custom.action.schema.json'),
        import('@/schema/custom.recognition.schema.json'),
        warmupMonacoJsonWorker(),
    ])

    jsonEditorModalComponent.value = defineAsyncComponent(() => Promise.resolve(modalModule.default))
    editorSchema.value = pipelineSchemaModule.default as Record<string, unknown>
    editorExternalSchemas.value = {
        './custom.action.schema.json': customActionSchemaModule.default as Record<string, unknown>,
        './custom.recognition.schema.json': customRecognitionSchemaModule.default as Record<string, unknown>,
    }
    editorAssetsLoaded = true
}

function buildOptionCaseItems(optionDef: InterfaceTaskOptionDefinition) {
    return (optionDef.cases ?? []).map((item) => ({
        label: item.name,
        value: item.name,
        description: item.description || (item.pipeline_override_keys?.length
            ? `Override: ${item.pipeline_override_keys.join(', ')}`
            : 'No override'),
    }))
}

function onInterfaceTaskSelected(value: string | number) {
    if (typeof value !== 'string') return
    selectInterfaceTask(value)
}

function onOptionCaseSelected(optionName: string, value: string) {
    if (!value) return
    setInterfaceOptionCase(optionName, value)
}

function onImageWheel(e: WheelEvent) {
    onWheel(e, !!imageUrl.value)
}

watch(isFullscreen, (val) => {
    handleFullscreenChange(val)
})

async function onEditOverride() {
    if (isPreparingOverrideEditor.value) {
        return
    }

    isPreparingOverrideEditor.value = true
    try {
        await ensureEditorAssetsLoaded()
        overrideEditorOpen.value = true
    } finally {
        isPreparingOverrideEditor.value = false
    }
}

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
