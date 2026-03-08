<template>
    <div :class="['reco-root', fullscreen ? 'h-full' : '']">
        <!-- LEFT PANEL: controls + result list + detail -->
        <div class="reco-left">
            <!-- Draw mode selector -->
            <div class="flex flex-row items-center gap-2 flex-wrap px-1">
                <span class="text-xs text-dimmed font-medium">Draw:</span>
                <UTabs :items="drawModeOptions" key="value" v-model="drawMode" />
                <span class="text-xs text-dimmed tabular-nums">
                    ({{ activeResults.length }})
                </span>
            </div>

            <!-- Custom: search + select/deselect -->
            <div v-if="drawMode === 'custom'" class="flex items-center gap-1.5 px-1">
                <UInput v-model="customSearch" icon="i-lucide-search" size="xs" placeholder="Filter..."
                    class="flex-1" />
                <UTooltip text="Select all">
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-check-check"
                        @click="selectAllCustom" />
                </UTooltip>
                <UTooltip text="Deselect all">
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-x" @click="deselectAllCustom" />
                </UTooltip>
            </div>

            <!-- Result list -->
            <div v-if="showResultList" class="reco-list">
                <template v-if="drawMode === 'custom'">
                    <label v-for="entry in filteredCustomEntries" :key="entry.idx" class="reco-list-item" :class="[
                        focusedResultIndex === entry.idx ? 'bg-primary/10 text-primary' : 'hover:bg-muted',
                        hoveredIndex === entry.idx ? 'ring-1 ring-primary/40' : ''
                    ]" @mouseenter="hoveredIndex = entry.idx" @mouseleave="hoveredIndex = -1"
                        @click.stop="setFocusedResultIndex(entry.idx)">
                        <input type="checkbox" :checked="customSelection.has(entry.idx)" class="accent-primary shrink-0"
                            @change.stop="toggleCustom(entry.idx)" @click.stop />
                        <span class="tabular-nums text-dimmed shrink-0">#{{ entry.idx }}</span>
                        <span v-if="entry.item.extra && 'score' in entry.item.extra" class="tabular-nums shrink-0">
                            {{ Number(entry.item.extra.score).toFixed(3) }}
                        </span>
                        <span v-if="entry.item.box" class="tabular-nums text-dimmed truncate flex-1 text-right">
                            [{{ entry.item.box.x }}, {{ entry.item.box.y }}, {{ entry.item.box.w }}, {{ entry.item.box.h
                            }}]
                        </span>
                    </label>
                </template>
                <template v-else>
                    <div v-for="(item, idx) in activeResults" :key="idx" class="reco-list-item" :class="[
                        focusedResultIndex === idx ? 'bg-primary/10 text-primary' : 'hover:bg-muted',
                        hoveredIndex === idx ? 'ring-1 ring-primary/40' : ''
                    ]" @mouseenter="hoveredIndex = idx" @mouseleave="hoveredIndex = -1"
                        @click.stop="setFocusedResultIndex(idx)">
                        <span class="tabular-nums text-dimmed shrink-0">#{{ idx }}</span>
                        <span v-if="item.extra && 'score' in item.extra" class="tabular-nums shrink-0">
                            {{ Number(item.extra.score).toFixed(3) }}
                        </span>
                        <span v-if="item.box" class="tabular-nums text-dimmed truncate flex-1 text-right">
                            [{{ item.box.x }}, {{ item.box.y }}, {{ item.box.w }}, {{ item.box.h }}]
                        </span>
                    </div>
                </template>
                <div v-if="(drawMode === 'custom' ? filteredCustomEntries.length : activeResults.length) === 0"
                    class="text-xs text-dimmed text-center py-3">
                    No results
                </div>
            </div>

            <!-- Detail panel -->
            <div class="reco-detail">
                <div v-if="focusedDetailItem" class="flex flex-col gap-2 text-xs">
                    <div class="flex items-center justify-between gap-2">
                        <div class="flex items-center gap-2 min-w-0">
                            <UBadge color="neutral" variant="subtle" size="xs">{{ activeModeLabel }}</UBadge>
                            <span class="font-semibold text-sm tabular-nums">#{{ focusedResultIndex }}</span>
                        </div>
                        <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-copy"
                            @click="copyResultJson(focusedResultIndex)" />
                    </div>
                    <div v-if="focusedDetailItem.box"
                        class="flex gap-x-3 gap-y-1 flex-wrap tabular-nums text-default/70">
                        <span><span class="text-default/40">x:</span> {{ focusedDetailItem.box.x }}</span>
                        <span><span class="text-default/40">y:</span> {{ focusedDetailItem.box.y }}</span>
                        <span><span class="text-default/40">w:</span> {{ focusedDetailItem.box.w }}</span>
                        <span><span class="text-default/40">h:</span> {{ focusedDetailItem.box.h }}</span>
                    </div>
                    <template v-if="focusedDetailItem.extra">
                        <USeparator />
                        <div v-for="(val, key) in focusedDetailItem.extra" :key="String(key)"
                            class="flex justify-between gap-2 text-default/70">
                            <span class="text-default/40 shrink-0">{{ key }}</span>
                            <span class="text-right break-all">{{ formatExtraValue(val) }}</span>
                        </div>
                    </template>
                    <img v-if="croppedImageUrl" :src="croppedImageUrl"
                        class="rounded border border-default bg-muted max-h-28 object-contain w-full mt-1"
                        draggable="false" />
                </div>
                <div v-else class="flex flex-col items-center justify-center h-full text-dimmed text-xs gap-1">
                    <UIcon name="i-lucide-pointer" class="size-5" />
                    <span>{{ emptyDetailText }}</span>
                </div>
            </div>
        </div>

        <!-- RIGHT PANEL: canvas + toolbar -->
        <div class="reco-right">
            <div ref="containerRef" class="reco-canvas-container" :style="containerStyle" @wheel.prevent="onWheel">
                <div v-if="rawImage" class="absolute inset-0 flex items-center justify-center select-none"
                    :class="[isDragging ? 'cursor-grabbing' : (hitTestCursor ? 'cursor-pointer' : 'cursor-grab')]"
                    @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp" @mouseleave="onMouseLeave"
                    @click="onCanvasClick">
                    <div class="relative" :style="canvasWrapperStyle">
                        <canvas ref="canvasRef" class="pointer-events-none block w-full h-full"></canvas>
                    </div>
                </div>
                <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                    <UIcon name="i-lucide-image-off" class="size-10" />
                    <span class="text-xs">No raw image available</span>
                </div>

                <!-- Tooltip -->
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
            <div v-if="rawImage" class="flex flex-row items-center gap-2 pt-1">
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
                    <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="xs"
                        @click="downloadCanvas" />
                </UTooltip>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { getTaskImageUrl } from '@/api/http'
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

