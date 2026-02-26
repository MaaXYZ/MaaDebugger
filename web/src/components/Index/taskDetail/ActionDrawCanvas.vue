<template>
    <div :class="['action-draw-root', fullscreen ? 'h-full' : '']">
        <!-- LEFT PANEL: action info -->
        <div class="action-draw-left">
            <ActionDetailItem :detail="detail" />
        </div>

        <!-- RIGHT PANEL: canvas + toolbar -->
        <div class="action-draw-right">
            <div ref="containerRef" class="action-draw-canvas-container" :style="containerStyle"
                @wheel.prevent="onWheel">
                <div v-if="rawImage" class="absolute inset-0 flex items-center justify-center select-none"
                    :class="[isDragging ? 'cursor-grabbing' : 'cursor-grab']" @mousedown="onMouseDown"
                    @mousemove="onMouseMove" @mouseup="onMouseUp" @mouseleave="onMouseLeave">
                    <div class="relative" :style="canvasWrapperStyle">
                        <canvas ref="canvasRef" class="pointer-events-none block w-full h-full" />
                    </div>
                </div>
                <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-muted">
                    <UIcon name="i-lucide-image-off" class="size-10" />
                    <span class="text-xs">No raw image available</span>
                </div>
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
import type {
    ActionDetailResponse,
    PointResponse,
    ClickActionResult,
    LongPressActionResult,
    SwipeActionResult,
    MultiSwipeActionResult,
    TouchActionResult,
    ScrollActionResult,
} from './types'
import { actionHasCoords } from './types'
import ActionDetailItem from './ActionDetailItem.vue'

const MIN_ZOOM = 0.5
const MAX_ZOOM = 10
const ZOOM_STEP = 0.15

const ACTION_COLOR = '#ef4444'
const SWIPE_COLORS = ['#3b82f6', '#22c55e', '#f59e0b', '#8b5cf6', '#ec4899', '#14b8a6']

