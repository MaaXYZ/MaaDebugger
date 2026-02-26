<template>
    <div ref="rootRef" :class="['flex flex-col gap-3', fullscreen ? 'h-full' : '']">
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

        <!-- Custom selection: split panel layout -->
        <div v-if="drawMode === 'custom' && allResults.length > 0"
            class="custom-panel-root rounded-lg border border-default bg-elevated overflow-hidden"
            :class="isNarrow ? 'custom-panel-vertical' : 'custom-panel-horizontal'">
            <!-- Left / Top: list panel -->
            <div class="custom-panel-list" :style="listPanelStyle">
                <!-- Search filter -->
                <div class="flex items-center gap-1.5 px-2 pt-2 pb-1">
                    <UInput v-model="customSearch" icon="i-lucide-search" size="xs"
                        placeholder="Filter..." class="flex-1" />
                    <UTooltip text="Select all">
                        <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-check-check"
                            @click="selectAllCustom" />
                    </UTooltip>
                    <UTooltip text="Deselect all">
                        <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-x"
                            @click="deselectAllCustom" />
                    </UTooltip>
                </div>
                <!-- List items -->
                <div class="flex-1 overflow-y-auto px-1 pb-1">
                    <label v-for="entry in filteredCustomEntries" :key="entry.idx"
                        class="flex items-center gap-2 text-xs cursor-pointer rounded px-1.5 py-1 transition-colors"
                        :class="[
                            customSelectedDetail === entry.idx ? 'bg-primary/10 text-primary' : 'hover:bg-muted',
                            hoveredIndex === entry.idx ? 'ring-1 ring-primary/40' : ''
                        ]"
                        @mouseenter="hoveredIndex = entry.idx"
                        @mouseleave="hoveredIndex = -1"
                        @click.stop="customSelectedDetail = entry.idx">
                        <input type="checkbox" :checked="customSelection.has(entry.idx)"
                            class="accent-primary shrink-0"
                            @change.stop="toggleCustom(entry.idx)"
                            @click.stop />
                        <span class="tabular-nums text-dimmed shrink-0">#{{ entry.idx }}</span>
                        <span v-if="entry.item.extra && 'score' in entry.item.extra"
                            class="tabular-nums shrink-0">
                            {{ Number(entry.item.extra.score).toFixed(3) }}
                        </span>
                        <span v-if="entry.item.box"
                            class="tabular-nums text-dimmed truncate flex-1 text-right">
                            [{{ entry.item.box.x }}, {{ entry.item.box.y }}, {{ entry.item.box.w }}, {{ entry.item.box.h }}]
                        </span>
                    </label>
                    <div v-if="filteredCustomEntries.length === 0"
                        class="text-xs text-dimmed text-center py-3">
                        No matching results
                    </div>
                </div>
            </div>

            <!-- Resize handle -->
            <div class="custom-panel-handle group"
                :class="isNarrow ? 'custom-panel-handle-h' : 'custom-panel-handle-v'"
                @mousedown.prevent="onResizeStart">
                <div class="custom-panel-handle-bar"
                    :class="isNarrow ? 'w-8 h-0.5' : 'h-8 w-0.5'" />
            </div>

            <!-- Right / Bottom: detail panel -->
            <div class="custom-panel-detail overflow-y-auto">
                <div v-if="customDetailItem" class="p-2.5 flex flex-col gap-2 text-xs">
                    <div class="flex items-center justify-between">
                        <span class="font-semibold text-sm">
                            #{{ customSelectedDetail }}
                        </span>
                        <UButton size="xs" variant="ghost" color="neutral"
                            icon="i-lucide-copy" @click="copyResultJson(customSelectedDetail)" />
                    </div>
                    <div v-if="customDetailItem.box"
                        class="flex gap-x-3 gap-y-1 flex-wrap tabular-nums text-default/70">
                        <span><span class="text-default/40">x:</span> {{ customDetailItem.box.x }}</span>
                        <span><span class="text-default/40">y:</span> {{ customDetailItem.box.y }}</span>
                        <span><span class="text-default/40">w:</span> {{ customDetailItem.box.w }}</span>
                        <span><span class="text-default/40">h:</span> {{ customDetailItem.box.h }}</span>
                    </div>
                    <template v-if="customDetailItem.extra">
                        <USeparator />
                        <div v-for="(val, key) in customDetailItem.extra" :key="String(key)"
                            class="flex justify-between gap-2 text-default/70">
                            <span class="text-default/40 shrink-0">{{ key }}</span>
                            <span class="text-right break-all">{{ formatExtraValue(val) }}</span>
                        </div>
                    </template>
                    <!-- Cropped image preview -->
                    <img v-if="croppedImageUrl" :src="croppedImageUrl"
                        class="rounded border border-default bg-muted max-h-32 object-contain w-full mt-1"
                        draggable="false" />
                </div>
                <div v-else class="flex flex-col items-center justify-center h-full text-dimmed text-xs gap-1 p-4">
                    <UIcon name="i-lucide-pointer" class="size-5" />
                    <span>Select an item to view details</span>
                </div>
            </div>
        </div>

        <!-- Canvas container -->
        <div ref="containerRef" class="relative overflow-hidden rounded-lg border border-default bg-muted"
            :style="containerStyle" @wheel.prevent="onWheel">
            <div v-if="rawImage" class="absolute inset-0 flex items-center justify-center select-none"
                :class="isDragging ? 'cursor-grabbing' : 'cursor-grab'" @mousedown="onDragStart" @mousemove="onDragMove"
                @mouseup="onDragEnd" @mouseleave="onDragEnd">
                <!-- Canvas with image and boxes: wrapper sizes to fit canvas -->
                <div ref="canvasWrapperRef" class="relative" :style="canvasWrapperStyle">
                    <canvas ref="canvasRef" class="pointer-events-none block"
                        :style="canvasElStyle" />
                    <!-- Interactive overlay for hover detection -->
                    <div class="absolute inset-0">
                        <div v-for="overlay in visibleOverlays" :key="overlay.resultIdx"
                            class="absolute transition-opacity duration-100 cursor-pointer"
                            :style="overlayBoxStyle(overlay)"
                            @mouseenter.stop="onOverlayEnter(overlay.resultIdx, $event)"
                            @mouseleave.stop="onOverlayLeave"
                            @mousedown="onOverlayMouseDown"
                            @mousemove.stop
                            @click.stop="copyResultJson(overlay.resultIdx)">
                        </div>
                    </div>
                </div>
            </div>
            <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                <UIcon name="i-lucide-image-off" class="size-10" />
                <span class="text-xs">No raw image available</span>
                <span class="text-xs text-dimmed">Enable debug mode or save_draw in pipeline config</span>
            </div>

            <!-- Tooltip popup -->
            <Teleport to="body">
                <Transition name="reco-tooltip">
                    <div v-if="tooltipVisible && tooltipData" class="reco-tooltip-popup" :style="tooltipStyle">
                        <div class="flex flex-col gap-1.5 text-xs leading-relaxed">
                            <div class="font-semibold" :style="{ color: tooltipData.color }">
                                #{{ tooltipData.idx }}
                            </div>
                            <div v-if="tooltipData.box" class="tabular-nums text-default/70">
                                Box: [{{ tooltipData.box.x }}, {{ tooltipData.box.y }},
                                {{ tooltipData.box.w }}, {{ tooltipData.box.h }}]
                            </div>
                            <template v-if="tooltipData.extra">
                                <div v-for="(val, key) in tooltipData.extra" :key="String(key)"
                                    class="text-default/70">
                                    <span class="text-default/50">{{ key }}:</span>
                                    {{ formatExtraValue(val) }}
                                </div>
                            </template>
                            <div class="flex items-center gap-1 text-default/40 pt-0.5 border-t border-default/10">
                                <UIcon name="i-lucide-copy" class="size-3" />
                                <span>Click to copy JSON</span>
                            </div>
                        </div>
                    </div>
                </Transition>
            </Teleport>
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
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import type { RecoDetailResponse, RecoResultItem, RectResponse } from './types'

