<template>
    <div class="relative overflow-hidden rounded-md border border-default bg-muted" :style="containerStyle"
        @wheel.prevent="emit('wheel', $event)">
        <div v-if="imageUrl" class="absolute inset-0 z-0 flex items-center justify-center cursor-grab select-none"
            :class="{ 'cursor-grabbing': isDragging }" @mousedown="emit('drag-start', $event)"
            @mousemove="emit('drag-move', $event)" @mouseup="emit('drag-end')" @mouseleave="emit('drag-end')">
            <img :src="imageUrl" :alt="t('task.screenshotImgAlt')" draggable="false"
                class="pointer-events-none w-full h-full object-contain" :style="imageStyle" />
        </div>

        <div v-if="overlayVisible"
            class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 px-6 text-center backdrop-blur-sm"
            :class="overlayClass">
            <UIcon :name="overlayIcon" class="size-12" />
            <span class="text-sm font-semibold">{{ overlayTitle }}</span>
            <span v-if="overlayDescription" class="max-w-xs text-xs text-dimmed">{{ overlayDescription }}</span>
        </div>

        <div v-else-if="screenshotError"
            class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-error">
            <UIcon name="i-lucide-circle-x" class="size-12" />
            <span class="text-sm font-medium">{{ t('task.screenshotFailed') }}</span>
            <span class="text-xs text-dimmed max-w-xs text-center">{{ screenshotError }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CSSProperties } from 'vue'
import type { ScreenshotOverlayState } from '@/stores/screenshot'

const props = defineProps<{
    imageUrl: string
    screenshotError: string
    overlayState: ScreenshotOverlayState
    overlayMessage: string
    isDragging: boolean
    containerStyle: string | CSSProperties
    imageStyle: string | CSSProperties
}>()

const emit = defineEmits<{
    wheel: [event: WheelEvent]
    'drag-start': [event: MouseEvent]
    'drag-move': [event: MouseEvent]
    'drag-end': []
}>()

const { t } = useI18n()

const overlayVisible = computed(() => props.overlayState !== 'none')
const overlayTitle = computed(() => {
    switch (props.overlayState) {
        case 'disconnected':
            return t('task.controllerDisconnected')
        case 'paused':
            return t('task.screenshotPaused')
        case 'failed':
            return t('task.screenshotFailed')
        default:
            return ''
    }
})
const overlayDescription = computed(() => props.overlayMessage || '')
const overlayIcon = computed(() => {
    switch (props.overlayState) {
        case 'disconnected':
            return 'i-lucide-unlink'
        case 'paused':
            return 'i-lucide-pause-circle'
        case 'failed':
            return 'i-lucide-circle-x'
        default:
            return 'i-lucide-image'
    }
})
const overlayClass = computed(() => {
    switch (props.overlayState) {
        case 'disconnected':
            return 'text-warning bg-default/70'
        case 'paused':
            return 'text-info bg-default/70'
        case 'failed':
            return 'text-error bg-default/70'
        default:
            return 'text-muted bg-default/70'
    }
})
</script>
