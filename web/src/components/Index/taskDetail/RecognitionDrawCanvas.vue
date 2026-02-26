<template>
    <div :class="['flex flex-col gap-3', fullscreen ? 'h-full' : '']">
        <!-- Draw mode selector -->
        <div class="flex flex-row items-center gap-2 flex-wrap">
            <span class="text-xs text-dimmed font-medium">Draw:</span>
            <UButtonGroup size="xs">
                <UButton :variant="drawMode === 'best' ? 'solid' : 'outline'" color="primary"
                    @click="setDrawMode('best')">
                    Best
                </UButton>
                <UButton :variant="drawMode === 'filtered' ? 'solid' : 'outline'" color="primary"
                    @click="setDrawMode('filtered')">
                    Filtered
                </UButton>
                <UButton :variant="drawMode === 'all' ? 'solid' : 'outline'" color="primary"
                    @click="setDrawMode('all')">
                    All
                </UButton>
                <UButton :variant="drawMode === 'custom' ? 'solid' : 'outline'" color="primary"
                    @click="setDrawMode('custom')">
                    Custom
                </UButton>
            </UButtonGroup>
            <span class="text-xs text-dimmed tabular-nums">
                ({{ activeResults.length }} result{{ activeResults.length !== 1 ? 's' : '' }})
            </span>
        </div>

        <!-- Custom selection: checkboxes for individual results -->
        <div v-if="drawMode === 'custom' && allResults.length > 0"
            class="flex flex-col gap-1 max-h-40 overflow-y-auto rounded-lg border border-default p-2 bg-elevated">
            <label v-for="(item, idx) in allResults" :key="idx"
                class="flex items-center gap-2 text-xs cursor-pointer hover:bg-muted rounded px-1 py-0.5 transition-colors">
                <input type="checkbox" :checked="customSelection.has(idx)" class="accent-[var(--ui-primary)]"
                    @change="toggleCustom(idx)" />
                <span class="tabular-nums text-dimmed">#{{ idx }}</span>
                <span v-if="item.box">
                    [{{ item.box.x }}, {{ item.box.y }}, {{ item.box.w }}, {{ item.box.h }}]
                </span>
                <span v-if="getExtraLabel(item)" class="text-dimmed truncate max-w-48">
                    {{ getExtraLabel(item) }}
                </span>
            </label>
        </div>

        <!-- Canvas container -->
        <div ref="containerRef" class="relative overflow-hidden rounded-lg border border-default bg-muted"
            :style="containerStyle" @wheel.prevent="onWheel">
            <div v-if="rawImage" class="absolute inset-0 flex items-center justify-center select-none"
                :class="isDragging ? 'cursor-grabbing' : 'cursor-grab'" @mousedown="onDragStart" @mousemove="onDragMove"
                @mouseup="onDragEnd" @mouseleave="onDragEnd">
                <canvas ref="canvasRef" class="pointer-events-none" :style="canvasDisplayStyle" />
            </div>
            <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                <UIcon name="i-lucide-image-off" class="size-10" />
                <span class="text-xs">No raw image available</span>
                <span class="text-xs text-dimmed">Enable debug mode or save_draw in pipeline config</span>
            </div>
        </div>

        <!-- Toolbar -->
        <div v-if="rawImage" class="flex flex-row items-center gap-2">
            <div class="flex items-center gap-1">
                <UTooltip text="Zoom out">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="xs"
                        :disabled="zoomLevel <= MIN_ZOOM" @click="zoomOut" />
                </UTooltip>
                <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                    {{ zoomPercentage }}%
                </span>
                <UTooltip text="Zoom in">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="xs"
                        :disabled="zoomLevel >= MAX_ZOOM" @click="zoomIn" />
                </UTooltip>
            </div>
            <USeparator orientation="vertical" class="h-4" />
            <UTooltip text="Fit to view">
                <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="xs" @click="resetView" />
            </UTooltip>
            <USeparator orientation="vertical" class="h-4" />
            <UTooltip text="Download drawn image">
                <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="xs" @click="downloadCanvas" />
            </UTooltip>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import type { RecoDetailResponse, RecoResultItem } from './types'

// --- Constants ---
const MIN_ZOOM = 0.5
const MAX_ZOOM = 10
const ZOOM_STEP = 0.15