type DrawMode = 'best' | 'filtered' | 'custom'

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

interface LabelRect {
    resultIdx: number
    x: number
    y: number
    w: number
    h: number
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
const focusedResultIndex = ref(-1)

const drawModeOptions = [
    { label: 'Best', value: 'best' },
    { label: 'Filtered', value: 'filtered' },
    { label: 'Custom', value: 'custom' },
] satisfies Array<{ label: string, value: DrawMode }>

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

const showResultList = computed(() => drawMode.value !== 'best')

const activeModeLabel = computed(() => {
    if (drawMode.value === 'best') return 'Best'
    if (drawMode.value === 'filtered') return 'Filtered'
    return 'Custom'
})

const focusedDetailItem = computed<RecoResultItem | null>(() => {
    if (focusedResultIndex.value < 0) return null
    return activeResults.value[focusedResultIndex.value] ?? null
})

const emptyDetailText = computed(() => {
    if (drawMode.value === 'best') return 'No best result'
    return 'Select an item'
})

// --- Cropped image preview (cached) ---
const cropCache = new Map<string, string>()
const cropCanvas = document.createElement('canvas')

function getCropCacheKey(idx: number, box: RectResponse): string {
    return `${idx}:${box.x},${box.y},${box.w},${box.h}`
}

function buildCroppedUrl(item: RecoResultItem, img: HTMLImageElement, idx: number): string | null {
    if (!item.box) return null

    const key = getCropCacheKey(idx, item.box)
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
}

const croppedImageUrl = computed<string | null>(() => {
    const item = focusedDetailItem.value
    const img = rawImageObj.value
    if (!item?.box || !img || focusedResultIndex.value < 0) return null
    return buildCroppedUrl(item, img, focusedResultIndex.value)
})

function selectAllCustom() {
    const newSet = new Set<number>()
    for (let i = 0; i < allResults.value.length; i++) newSet.add(i)
    customSelection.value = newSet
}

function deselectAllCustom() {
    customSelection.value = new Set()
}

const rawImage = computed(() => props.detail.raw_image)
const results = computed(() => props.detail.results)

const imgWidth = computed(() => rawImageObj.value?.naturalWidth ?? 0)
const imgHeight = computed(() => rawImageObj.value?.naturalHeight ?? 0)
const containerStyle = computed(() => {
    if (props.fullscreen) {
        return { flex: '1 1 0', minHeight: '0' }
    }
    return { maxHeight: '60vh' }
})

const canvasWrapperStyle = computed(() => ({
    maxWidth: '100%',
    maxHeight: '100%',
    aspectRatio: `${imgWidth.value} / ${imgHeight.value}`,
    transform: `translate(${panOffset.value.x}px, ${panOffset.value.y}px) scale(${zoomLevel.value})`,
    transformOrigin: 'center center',
    transition: isDragging.value ? 'none' : 'transform 0.15s ease-out',
}))

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
        case 'custom': return allResults.value.filter((_, idx) => customSelection.value.has(idx))
        default: return []
    }
})