const props = defineProps<{
    detail: ActionDetailResponse
    rawImage: string
    fullscreen?: boolean
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const containerRef = ref<HTMLDivElement | null>(null)
const zoomLevel = ref(1)
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const panOffset = ref({ x: 0, y: 0 })
const rawImageObj = ref<HTMLImageElement | null>(null)

const imgWidth = computed(() => rawImageObj.value?.naturalWidth ?? 0)
const imgHeight = computed(() => rawImageObj.value?.naturalHeight ?? 0)

const containerStyle = computed(() => {
    if (props.fullscreen) return { flex: '1 1 0', minHeight: '0' }
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

const zoomPercentage = computed(() => Math.round(zoomLevel.value * 100))

function loadRawImage() {
    if (!props.rawImage) {
        rawImageObj.value = null
        return
    }
    const img = new Image()
    img.onload = () => {
        rawImageObj.value = img
        nextTick(() => drawCanvas())
    }
    img.src = props.rawImage
}

// ---- Drawing ----

function drawPoint(ctx: CanvasRenderingContext2D, p: PointResponse, color: string, radius: number, label?: string) {
    ctx.beginPath()
    ctx.arc(p.x, p.y, radius, 0, Math.PI * 2)
    ctx.fillStyle = color + '80'
    ctx.fill()
    ctx.strokeStyle = color
    ctx.lineWidth = Math.max(2, radius / 3)
    ctx.stroke()

    // Crosshair
    const cr = radius * 1.5
    ctx.beginPath()
    ctx.moveTo(p.x - cr, p.y)
    ctx.lineTo(p.x + cr, p.y)
    ctx.moveTo(p.x, p.y - cr)
    ctx.lineTo(p.x, p.y + cr)
    ctx.strokeStyle = color
    ctx.lineWidth = Math.max(1, radius / 5)
    ctx.stroke()

    if (label) {
        const fontSize = Math.max(12, radius * 1.2)
        ctx.font = `bold ${fontSize}px system-ui, sans-serif`
        const metrics = ctx.measureText(label)
        const pad = fontSize * 0.3
        const lx = p.x + radius + 4
        const ly = p.y - fontSize / 2

        ctx.fillStyle = color + 'DD'
        ctx.fillRect(lx, ly - pad, metrics.width + pad * 2, fontSize + pad * 2)
        ctx.fillStyle = '#ffffff'
        ctx.textBaseline = 'top'
        ctx.fillText(label, lx + pad, ly)
    }
}

function drawArrow(ctx: CanvasRenderingContext2D, from: PointResponse, to: PointResponse, color: string, lineWidth: number) {
    const dx = to.x - from.x
    const dy = to.y - from.y
    const angle = Math.atan2(dy, dx)
    const headLen = Math.max(10, lineWidth * 4)

    ctx.beginPath()
    ctx.moveTo(from.x, from.y)
    ctx.lineTo(to.x, to.y)
    ctx.strokeStyle = color
    ctx.lineWidth = lineWidth
    ctx.stroke()

    ctx.beginPath()
    ctx.moveTo(to.x, to.y)
    ctx.lineTo(to.x - headLen * Math.cos(angle - Math.PI / 6), to.y - headLen * Math.sin(angle - Math.PI / 6))
    ctx.lineTo(to.x - headLen * Math.cos(angle + Math.PI / 6), to.y - headLen * Math.sin(angle + Math.PI / 6))
    ctx.closePath()
    ctx.fillStyle = color
    ctx.fill()
}

function drawSwipePath(ctx: CanvasRenderingContext2D, begin: PointResponse, end: PointResponse[], color: string, baseSize: number) {
    const lineWidth = Math.max(2, baseSize / 150)
    const points = [begin, ...end]

    // Draw path segments with arrows
    for (let i = 0; i < points.length - 1; i++) {
        drawArrow(ctx, points[i]!, points[i + 1]!, color, lineWidth)
    }

    // Draw begin point
    drawPoint(ctx, begin, color, baseSize / 60, 'Start')

    // Draw end point
    if (end.length > 0) {
        drawPoint(ctx, end[end.length - 1]!, color, baseSize / 80, 'End')
    }
}

function drawScrollIndicator(ctx: CanvasRenderingContext2D, p: PointResponse, dx: number, dy: number, color: string, baseSize: number) {
    const radius = baseSize / 40
    drawPoint(ctx, p, color, radius)

    // Draw scroll direction arrow
    const scale = baseSize / 8
    const arrowDx = Math.sign(dx) * Math.min(Math.abs(dx), 1) * scale
    const arrowDy = Math.sign(dy) * Math.min(Math.abs(dy), 1) * scale
    const target = { x: p.x + arrowDx, y: p.y + arrowDy }

    if (arrowDx !== 0 || arrowDy !== 0) {
        drawArrow(ctx, p, target, color, Math.max(2, baseSize / 200))

        const fontSize = Math.max(12, baseSize / 50)
        ctx.font = `bold ${fontSize}px system-ui, sans-serif`
        const label = `Scroll (${dx}, ${dy})`
        const metrics = ctx.measureText(label)
        const pad = fontSize * 0.3
        const lx = p.x + radius + 4
        const ly = p.y + radius + 4

        ctx.fillStyle = color + 'DD'
        ctx.fillRect(lx, ly, metrics.width + pad * 2, fontSize + pad * 2)
        ctx.fillStyle = '#ffffff'
        ctx.textBaseline = 'top'
        ctx.fillText(label, lx + pad, ly + pad * 0.5)
    }
}

function drawBox(ctx: CanvasRenderingContext2D, box: { x: number; y: number; w: number; h: number }, color: string, lineWidth: number) {
    ctx.strokeStyle = color
    ctx.lineWidth = lineWidth
    ctx.setLineDash([6, 4])
    ctx.strokeRect(box.x, box.y, box.w, box.h)
    ctx.setLineDash([])
    ctx.fillStyle = color + '15'
    ctx.fillRect(box.x, box.y, box.w, box.h)
}

function drawCanvas() {
    const canvas = canvasRef.value
    const img = rawImageObj.value
    if (!canvas || !img) return

    canvas.width = img.naturalWidth
    canvas.height = img.naturalHeight

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.drawImage(img, 0, 0)

    const baseSize = Math.min(img.naturalWidth, img.naturalHeight)
    const result = props.detail.result

    // Draw box (from recognition)
    if (props.detail.box) {
        drawBox(ctx, props.detail.box, '#f59e0b', Math.max(2, baseSize / 300))
    }

    if (!result || !actionHasCoords(result)) return

    const pointRadius = baseSize / 50

    switch (result.type) {
        case 'Click':
            drawPoint(ctx, (result as ClickActionResult).point, ACTION_COLOR, pointRadius, 'Click')
            break

        case 'LongPress':
            drawPoint(ctx, (result as LongPressActionResult).point, ACTION_COLOR, pointRadius, `LongPress ${(result as LongPressActionResult).duration}ms`)
            break

        case 'Swipe':
            drawSwipePath(ctx, (result as SwipeActionResult).begin, (result as SwipeActionResult).end, ACTION_COLOR, baseSize)
            break

        case 'MultiSwipe': {
            const swipes = (result as MultiSwipeActionResult).swipes
            swipes.forEach((swipe, idx) => {
                const color = SWIPE_COLORS[idx % SWIPE_COLORS.length]!
                drawSwipePath(ctx, swipe.begin, swipe.end, color, baseSize)
            })
            break
        }

        case 'TouchDown':
        case 'TouchMove':
        case 'TouchUp':
            drawPoint(ctx, (result as TouchActionResult).point, ACTION_COLOR, pointRadius, result.type)
            break

        case 'Scroll':
            drawScrollIndicator(ctx, (result as ScrollActionResult).point, (result as ScrollActionResult).dx, (result as ScrollActionResult).dy, ACTION_COLOR, baseSize)
            break
    }
}

// ---- Zoom & Pan ----

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
    if (!props.rawImage) return
    if (e.deltaY < 0) zoomIn()
    else zoomOut()
}

const dragThreshold = 4
const mouseDownPos = ref<{ x: number; y: number } | null>(null)
let didDrag = false

function onMouseDown(e: MouseEvent) {
    mouseDownPos.value = { x: e.clientX, y: e.clientY }
    didDrag = false
    isDragging.value = true
    dragStart.value = { x: e.clientX, y: e.clientY }
    dragOffset.value = { ...panOffset.value }
}

function onMouseMove(e: MouseEvent) {
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
    }
}

function onMouseUp() {
    isDragging.value = false
    mouseDownPos.value = null
}

function onMouseLeave() {
    isDragging.value = false
    mouseDownPos.value = null
}

// ---- Download ----

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
    a.download = `action_draw_${timestamp}.png`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
}

// ---- Watchers ----

watch(() => props.rawImage, () => {
    loadRawImage()
    resetView()
})

watch(() => props.detail, () => {
    nextTick(() => drawCanvas())
}, { deep: true })

watch(zoomLevel, () => {
    nextTick(() => drawCanvas())
})

onMounted(() => {
    loadRawImage()
})
</script>

<style scoped>
.action-draw-root {
    display: flex;
    flex-direction: row;
    gap: 12px;
    min-height: 60vh;
}

.action-draw-left {
    width: 280px;
    min-width: 220px;
    max-width: 340px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
    overflow-y: auto;
}

.action-draw-right {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.action-draw-canvas-container {
    position: relative;
    overflow: hidden;
    border-radius: 8px;
    border: 1px solid var(--ui-border-default, #333);
    background: var(--ui-bg-muted);
    flex: 1;
    min-height: 200px;
}
</style>