// --- Constants ---
const MIN_ZOOM = 0.5
const MAX_ZOOM = 10
const ZOOM_STEP = 0.15
const LABEL_COLLISION_MARGIN = 4

const BOX_COLORS = [
    '#22c55e', '#3b82f6', '#f59e0b', '#ef4444',
    '#8b5cf6', '#ec4899', '#14b8a6', '#f97316',
]

// Zoom thresholds for progressive label detail
const ZOOM_LEVEL_MINIMAL = 0.8   // below: hide most labels
const ZOOM_LEVEL_SHORT = 1.5     // below: short "#N" labels
const ZOOM_LEVEL_MEDIUM = 3.0    // below: "#N score"
// above: full detail

const props = defineProps<{
    detail: RecoDetailResponse
    fullscreen?: boolean
}>()

type DrawMode = 'best' | 'filtered' | 'all' | 'custom'

interface LabelLayout {
    resultIdx: number
    x: number
    y: number
    w: number
    h: number
    text: string
    color: string
    visible: boolean
    priority: number
}

interface OverlayInfo {
    resultIdx: number
    box: RectResponse
    color: string
}

interface TooltipData {
    idx: number
    color: string
    box?: RectResponse
    extra?: Record<string, unknown>
}

const drawMode = ref<DrawMode>('best')
const customSelection = ref<Set<number>>(new Set())
const canvasRef = ref<HTMLCanvasElement | null>(null)
const canvasWrapperRef = ref<HTMLElement | null>(null)
const zoomLevel = ref(1)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const panOffset = ref({ x: 0, y: 0 })
const rawImageObj = ref<HTMLImageElement | null>(null)

