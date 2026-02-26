<template>
    <UModal v-model:open="open" title="Recognition Detail" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
        <template #body>
            <div v-if="loading" class="flex items-center justify-center p-8">
                <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
            </div>
            <div v-else-if="detail" class="flex flex-col gap-4">
                <!-- Header info -->
                <div class="flex flex-row items-center gap-2 flex-wrap">
                    <UBadge :color="detail.hit ? 'success' : 'error'" variant="subtle">
                        {{ detail.hit ? 'Hit' : 'Miss' }}
                    </UBadge>
                    <UBadge color="info" variant="subtle">{{ detail.algorithm }}</UBadge>
                    <span class="text-sm font-medium">{{ detail.name }}</span>
                </div>

                <!-- Box -->
                <div v-if="detail.box" class="text-xs text-dimmed">
                    Box: [{{ detail.box.x }}, {{ detail.box.y }}, {{ detail.box.w }}, {{ detail.box.h }}]
                </div>

                <!-- Combined Result (And/Or nesting) -->
                <div v-if="detail.combined_result && detail.combined_result.length > 0"
                     class="flex flex-col gap-2">
                    <span class="text-sm font-medium text-dimmed">Combined ({{ detail.algorithm }}):</span>
                    <div class="pl-3 border-l-2 border-default flex flex-col gap-2">
                        <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub"
                                        :depth="1" />
                    </div>
                </div>

                <!-- Canvas draw (raw image) or fallback draw images -->
                <div v-if="detail.raw_image && detail.results">
                    <div class="flex items-center gap-2 mb-2">
                        <span class="text-xs text-dimmed font-medium">Recognition Draw:</span>
                        <UTooltip text="Fullscreen">
                            <UButton color="neutral" variant="ghost" icon="i-lucide-fullscreen" size="xs"
                                     @click="isFullscreen = true" />
                        </UTooltip>
                    </div>
                    <RecognitionDrawCanvas :detail="detail" />
                </div>
                <div v-else-if="detail.draw_images && detail.draw_images.length > 0"
                     class="flex flex-col gap-2">
                    <span class="text-xs text-dimmed font-medium">Draw:</span>
                    <div class="flex flex-col gap-2">
                        <div v-for="(img, idx) in detail.draw_images" :key="idx" class="relative group">
                            <img :src="img"
                                 class="max-w-full rounded-lg border border-default cursor-pointer hover:opacity-80 transition-opacity"
                                 @click="openImagePreview(img)" />
                            <div
                                class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                <UButton color="neutral" variant="solid" icon="i-lucide-fullscreen" size="xs"
                                         @click="openImagePreview(img)" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                No detail available
            </div>
        </template>
    </UModal>

    <!-- Fullscreen Canvas Draw Modal -->
    <UModal v-model:open="isFullscreen" title="Recognition Draw" fullscreen>
        <template #body>
            <div v-if="detail?.raw_image && detail?.results"
                 class="w-full h-full flex flex-col overflow-hidden bg-muted p-4">
                <RecognitionDrawCanvas :detail="detail" fullscreen />
            </div>
        </template>
    </UModal>

    <!-- Fullscreen Image Preview (for original draw_images) -->
    <UModal v-model:open="imagePreviewOpen" title="Image Preview" fullscreen>
        <template #body>
            <div class="relative w-full h-full flex items-center justify-center overflow-hidden bg-muted"
                 @wheel.prevent="onPreviewWheel">
                <div class="flex items-center justify-center cursor-grab select-none"
                     :class="{ 'cursor-grabbing': isPreviewDragging }" @mousedown="onPreviewDragStart"
                     @mousemove="onPreviewDragMove" @mouseup="onPreviewDragEnd" @mouseleave="onPreviewDragEnd">
                    <img v-if="previewImageSrc" :src="previewImageSrc" alt="Preview" draggable="false"
                         class="pointer-events-none max-w-none" :style="previewImageStyle" />
                </div>
                <!-- Fullscreen toolbar -->
                <div
                    class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-elevated/90 backdrop-blur-sm rounded-lg px-3 py-2 border border-default shadow-lg">
                    <UTooltip text="Zoom out">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="sm"
                                 :disabled="previewZoom <= MIN_ZOOM" @click="previewZoomOut" />
                    </UTooltip>
                    <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                        {{ previewZoomPercentage }}%
                    </span>
                    <UTooltip text="Zoom in">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="sm"
                                 :disabled="previewZoom >= MAX_ZOOM" @click="previewZoomIn" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Fit to view">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm"
                                 @click="resetPreviewZoom" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip text="Download">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm"
                                 @click="downloadPreviewImage" />
                    </UTooltip>
                </div>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getRecoDetailById } from '@/api/http'