const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))

// Compute overlay data for hover-interactive regions
// Label rects from last draw, used for hit testing
let lastLabelRects: LabelRect[] = []

function ensureCustomSelection() {
    if (customSelection.value.size > 0) return
    const newSet = new Set<number>()
    for (let i = 0; i < allResults.value.length; i++) newSet.add(i)
    customSelection.value = newSet
}

function setFocusedResultIndex(nextIndex: number) {
    focusedResultIndex.value = nextIndex >= 0 && nextIndex < activeResults.value.length ? nextIndex : -1
}

function syncFocusedResultIndex() {
    if (drawMode.value === 'best') {
        setFocusedResultIndex(activeResults.value.length > 0 ? 0 : -1)
        return
    }

    if (drawMode.value === 'custom') {
        ensureCustomSelection()
    }

    if (activeResults.value.length === 0) {
        setFocusedResultIndex(-1)
        return
    }

    if (focusedResultIndex.value < 0 || focusedResultIndex.value >= activeResults.value.length) {
        setFocusedResultIndex(0)
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

// --- Canvas hit-testing: convert mouse coords to image coords ---
function mouseToImageCoords(e: MouseEvent): { ix: number, iy: number } | null {
    const canvas = canvasRef.value
    if (!canvas || !imgWidth.value || !imgHeight.value) return null

    const rect = canvas.getBoundingClientRect()
    const scaleX = imgWidth.value / rect.width
    const scaleY = imgHeight.value / rect.height
    return {
        ix: (e.clientX - rect.left) * scaleX,
        iy: (e.clientY - rect.top) * scaleY,
    }
}

// Distance from point to the nearest edge of a rect (0 if on the edge)
function edgeDist(ix: number, iy: number, x: number, y: number, w: number, h: number): number {
    const dl = Math.abs(ix - x)
    const dr = Math.abs(ix - (x + w))
    const dt = Math.abs(iy - y)
    const db = Math.abs(iy - (y + h))
    return Math.min(dl, dr, dt, db)
}

// Hit test: prioritize label rects first, then the box whose edge is closest
function hitTest(ix: number, iy: number): number {
    // Label rects have highest priority (always unambiguous)
    for (const lr of lastLabelRects) {
        if (ix >= lr.x && ix <= lr.x + lr.w && iy >= lr.y && iy <= lr.y + lr.h) {
            return lr.resultIdx
        }
    }

    // Collect all boxes that contain the point
    const hits: { idx: number, dist: number }[] = []
    const items = activeResults.value
    items.forEach((item, idx) => {
        if (!item.box) return
        const { x, y, w, h } = item.box
        if (ix >= x && ix <= x + w && iy >= y && iy <= y + h) {
            hits.push({ idx, dist: edgeDist(ix, iy, x, y, w, h) })
        }
    })

    if (hits.length === 0) return -1
    // Pick the one whose edge is closest — allows selecting both inner and outer boxes
    hits.sort((a, b) => a.dist - b.dist)
    return hits[0]!.idx
}

const hitTestCursor = computed(() => hoveredIndex.value >= 0)

// --- Unified mouse handlers ---
const dragThreshold = 4
const mouseDownPos = ref<{ x: number, y: number } | null>(null)
let didDrag = false

function onMouseDown(e: MouseEvent) {
    mouseDownPos.value = { x: e.clientX, y: e.clientY }
    didDrag = false
    isDragging.value = true
    dragStart.value = { x: e.clientX, y: e.clientY }
    dragOffset.value = { ...panOffset.value }
}

function onMouseMove(e: MouseEvent) {
    // Drag detection
    if (isDragging.value && mouseDownPos.value) {
        const dx = e.clientX - mouseDownPos.value.x
        const dy = e.clientY - mouseDownPos.value.y
        if (Math.abs(dx) > dragThreshold || Math.abs(dy) > dragThreshold) {
            didDrag = true
        }
    }

    if (didDrag && isDragging.value) {
        panOffset.value = {
            x: dragOffset.value.x + (e.clientX - dragStart.value.x),
            y: dragOffset.value.y + (e.clientY - dragStart.value.y),
        }
        return
    }

    // Hit test for hover
    const coords = mouseToImageCoords(e)
    if (!coords) return
    const idx = hitTest(coords.ix, coords.iy)
    const prevIdx = hoveredIndex.value
    hoveredIndex.value = idx

    if (idx >= 0) {
        const item = activeResults.value[idx]
        if (item) {
            tooltipData.value = {
                idx,
                color: getBoxColor(idx),
                box: item.box,
                extra: item.extra,
            }
            tooltipPos.value = { x: e.clientX, y: e.clientY }
            tooltipVisible.value = true
        }
    } else {
        tooltipVisible.value = false
        tooltipData.value = null
    }

    if (idx !== prevIdx) nextTick(() => drawCanvas())
}

function onMouseUp() {
    isDragging.value = false
    mouseDownPos.value = null
}

function onMouseLeave() {
    isDragging.value = false
    mouseDownPos.value = null
    if (hoveredIndex.value >= 0) {
        hoveredIndex.value = -1
        tooltipVisible.value = false
        tooltipData.value = null
        nextTick(() => drawCanvas())
    }
}

function onCanvasClick(e: MouseEvent) {
    if (didDrag) return
    const coords = mouseToImageCoords(e)
    if (!coords) return
    const idx = hitTest(coords.ix, coords.iy)
    if (idx >= 0) {
        setFocusedResultIndex(idx)
        copyResultJson(idx)
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

watch(drawMode, (mode) => {
    if (mode === 'custom') {
        ensureCustomSelection()
    }
    syncFocusedResultIndex()
})

watch(activeResults, () => {
    syncFocusedResultIndex()
    nextTick(() => drawCanvas())
}, { deep: true })

// --- Label detail level based on zoom and draw mode ---
// all/filtered: canvas labels always show index only; detail via tooltip
// best/custom: progressive detail based on zoom level
function getLabelDetailLevel(): 'hidden' | 'minimal' | 'short' | 'medium' | 'full' {
    const mode = drawMode.value
    if (mode === 'filtered') {
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

    lastLabelRects = []
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
    ctx.font = `bold ${fontSize}px system-ui, sans-serif`

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
        const textH = fontSize * 1.3
        const padding = fontSize * 0.3
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

        const drawFontSize = isThisHovered ? baseFontSize : fontSize
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

        lastLabelRects.push({ resultIdx: label.resultIdx, x: label.x, y: label.y, w: finalW, h: finalH })
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
    img.src = rawImage.value.url || getTaskImageUrl(rawImage.value.id)
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

watch(zoomLevel, () => {
    nextTick(() => drawCanvas())
})

onMounted(() => {
    syncFocusedResultIndex()
    loadRawImage()
})
</script>

<style scoped>
/* Two-column root layout */
.reco-root {
    display: flex;
    flex-direction: row;
    gap: 12px;
    min-height: 60vh;
}

.reco-left {
    width: 260px;
    min-width: 200px;
    max-width: 320px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow: hidden;
}

.reco-right {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.reco-canvas-container {
    position: relative;
    overflow: hidden;
    border-radius: 8px;
    border: 1px solid var(--ui-border-default, #333);
    background: var(--ui-bg-muted);
    flex: 1;
    min-height: 200px;
}

.reco-list {
    flex: 0 1 auto;
    min-height: 0;
    max-height: 130px;
    overflow-y: auto;
    border: 1px solid var(--ui-border-default, #333);
    border-radius: 8px;
    background: var(--ui-bg-elevated);
    padding: 2px;
}

.reco-list-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.75rem;
    cursor: pointer;
    border-radius: 6px;
    padding: 4px 6px;
    transition: background 0.1s, box-shadow 0.1s;
}

.reco-detail {
    border: 1px solid var(--ui-border-default, #333);
    border-radius: 8px;
    background: var(--ui-bg-elevated);
    padding: 8px;
    overflow-y: auto;
    max-height: 220px;
    flex-shrink: 0;
}

/* Tooltip */
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
</style>