const hoveredIndex = ref(-1)
const tooltipVisible = ref(false)
const tooltipData = ref<TooltipData | null>(null)
const tooltipPos = ref({ x: 0, y: 0 })
const toast = useToast()

// Custom panel state
const customSearch = ref('')
const customSelectedDetail = ref(-1)
const isNarrow = ref(false)
const splitRatio = ref(0.45)
const isResizing = ref(false)
const rootRef = ref<HTMLElement | null>(null)
const containerRef = ref<HTMLElement | null>(null)
const resizeContainerRect = ref<DOMRect | null>(null)

const filteredCustomEntries = computed(() => {
    const q = customSearch.value.toLowerCase().trim()
    return allResults.value
        .map((item, idx) => ({ item, idx }))
        .filter(({ item, idx }) => {
            if (!q) return true
            if (String(idx).includes(q)) return true
            if (item.box && `${item.box.x},${item.box.y},${item.box.w},${item.box.h}`.includes(q)) return true
            const extra = getExtraLabel(item).toLowerCase()
            return extra.includes(q)
        })
})

const customDetailItem = computed<RecoResultItem | null>(() => {
    if (customSelectedDetail.value < 0) return null
    return allResults.value[customSelectedDetail.value] ?? null
})

// --- Cropped image preview (cached) ---
const cropCache = new Map<string, string>()
const cropCanvas = document.createElement('canvas')

function getCropCacheKey(idx: number, box: RectResponse): string {
    return `${idx}:${box.x},${box.y},${box.w},${box.h}`
}

const croppedImageUrl = computed<string | null>(() => {
    const item = customDetailItem.value
    const img = rawImageObj.value
    if (!item?.box || !img) return null

    const key = getCropCacheKey(customSelectedDetail.value, item.box)
    const cached = cropCache.get(key)
    if (cached) return cached

    const { x, y, w, h } = item.box
    if (w <= 0 || h <= 0) return null

    const sx = Math.max(0, Math.min(x, img.naturalWidth))
    const sy = Math.max(0, Math.min(y, img.naturalHeight))
    const sw = Math.min(w, img.naturalWidth - sx)
    const sh = Math.min(h, img.naturalHeight - sy)
    if (sw <= 0 || sh <= 0) return null

    cropCanvas.width = sw
    cropCanvas.height = sh
    const ctx = cropCanvas.getContext('2d')
    if (!ctx) return null

    ctx.drawImage(img, sx, sy, sw, sh, 0, 0, sw, sh)
    const url = cropCanvas.toDataURL('image/png')
    cropCache.set(key, url)
    return url
})

const listPanelStyle = computed(() => {
    if (isNarrow.value) {
        return { height: `${splitRatio.value * 100}%` }
    }
    return { width: `${splitRatio.value * 100}%` }
})

function selectAllCustom() {
    const newSet = new Set<number>()
    for (let i = 0; i < allResults.value.length; i++) newSet.add(i)
    customSelection.value = newSet
}

function deselectAllCustom() {
    customSelection.value = new Set()
}

// --- Resize handle ---
function onResizeStart(e: MouseEvent) {
    const root = (e.target as HTMLElement).closest('.custom-panel-root') as HTMLElement | null
    if (!root) return
    resizeContainerRect.value = root.getBoundingClientRect()
    isResizing.value = true
    document.addEventListener('mousemove', onResizeMove)
    document.addEventListener('mouseup', onResizeEnd)
}

