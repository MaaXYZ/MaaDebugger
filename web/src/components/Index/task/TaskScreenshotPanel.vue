<template>
    <div class="relative overflow-hidden rounded-md border border-default bg-muted" :style="containerStyle"
         @wheel.prevent="emit('wheel', $event)">
        <div v-if="imageUrl" class="absolute inset-0 flex items-center justify-center cursor-grab select-none"
             :class="{ 'cursor-grabbing': isDragging }" @mousedown="emit('drag-start', $event)"
             @mousemove="emit('drag-move', $event)" @mouseup="emit('drag-end')" @mouseleave="emit('drag-end')">
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
</template>

<script setup lang="ts">
import type { CSSProperties } from 'vue'

defineProps<{
    imageUrl: string
    screenshotError: string
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
</script>
