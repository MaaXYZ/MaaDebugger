<template>
    <UModal :open="open" title="Screenshot" fullscreen @update:open="emit('update:open', $event)">
        <template #body>
            <div class="relative w-full h-full flex items-center justify-center overflow-hidden bg-muted"
                 @wheel.prevent="emit('wheel', $event)">
                <div class="flex items-center justify-center cursor-grab select-none"
                     :class="{ 'cursor-grabbing': isDragging }" @mousedown="emit('drag-start', $event)"
                     @mousemove="emit('drag-move', $event)" @mouseup="emit('drag-end')" @mouseleave="emit('drag-end')">
                    <img v-if="imageUrl" :src="imageUrl" alt="Screenshot" draggable="false"
                         class="pointer-events-none max-w-none" :style="imageStyle" />
                </div>
                <div
                    class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-elevated/90 backdrop-blur-sm rounded-lg px-3 py-2 border border-default shadow-lg">
                    <UTooltip text="Zoom out">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="sm"
                                 :disabled="zoomLevel <= minZoom" @click="emit('zoom-out')" />
                    </UTooltip>
                    <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                        {{ zoomPercentage }}%
                    </span>
                    <UTooltip text="Zoom in">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="sm"
                                 :disabled="zoomLevel >= maxZoom" @click="emit('zoom-in')" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Fit to view">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm"
                                 @click="emit('reset-zoom')" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Download">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm"
                                 @click="emit('download')" />
                    </UTooltip>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import type { CSSProperties } from 'vue'

defineProps<{
    open: boolean
    imageUrl: string
    isDragging: boolean
    imageStyle: string | CSSProperties
    zoomLevel: number
    zoomPercentage: number
    minZoom: number
    maxZoom: number
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
    wheel: [event: WheelEvent]
    'drag-start': [event: MouseEvent]
    'drag-move': [event: MouseEvent]
    'drag-end': []
    'zoom-in': []
    'zoom-out': []
    'reset-zoom': []
    download: []
}>()
</script>