function onResizeMove(e: MouseEvent) {
    if (!isResizing.value || !resizeContainerRect.value) return
    const rect = resizeContainerRect.value
    let ratio: number
    if (isNarrow.value) {
        ratio = (e.clientY - rect.top) / rect.height
    } else {
        ratio = (e.clientX - rect.left) / rect.width
    }
    splitRatio.value = Math.max(0.2, Math.min(0.7, ratio))
}

function onResizeEnd() {
    isResizing.value = false
    resizeContainerRect.value = null
    document.removeEventListener('mousemove', onResizeMove)
    document.removeEventListener('mouseup', onResizeEnd)
}

// --- Responsive: detect narrow width ---
let resizeObserver: ResizeObserver | null = null
const customPanelBreakpoint = 420

function setupResizeObserver() {
    if (typeof ResizeObserver === 'undefined') return
    resizeObserver = new ResizeObserver((entries) => {
        for (const entry of entries) {
            isNarrow.value = entry.contentRect.width < customPanelBreakpoint
        }
    })
}

const rawImage = computed(() => props.detail.raw_image)
const results = computed(() => props.detail.results)

const imgWidth = computed(() => rawImageObj.value?.naturalWidth ?? 0)
const imgHeight = computed(() => rawImageObj.value?.naturalHeight ?? 0)
const imgAspect = computed(() => imgWidth.value && imgHeight.value ? imgWidth.value / imgHeight.value : 16 / 9)

const containerStyle = computed(() => {
    if (props.fullscreen) {
        return { width: '100%', height: '100%', flex: '1 1 0', minHeight: '0' }
    }
    return { aspectRatio: `${imgAspect.value}`, maxHeight: '60vh' }
})

// Explicit canvas display size: object-fit contain logic via CSS
const canvasElStyle = computed(() => {
    if (!imgWidth.value || !imgHeight.value) return {}
    return {
        width: '100%',
        height: '100%',
        objectFit: 'contain' as const,
    }
})

const canvasWrapperStyle = computed(() => {
    const base: Record<string, string> = {
        maxWidth: '100%',
        maxHeight: '100%',
        aspectRatio: `${imgWidth.value} / ${imgHeight.value}`,
        transform: `translate(${panOffset.value.x}px, ${panOffset.value.y}px) scale(${zoomLevel.value})`,
        transformOrigin: 'center center',
        transition: isDragging.value ? 'none' : 'transform 0.15s ease-out',
    }
    return base
})

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

const activeResults = computed<RecoResultItem[]>(() => {
    switch (drawMode.value) {
        case 'best': return bestResults.value
        case 'filtered': return filteredResults.value
        case 'all': return allResults.value
        case 'custom': return allResults.value.filter((_, idx) => customSelection.value.has(idx))
    }
})

const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))

// Compute overlay data for hover-interactive regions
const visibleOverlays = computed<OverlayInfo[]>(() => {
    return activeResults.value
        .map((item, idx) => {
            if (!item.box) return null
            return {
                resultIdx: idx,
                box: item.box,
                color: getBoxColor(idx),
            }
        })
        .filter((v): v is OverlayInfo => v !== null)
})

// --- Draw mode ---
function setDrawMode(mode: DrawMode) {
    drawMode.value = mode
    if (mode === 'custom' && customSelection.value.size === 0) {
        const newSet = new Set<number>()
        for (let i = 0; i < allResults.value.length; i++) newSet.add(i)
        customSelection.value = newSet
    }
}

function toggleCustom(idx: number) {
    const newSet = new Set(customSelection.value)
    if (newSet.has(idx)) newSet.delete(idx)
    else newSet.add(idx)
    customSelection.value = newSet
}

function getExtraLabel(item: RecoResultItem): string {
    if (!item.extra) return ''
    const parts: string[] = []
    if ('text' in item.extra) parts.push(`"${item.extra.text}"`)
    if ('score' in item.extra) parts.push(`score: ${Number(item.extra.score).toFixed(3)}`)
    if ('count' in item.extra) parts.push(`count: ${item.extra.count}`)
    if ('label' in item.extra) parts.push(`label: ${item.extra.label}`)
    return parts.join(', ')
}

