<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task</span>
                <TaskStatusBadge :status="taskStatus" />
                <div class="flex-1"></div>

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
                                <USlider v-model="fpsSlider" :min="1" :max="30" :step="1" />
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
                <TaskLaunchPanel v-model:selected-entry="selectedEntry" v-model:entry-search-term="entrySearchTerm"
                                 v-model:task-launch-mode="taskLaunchMode" :entry-select-items="entrySelectItems"
                                 :is-running="isRunning" :can-start="canStart" :is-stopping="isStopping"
                                 :start-stop-keys="startStopKeys" :is-preparing-override-editor="isPreparingOverrideEditor"
                                 :has-interface-tasks="hasInterfaceTasks" :interface-task-items="interfaceTaskItems"
                                 :selected-interface-task="selectedInterfaceTask"
                                 :selected-task-option-selections="selectedTaskOptionSelections" :effective-entry="effectiveEntry"
                                 @interface-task-selected="onInterfaceTaskSelected"
                                 @open-interface-task-modal="openInterfaceTaskModal" @edit-override="onEditOverride" @start="onStart"
                                 @stop="onStop" />

                <TaskScreenshotPanel :image-url="imageUrl || ''" :screenshot-error="screenshotError || ''"
                                     :is-dragging="isDragging" :container-style="containerStyle" :image-style="imageStyle"
                                     @wheel="onImageWheel" @drag-start="onDragStart" @drag-move="onDragMove" @drag-end="onDragEnd" />
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

    <TaskFullscreenModal v-model:open="isFullscreen" :image-url="imageUrl || ''" :is-dragging="isFullscreenDragging"
                         :image-style="fullscreenImageStyle" :zoom-level="fullscreenZoom" :zoom-percentage="fullscreenZoomPercentage"
                         :min-zoom="MIN_ZOOM" :max-zoom="MAX_ZOOM" @wheel="onFullscreenWheel" @drag-start="onFullscreenDragStart"
                         @drag-move="onFullscreenDragMove" @drag-end="onFullscreenDragEnd" @zoom-in="fullscreenZoomIn"
                         @zoom-out="fullscreenZoomOut" @reset-zoom="resetFullscreenZoom" @download="downloadImage" />

    <TaskInterfaceModal v-model:open="interfaceTaskModalOpen" :selected-task="draftSelectedTask"
                        :option-defs="draftTaskOptionDefs" :selected-case-map="draftSelectedCaseMap"
                        @select-case="onOptionCaseDraftSelected" @cancel="onInterfaceTaskCancel" @confirm="onInterfaceTaskConfirm" />

    <component :is="jsonEditorModalComponent" v-if="jsonEditorModalComponent" v-model:open="overrideEditorOpen"
               v-model="overrideEditorDraft" title="Pipeline Override"
               description="Edit the effective override JSON. Interface-generated values remain synced until you change them manually."
               :schema="editorSchema" :external-schemas="editorExternalSchemas" />
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, nextTick, onMounted, onUnmounted, ref, shallowRef, watch, watchEffect } from 'vue'
import type { Component } from 'vue'
import TaskStatusBadge from './task/TaskStatusBadge.vue'
import TaskLaunchPanel from './task/TaskLaunchPanel.vue'
import TaskScreenshotPanel from './task/TaskScreenshotPanel.vue'
import TaskFullscreenModal from './task/TaskFullscreenModal.vue'
import TaskInterfaceModal from './task/TaskInterfaceModal.vue'
import useTaskControls from './task/useTaskControls'
import { useScreenshotStream } from './task/useScreenshotStream'
import { MAX_ZOOM, MIN_ZOOM, usePanZoom } from './task/usePanZoom'
import { warmupMonacoJsonWorker } from '@/components/MonacoEditor'

const toast = useToast()
const jsonEditorModalComponent = shallowRef<Component | null>(null)
const editorSchema = shallowRef<Record<string, unknown> | undefined>()
const editorExternalSchemas = shallowRef<Record<string, Record<string, unknown>> | undefined>()
const isPreparingOverrideEditor = ref(false)
const interfaceTaskModalOpen = ref(false)
const overrideEditorDraft = ref('{}')
const interfaceTaskDraftName = ref('')
const interfaceOptionDraftSelections = ref<Array<{ optionName: string, caseName: string }>>([])
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
    selectedTaskOptionSelections,
    taskLaunchMode,
    usingInterfaceTask,
    effectiveEntry,
    selectInterfaceTask,
    setInterfaceOptionCase,
    setOverrideJson,
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
const draftSelectedTask = computed(() =>
    taskStore.interfaceTasks.find((task) => task.name === interfaceTaskDraftName.value) ?? null,
)
const draftTaskOptionDefs = computed(() => draftSelectedTask.value?.option_defs ?? [])
const draftSelectedCaseMap = computed<Record<string, string>>(() =>
    Object.fromEntries(interfaceOptionDraftSelections.value.map((item) => [item.optionName, item.caseName])),
)

watch(() => taskStore.overrideJson, (value) => {
    if (value !== overrideEditorDraft.value) {
        overrideEditorDraft.value = value
    }
}, { immediate: true })

watch(overrideEditorDraft, (value) => {
    if (value !== taskStore.overrideJson) {
        setOverrideJson(value)
    }
})

watch(() => hasInterfaceTasks.value, (visible) => {
    if (!visible && taskLaunchMode.value !== 'manual') {
        taskLaunchMode.value = 'manual'
    }
}, { immediate: true })

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

function snapshotInterfaceDraft() {
    interfaceTaskDraftName.value = taskStore.selectedInterfaceTaskName
    interfaceOptionDraftSelections.value = selectedTaskOptionSelections.value.map((item) => ({ ...item }))
}

function openInterfaceTaskModal() {
    if (!taskStore.selectedInterfaceTaskName && interfaceTaskItems.value.length > 0) {
        selectInterfaceTask(interfaceTaskItems.value[0]?.value ?? '')
    }
    snapshotInterfaceDraft()
    interfaceTaskModalOpen.value = true
}

function onInterfaceTaskSelected(value: string) {
    if (!value) return
    selectInterfaceTask(value)
}

function onOptionCaseDraftSelected(optionName: string, value: string) {
    if (!value) return
    interfaceOptionDraftSelections.value = [
        ...interfaceOptionDraftSelections.value.filter((item) => item.optionName !== optionName),
        { optionName, caseName: value },
    ]
}

async function onInterfaceTaskConfirm() {
    if (interfaceTaskDraftName.value) {
        taskLaunchMode.value = 'interface'
        selectInterfaceTask(interfaceTaskDraftName.value)
        await nextTick()
        for (const selection of interfaceOptionDraftSelections.value) {
            setInterfaceOptionCase(selection.optionName, selection.caseName)
        }
    }
    interfaceTaskModalOpen.value = false
}

function onInterfaceTaskCancel() {
    snapshotInterfaceDraft()
    interfaceTaskModalOpen.value = false
}

function onImageWheel(e: WheelEvent) {
    onWheel(e, !!imageUrl.value)
}

watch(isFullscreen, (val) => {
    handleFullscreenChange(val)
})

watch(usingInterfaceTask, (enabled) => {
    if (enabled && !taskStore.selectedInterfaceTaskName && interfaceTaskItems.value.length > 0) {
        interfaceTaskDraftName.value = interfaceTaskItems.value[0]?.value ?? ''
    }
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
    effectiveEntry,
})
</script>