// Color palette for drawing boxes
const BOX_COLORS = [
    '#22c55e', // green (best)
    '#3b82f6', // blue (filtered)
    '#f59e0b', // amber (all others)
    '#ef4444', // red
    '#8b5cf6', // violet
    '#ec4899', // pink
    '#14b8a6', // teal
    '#f97316', // orange
]

const props = defineProps<{
    detail: RecoDetailResponse
    fullscreen?: boolean
}>()

type DrawMode = 'best' | 'filtered' | 'all' | 'custom'

const drawMode = ref<DrawMode>('best')
const customSelection = ref<Set<number>>(new Set())
const canvasRef = ref<HTMLCanvasElement | null>(null)
const containerRef = ref<HTMLElement | null>(null)
const zoomLevel = ref(1)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const panOffset = ref({ x: 0, y: 0 })
const rawImageObj = ref<HTMLImageElement | null>(null)

const rawImage = computed(() => props.detail.raw_image)
const results = computed(() => props.detail.results)

// Image natural dimensions
const imgWidth = computed(() => rawImageObj.value?.naturalWidth ?? 0)
const imgHeight = computed(() => rawImageObj.value?.naturalHeight ?? 0)
const imgAspect = computed(() => imgWidth.value && imgHeight.value ? imgWidth.value / imgHeight.value : 16 / 9)

// Container uses aspect-ratio to maintain proportions
const containerStyle = computed(() => {
    if (props.fullscreen) {
        return {
            width: '100%',
            height: '100%',
            flex: '1 1 0',
            minHeight: '0',
        }
    }
    return {
        aspectRatio: `${imgAspect.value}`,
        maxHeight: '60vh',
    }
})

// Canvas CSS display style: fit to container then apply zoom + pan
const canvasDisplayStyle = computed(() => {
    // Use max-width/max-height to fit the canvas within the container while preserving aspect ratio
    // The canvas element has its internal resolution set to the image's natural size,
    // and CSS constrains its display size to fit the container.
    return {
        maxWidth: '100%',
        maxHeight: '100%',
        transform: `translate(${panOffset.value.x}px, ${panOffset.value.y}px) scale(${zoomLevel.value})`,
        transformOrigin: 'center center',
        transition: isDragging.value ? 'none' : 'transform 0.15s ease-out',
    }
})

// Get all results from the detail
const allResults = computed<RecoResultItem[]>(() => {
    if (!results.value) return []
    return results.value.all ?? []
})

const bestResults = computed<RecoResultItem[]>(() => {
    if (!results.value) return []
    return results.value.best ?? []
})

const filteredResults = computed<RecoResultItem[]>(() => {
    if (!results.value) return []
    return results.value.filtered ?? []
})

// Active results based on draw mode
const activeResults = computed<RecoResultItem[]>(() => {
    switch (drawMode.value) {
        case 'best':
            return bestResults.value
        case 'filtered':
            return filteredResults.value
        case 'all':
            return allResults.value
        case 'custom':
            return allResults.value.filter((_, idx) => customSelection.value.has(idx))
    }
})

const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))

// --- Draw mode ---
function setDrawMode(mode: DrawMode) {
    drawMode.value = mode
    if (mode === 'custom' && customSelection.value.size === 0) {
        // Default: select all
        const newSet = new Set<number>()
        for (let i = 0; i < allResults.value.length; i++) {
            newSet.add(i)
        }
        customSelection.value = newSet
    }
}

function toggleCustom(idx: number) {
    const newSet = new Set(customSelection.value)
    if (newSet.has(idx)) {
        newSet.delete(idx)
    } else {
        newSet.add(idx)
    }
    customSelection.value = newSet
}

// --- Extra label ---
function getExtraLabel(item: RecoResultItem): string {
    if (!item.extra) return ''
    const parts: string[] = []
    if ('text' in item.extra) parts.push(`"${item.extra.text}"`)
    if ('score' in item.extra) parts.push(`score: ${Number(item.extra.score).toFixed(3)}`)
    if ('count' in item.extra) parts.push(`count: ${item.extra.count}`)
    if ('label' in item.extra) parts.push(`label: ${item.extra.label}`)
    return parts.join(', ')
}