function formatExtraValue(val: unknown): string {
    if (typeof val === 'number') return val % 1 === 0 ? String(val) : Number(val).toFixed(4)
    if (typeof val === 'string') return val
    return JSON.stringify(val)
}

// --- Overlay positioning (percentage-based relative to canvas) ---
function overlayBoxStyle(overlay: OverlayInfo) {
    if (!imgWidth.value || !imgHeight.value) return { display: 'none' }
    const { x, y, w, h } = overlay.box
    return {
        left: `${(x / imgWidth.value) * 100}%`,
        top: `${(y / imgHeight.value) * 100}%`,
        width: `${(w / imgWidth.value) * 100}%`,
        height: `${(h / imgHeight.value) * 100}%`,
        pointerEvents: 'auto' as const,
        cursor: 'default',
    }
}

// --- Hover / Tooltip ---
function onOverlayEnter(resultIdx: number, e: MouseEvent) {
    hoveredIndex.value = resultIdx
    const item = activeResults.value[resultIdx]
    if (!item) return
    tooltipData.value = {
        idx: resultIdx,
        color: getBoxColor(resultIdx),
        box: item.box,
        extra: item.extra,
    }
    tooltipPos.value = { x: e.clientX, y: e.clientY }
    tooltipVisible.value = true
    nextTick(() => drawCanvas())
}

function onOverlayLeave() {
    hoveredIndex.value = -1
    tooltipVisible.value = false
    tooltipData.value = null
    nextTick(() => drawCanvas())
}

function onOverlayMouseDown(e: MouseEvent) {
    // Only stop propagation for left click (copy action);
    // let other buttons bubble for potential drag
    if (e.button === 0) {
        e.stopPropagation()
    }
}

function copyResultJson(resultIdx: number) {
    const item = activeResults.value[resultIdx]
    if (!item) return
    const obj: Record<string, unknown> = {}
    if (item.box) obj.box = item.box
    if (item.extra) obj.extra = item.extra
    const json = JSON.stringify(obj, null, 2)
    navigator.clipboard.writeText(json).then(() => {
        toast.add({ id: 'reco-copy', title: `#${resultIdx} JSON copied`, icon: 'i-lucide-check', color: 'success' })
    })
}

const tooltipStyle = computed(() => {
    return {
        position: 'fixed' as const,
        left: `${tooltipPos.value.x + 12}px`,
        top: `${tooltipPos.value.y + 12}px`,
        zIndex: 9999,
    }
})

// --- Label detail level based on zoom and draw mode ---
// all/filtered: canvas labels always show index only; detail via tooltip
// best/custom: progressive detail based on zoom level
function getLabelDetailLevel(): 'hidden' | 'minimal' | 'short' | 'medium' | 'full' {
    const mode = drawMode.value
    if (mode === 'all' || mode === 'filtered') {
        const z = zoomLevel.value
        if (z < ZOOM_LEVEL_MINIMAL && activeResults.value.length > 5) return 'minimal'
        return 'short'
    }
    const z = zoomLevel.value
    const count = activeResults.value.length
    if (count <= 3) {
        if (z < ZOOM_LEVEL_MINIMAL) return 'short'
        if (z < ZOOM_LEVEL_SHORT) return 'medium'
        return 'full'
    }
    if (z < ZOOM_LEVEL_MINIMAL) return 'minimal'
    if (z < ZOOM_LEVEL_SHORT) return 'short'
    if (z < ZOOM_LEVEL_MEDIUM) return 'medium'
    return 'full'
}