import type { RecoDetailResponse } from './types'
import RecoDetailItem from './RecoDetailItem.vue'
import RecognitionDrawCanvas from './RecognitionDrawCanvas.vue'

// --- Constants ---
const MIN_ZOOM = 0.1
const MAX_ZOOM = 5
const ZOOM_STEP = 0.15

const props = defineProps<{
    recoId: number | null
}>()

const open = defineModel<boolean>('open', { default: false })
const loading = ref(false)
const detail = ref<RecoDetailResponse | null>(null)

// Fullscreen canvas draw
const isFullscreen = ref(false)

// Image preview state (for original draw_images)
const imagePreviewOpen = ref(false)
const previewImageSrc = ref('')

// Preview zoom & drag
const previewZoom = ref(1)
const isPreviewDragging = ref(false)
const previewDragStart = ref({ x: 0, y: 0 })
const previewDragOffset = ref({ x: 0, y: 0 })
const previewPanOffset = ref({ x: 0, y: 0 })

const previewZoomPercentage = computed(() => Math.round(previewZoom.value * 100))

const previewImageStyle = computed(() => ({
    transform: `translate(${previewPanOffset.value.x}px, ${previewPanOffset.value.y}px) scale(${previewZoom.value})`,
    transformOrigin: 'center center',
    transition: isPreviewDragging.value ? 'none' : 'transform 0.2s ease',
    maxWidth: '90vw',
    maxHeight: '85vh',
    objectFit: 'contain' as const,
}))

function openImagePreview(src: string) {
    previewImageSrc.value = src
    imagePreviewOpen.value = true
}

// --- Preview zoom ---
function previewZoomIn() {
    previewZoom.value = Math.min(MAX_ZOOM, previewZoom.value + ZOOM_STEP)
}

function previewZoomOut() {
    previewZoom.value = Math.max(MIN_ZOOM, previewZoom.value - ZOOM_STEP)
    if (previewZoom.value <= 1) {
        previewPanOffset.value = { x: 0, y: 0 }
    }
}

function resetPreviewZoom() {
    previewZoom.value = 1
    previewPanOffset.value = { x: 0, y: 0 }
}

function onPreviewWheel(e: WheelEvent) {
    if (e.deltaY < 0) {
        previewZoomIn()
    } else {
        previewZoomOut()
    }
}

// --- Preview drag ---
function onPreviewDragStart(e: MouseEvent) {
    if (previewZoom.value <= 1) return
    isPreviewDragging.value = true
    previewDragStart.value = { x: e.clientX, y: e.clientY }
    previewDragOffset.value = { ...previewPanOffset.value }
}

function onPreviewDragMove(e: MouseEvent) {
    if (!isPreviewDragging.value) return
    previewPanOffset.value = {
        x: previewDragOffset.value.x + (e.clientX - previewDragStart.value.x),
        y: previewDragOffset.value.y + (e.clientY - previewDragStart.value.y),
    }
}

function onPreviewDragEnd() {
    isPreviewDragging.value = false
}

// --- Download ---
function downloadPreviewImage() {
    if (!previewImageSrc.value) return
    const a = document.createElement('a')
    a.href = previewImageSrc.value
    const now = new Date()
    const timestamp = now.getFullYear().toString()
        + String(now.getMonth() + 1).padStart(2, '0')
        + String(now.getDate()).padStart(2, '0')
        + '_'
        + String(now.getHours()).padStart(2, '0')
        + String(now.getMinutes()).padStart(2, '0')
        + String(now.getSeconds()).padStart(2, '0')
    a.download = `reco_draw_${timestamp}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
}

// Reset preview zoom when modal closes
watch(imagePreviewOpen, (val) => {
    if (!val) {
        resetPreviewZoom()
    }
})

watch([() => props.recoId, open], async ([id, isOpen]) => {
    if (!isOpen || id == null) {
        detail.value = null
        return
    }
    loading.value = true
    try {
        detail.value = await getRecoDetailById(id)
    } catch {
        detail.value = null
    } finally {
        loading.value = false
    }
})
</script>