// --- Canvas drawing ---
function drawCanvas() {
    const canvas = canvasRef.value
    const img = rawImageObj.value
    if (!canvas || !img) return

    canvas.width = img.naturalWidth
    canvas.height = img.naturalHeight

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Draw the raw image
    ctx.drawImage(img, 0, 0)

    // Draw result boxes
    const items = activeResults.value
    items.forEach((item, idx) => {
        if (!item.box) return

        const color = getBoxColor(idx)
        const { x, y, w, h } = item.box

        // Draw rectangle
        ctx.strokeStyle = color
        ctx.lineWidth = Math.max(2, Math.min(img.naturalWidth, img.naturalHeight) / 300)
        ctx.setLineDash([])
        ctx.strokeRect(x, y, w, h)

        // Semi-transparent fill
        ctx.fillStyle = color + '20'
        ctx.fillRect(x, y, w, h)

        // Label background
        const label = buildLabel(item, idx)
        const fontSize = Math.max(12, Math.min(img.naturalWidth, img.naturalHeight) / 60)
        ctx.font = `bold ${fontSize}px system-ui, sans-serif`
        const textMetrics = ctx.measureText(label)
        const textHeight = fontSize * 1.3
        const padding = fontSize * 0.3

        const labelX = x
        const labelY = y > textHeight + padding * 2 ? y - textHeight - padding * 2 : y + h

        ctx.fillStyle = color + 'DD'
        ctx.fillRect(labelX, labelY, textMetrics.width + padding * 2, textHeight + padding)

        // Label text
        ctx.fillStyle = '#ffffff'
        ctx.textBaseline = 'top'
        ctx.fillText(label, labelX + padding, labelY + padding * 0.5)
    })
}

function getBoxColor(idx: number): string {
    if (drawMode.value === 'best') return BOX_COLORS[0]!
    if (drawMode.value === 'filtered') return BOX_COLORS[1]!
    // For all/custom, cycle through colors
    return BOX_COLORS[idx % BOX_COLORS.length]!
}

function buildLabel(item: RecoResultItem, idx: number): string {
    const parts: string[] = [`#${idx}`]
    if (item.extra) {
        if ('score' in item.extra) parts.push(`${Number(item.extra.score).toFixed(3)}`)
        if ('text' in item.extra) parts.push(`"${item.extra.text}"`)
        if ('count' in item.extra) parts.push(`n=${item.extra.count}`)
        if ('label' in item.extra) parts.push(`${item.extra.label}`)
    }
    return parts.join(' ')
}

// --- Image loading ---
function loadRawImage() {
    if (!rawImage.value) {
        rawImageObj.value = null
        return
    }

    const img = new Image()
    img.onload = () => {
        rawImageObj.value = img
        nextTick(() => {
            drawCanvas()
        })
    }
    img.src = rawImage.value
}

// --- Zoom ---
function zoomIn() {
    zoomLevel.value = Math.min(MAX_ZOOM, +(zoomLevel.value + ZOOM_STEP).toFixed(2))
}

function zoomOut() {
    zoomLevel.value = Math.max(MIN_ZOOM, +(zoomLevel.value - ZOOM_STEP).toFixed(2))
    if (zoomLevel.value <= 1) {
        panOffset.value = { x: 0, y: 0 }
    }
}

function resetView() {
    zoomLevel.value = 1
    panOffset.value = { x: 0, y: 0 }
}

function onWheel(e: WheelEvent) {
    if (!rawImage.value) return
    if (e.deltaY < 0) {
        zoomIn()
    } else {
        zoomOut()
    }
}

// --- Drag ---
function onDragStart(e: MouseEvent) {
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

// --- Download ---
function downloadCanvas() {
    const canvas = canvasRef.value
    if (!canvas) return

    const url = canvas.toDataURL('image/png')
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
    a.download = `reco_draw_${drawMode.value}_${timestamp}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
}

// --- Watchers ---
watch(rawImage, () => {
    loadRawImage()
    resetView()
})

watch(activeResults, () => {
    nextTick(() => drawCanvas())
}, { deep: true })

watch(drawMode, () => {
    nextTick(() => drawCanvas())
})

onMounted(() => {
    loadRawImage()
})
</script>