function buildLabelByLevel(item: RecoResultItem, idx: number, level: string): string {
    if (level === 'hidden') return ''
    if (level === 'minimal') return `${idx}`
    if (level === 'short') return `#${idx}`

    const algo = props.detail.algorithm?.toLowerCase() ?? ''
    const isOcr = algo.includes('ocr')

    if (level === 'medium') {
        const parts = [`#${idx}`]
        if (item.extra) {
            if ('score' in item.extra) parts.push(Number(item.extra.score).toFixed(2))
            if (isOcr && 'text' in item.extra) parts.push(`"${item.extra.text}"`)
        }
        return parts.join(' ')
    }
    // full
    const parts: string[] = [`#${idx}`]
    if (item.extra) {
        if ('score' in item.extra) parts.push(Number(item.extra.score).toFixed(3))
        if (isOcr && 'text' in item.extra) parts.push(`"${item.extra.text}"`)
        if ('count' in item.extra) parts.push(`n=${item.extra.count}`)
        if ('label' in item.extra) parts.push(`${item.extra.label}`)
    }
    return parts.join(' ')
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

    ctx.drawImage(img, 0, 0)

    const items = activeResults.value
    const isHovering = hoveredIndex.value >= 0
    const baseFontSize = Math.max(12, Math.min(img.naturalWidth, img.naturalHeight) / 60)
    const labelLevel = getLabelDetailLevel()
    const fontSize = labelLevel === 'full' ? baseFontSize : baseFontSize * 0.7

    // First pass: draw all boxes (dimmed if hovering a different one)
    items.forEach((item, idx) => {
        if (!item.box) return
        const color = getBoxColor(idx)
        const { x, y, w, h } = item.box
        const isThisHovered = hoveredIndex.value === idx
        const dimmed = isHovering && !isThisHovered

        ctx.strokeStyle = dimmed ? color + '60' : color
        ctx.lineWidth = isThisHovered
            ? Math.max(3, Math.min(img.naturalWidth, img.naturalHeight) / 200)
            : Math.max(2, Math.min(img.naturalWidth, img.naturalHeight) / 300)
        ctx.setLineDash([])
        ctx.strokeRect(x, y, w, h)

        ctx.fillStyle = isThisHovered ? color + '35' : (dimmed ? color + '10' : color + '20')
        ctx.fillRect(x, y, w, h)
    })

    // Second pass: compute label layouts with collision detection
    const labelLayouts: LabelLayout[] = []
    const effectiveFontSize = fontSize
    ctx.font = `bold ${effectiveFontSize}px system-ui, sans-serif`

    items.forEach((item, idx) => {
        if (!item.box) return
        const color = getBoxColor(idx)
        const bx = item.box.x
        const by = item.box.y
        const bh = item.box.h
        const isThisHovered = hoveredIndex.value === idx

        const text = isThisHovered
            ? buildLabelByLevel(item, idx, 'full')
            : buildLabelByLevel(item, idx, labelLevel)

        if (!text) return

        const metrics = ctx.measureText(text)
        const textH = effectiveFontSize * 1.3
        const padding = effectiveFontSize * 0.3
        const lw = metrics.width + padding * 2
        const lh = textH + padding

        const lx = bx
        const ly = by > lh + 2 ? by - lh - 2 : by + bh

        // Priority: best > filtered > lower index; hovered always highest
        let priority = 1000 - idx
        if (isThisHovered) priority = 10000
        if (drawMode.value === 'best') priority += 5000
        if (drawMode.value === 'filtered') priority += 3000

        labelLayouts.push({
            resultIdx: idx, x: lx, y: ly, w: lw, h: lh,
            text, color, visible: true, priority,
        })
    })

    // Sort by priority descending; higher priority labels are placed first
    labelLayouts.sort((a, b) => b.priority - a.priority)

    // Collision detection: hide lower priority labels that overlap higher ones
    const placed: LabelLayout[] = []
    for (const label of labelLayouts) {
        const collides = placed.some(p => rectsOverlap(
            label.x - LABEL_COLLISION_MARGIN, label.y - LABEL_COLLISION_MARGIN,
            label.w + LABEL_COLLISION_MARGIN * 2, label.h + LABEL_COLLISION_MARGIN * 2,
            p.x, p.y, p.w, p.h
        ))
        if (collides && hoveredIndex.value !== label.resultIdx) {
            label.visible = false
        } else {
            placed.push(label)
        }
    }

    // Third pass: draw visible labels
    for (const label of labelLayouts) {
        if (!label.visible) continue
        const isThisHovered = hoveredIndex.value === label.resultIdx
        const dimmed = isHovering && !isThisHovered

        const drawFontSize = isThisHovered ? baseFontSize : effectiveFontSize
        ctx.font = `bold ${drawFontSize}px system-ui, sans-serif`

        const finalMetrics = ctx.measureText(label.text)
        const finalW = finalMetrics.width + (drawFontSize * 0.3) * 2
        const finalH = drawFontSize * 1.3 + drawFontSize * 0.3
        const pad = drawFontSize * 0.3

        ctx.fillStyle = dimmed ? label.color + '80' : label.color + 'DD'
        ctx.fillRect(label.x, label.y, finalW, finalH)

        ctx.fillStyle = dimmed ? '#ffffffA0' : '#ffffff'
        ctx.textBaseline = 'top'
        ctx.fillText(label.text, label.x + pad, label.y + pad * 0.5)
    }
}

function rectsOverlap(
    x1: number, y1: number, w1: number, h1: number,
    x2: number, y2: number, w2: number, h2: number,
): boolean {
    return x1 < x2 + w2 && x1 + w1 > x2 && y1 < y2 + h2 && y1 + h1 > y2
}

function getBoxColor(idx: number): string {
    if (drawMode.value === 'best') return BOX_COLORS[0]!
    if (drawMode.value === 'filtered') return BOX_COLORS[1]!
    return BOX_COLORS[idx % BOX_COLORS.length]!
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
        nextTick(() => drawCanvas())
    }
    img.src = rawImage.value
}

// --- Zoom ---
function zoomIn() {
    zoomLevel.value = Math.min(MAX_ZOOM, +(zoomLevel.value + ZOOM_STEP).toFixed(2))
}

function zoomOut() {
    zoomLevel.value = Math.max(MIN_ZOOM, +(zoomLevel.value - ZOOM_STEP).toFixed(2))
    if (zoomLevel.value <= 1) panOffset.value = { x: 0, y: 0 }
}

function resetView() {
    zoomLevel.value = 1
    panOffset.value = { x: 0, y: 0 }
}

function onWheel(e: WheelEvent) {
    if (!rawImage.value) return
    if (e.deltaY < 0) zoomIn()
    else zoomOut()
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
    cropCache.clear()
    loadRawImage()
    resetView()
})

watch(activeResults, () => {
    nextTick(() => drawCanvas())
}, { deep: true })

watch(drawMode, () => {
    nextTick(() => drawCanvas())
})

watch(zoomLevel, () => {
    nextTick(() => drawCanvas())
})

onMounted(() => {
    loadRawImage()
    setupResizeObserver()
    if (rootRef.value && resizeObserver) {
        resizeObserver.observe(rootRef.value)
    }
})

onBeforeUnmount(() => {
    resizeObserver?.disconnect()
    document.removeEventListener('mousemove', onResizeMove)
    document.removeEventListener('mouseup', onResizeEnd)
})
</script>

<style scoped>
.reco-tooltip-popup {
    background: var(--ui-bg-elevated, #1e1e2e);
    border: 1px solid var(--ui-border-muted, #333);
    border-radius: 8px;
    padding: 8px 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
    max-width: 320px;
    pointer-events: none;
}

.reco-tooltip-enter-active {
    transition: opacity 0.12s ease-out, transform 0.12s ease-out;
}

.reco-tooltip-leave-active {
    transition: opacity 0.08s ease-in;
}

.reco-tooltip-enter-from {
    opacity: 0;
    transform: translateY(4px);
}

.reco-tooltip-leave-to {
    opacity: 0;
}

/* Custom panel split layout */
.custom-panel-root {
    min-height: 0;
}

.custom-panel-horizontal {
    display: flex;
    flex-direction: row;
    height: 220px;
}

.custom-panel-vertical {
    display: flex;
    flex-direction: column;
    height: 280px;
}

.custom-panel-list {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
}

.custom-panel-detail {
    flex: 1;
    min-width: 0;
    min-height: 0;
}

.custom-panel-handle {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: background 0.15s;
    z-index: 1;
}

.custom-panel-handle:hover,
.custom-panel-handle:active {
    background: var(--ui-bg-muted, rgba(255, 255, 255, 0.06));
}

.custom-panel-handle-v {
    width: 8px;
    cursor: col-resize;
    border-left: 1px solid var(--ui-border-default, #333);
    border-right: 1px solid var(--ui-border-default, #333);
}

.custom-panel-handle-h {
    height: 8px;
    cursor: row-resize;
    border-top: 1px solid var(--ui-border-default, #333);
    border-bottom: 1px solid var(--ui-border-default, #333);
}

.custom-panel-handle-bar {
    border-radius: 9999px;
    background: var(--ui-text-dimmed, #666);
    opacity: 0.4;
    transition: opacity 0.15s;
}

.custom-panel-handle:hover .custom-panel-handle-bar {
    opacity: 0.8;
}
</style>
