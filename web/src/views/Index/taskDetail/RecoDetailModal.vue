<template>
    <UModal v-model:open="open" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
        <template #header>
            <div v-if="detail" class="flex flex-col gap-2 min-w-0">
                <UBreadcrumb v-if="breadcrumbItems.length > 1" :items="breadcrumbItems">
                    <template #item-label="{ item, active, index }">
                        <button v-if="!active" type="button"
                            class="cursor-pointer text-left hover:text-highlighted transition-colors"
                            @click="openRecoFromBreadcrumb(index)">
                            {{ item.label }}
                        </button>
                        <span v-else class="font-semibold">{{ item.label }}</span>
                    </template>
                </UBreadcrumb>
                <div class="flex flex-row items-center gap-2 flex-wrap min-w-0">
                    <span class="text-sm text-highlighted font-semibold truncate">{{ detail.name }}</span>
                    <UBadge :color="detail.hit ? 'success' : 'error'" variant="subtle"
                        :label="detail.hit ? t('taskDetail.hit') : t('taskDetail.miss')" />
                    <UBadge color="info" variant="subtle" :label="detail.algorithm" />
                    <UButton color="neutral" variant="ghost" size="xs" icon="i-lucide-file-json"
                        :label="t('taskDetail.nodeData')" @click="() => { nodeDataOpen = true }" />
                </div>
            </div>
        </template>

        <template #body>
            <div v-if="loading" class="flex items-center justify-center p-8">
                <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
            </div>
            <div v-else-if="detail" class="flex flex-col gap-4">
                <!-- Box -->
                <div v-if="detail.box" class="text-xs text-dimmed">
                    {{ t('taskDetail.box', { x: detail.box.x, y: detail.box.y, w: detail.box.w, h: detail.box.h }) }}
                </div>

                <!-- Combined Result (And/Or nesting) -->
                <div v-if="detail.combined_result && detail.combined_result.length > 0" class="flex flex-col gap-2">
                    <span class="text-sm font-medium text-dimmed">{{ t('taskDetail.combined', { algorithm: detail.algorithm }) }}:</span>
                    <div class="pl-3 border-l-2 border-default flex flex-col gap-2">
                        <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub" :depth="1"
                            @request-detail="openSubRecoDetail" />
                    </div>
                </div>

                <div v-if="detail.raw_image && detail.results">
                    <RecognitionDrawCanvas :detail="detail" :rois="parsedRois"
                        :on-toggle-fullscreen="() => { isFullscreen = true }" />
                </div>
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                {{ t('taskDetail.noDetailAvailable') }}
            </div>
        </template>
    </UModal>

    <!-- Fullscreen Canvas Draw Modal -->
    <UModal v-model:open="isFullscreen" :title="t('taskDetail.recognitionDraw')" fullscreen>
        <template #body>
            <div v-if="detail?.raw_image && detail?.results"
                class="w-full h-full flex flex-col overflow-hidden bg-muted p-4">
                <RecognitionDrawCanvas :detail="detail" :rois="parsedRois"
                    :on-toggle-fullscreen="() => { isFullscreen = false }" fullscreen />
            </div>
        </template>
    </UModal>

    <!-- Fullscreen Image Preview (for original draw_images) -->
    <UModal v-model:open="imagePreviewOpen" :title="t('taskDetail.imagePreview')" fullscreen>
        <template #body>
            <div class="relative w-full h-full flex items-center justify-center overflow-hidden bg-muted"
                @wheel.prevent="onPreviewWheel">
                <div class="flex items-center justify-center cursor-grab select-none"
                    :class="{ 'cursor-grabbing': isPreviewDragging }" @mousedown="onPreviewDragStart"
                    @mousemove="onPreviewDragMove" @mouseup="onPreviewDragEnd" @mouseleave="onPreviewDragEnd">
                    <img v-if="previewImageSrc" :src="previewImageSrc" :alt="t('taskDetail.preview')" draggable="false"
                        class="pointer-events-none max-w-none" :style="previewImageStyle" />
                </div>
                <!-- Fullscreen toolbar -->
                <div
                    class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 bg-elevated/90 backdrop-blur-sm rounded-lg px-3 py-2 border border-default shadow-lg">
                    <UTooltip :text="t('common.zoomOut')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-out" size="sm"
                            :disabled="previewZoom <= MIN_ZOOM" @click="previewZoomOut" />
                    </UTooltip>
                    <span class="text-xs text-muted min-w-10 text-center tabular-nums">
                        {{ previewZoomPercentage }}%
                    </span>
                    <UTooltip :text="t('common.zoomIn')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-zoom-in" size="sm"
                            :disabled="previewZoom >= MAX_ZOOM" @click="previewZoomIn" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip :text="t('common.fitToView')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-maximize" size="sm"
                            @click="resetPreviewZoom" />
                    </UTooltip>
                    <USeparator orientation="vertical" class="h-5" />
                    <UTooltip :text="t('common.download')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-download" size="sm"
                            @click="downloadPreviewImage" />
                    </UTooltip>
                </div>
            </div>
        </template>
    </UModal>
    <NodeDataModal v-model:open="nodeDataOpen" :node-name="props.nodeName ?? detail?.name ?? null"
        :reco-id="selectedRecoId" :initial-node-json="cachedNodeJson" />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNodeData, getRecoDetailById } from '@/api/http'
import type { RecoDetailResponse, RectResponse } from '@/types/taskDetail'
import RecoDetailItem from './RecoDetailItem.vue'
import RecognitionDrawCanvas from './RecognitionDrawCanvas.vue'
import NodeDataModal from './NodeDataModal.vue'

// --- Constants ---
const MIN_ZOOM = 0.1
const MAX_ZOOM = 5
const ZOOM_STEP = 0.15

const props = defineProps<{
    recoId: number | null
    nodeName?: string | null
}>()

const { t } = useI18n()

const open = defineModel<boolean>('open', { default: false })
const loading = ref(false)
const detail = ref<RecoDetailResponse | null>(null)
const nodeDataOpen = ref(false)
const recognitionNodeJson = ref<unknown>(null)
const cachedNodeJson = ref<string | null>(null)
const selectedRecoId = ref<number | null>(null)
const detailPath = ref<Array<{ recoId: number, name: string }>>([])

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

function isFiniteNumber(value: unknown): value is number {
    return typeof value === 'number' && Number.isFinite(value)
}

function toRectResponse(value: unknown): RectResponse | null {
    if (!Array.isArray(value) || value.length < 4) return null

    const [x, y, w, h] = value
    if (!isFiniteNumber(x) || !isFiniteNumber(y) || !isFiniteNumber(w) || !isFiniteNumber(h)) {
        return null
    }

    return { x, y, w, h }
}

function mergeRectWithOffset(base: RectResponse | null, offset: RectResponse | null): RectResponse | null {
    if (!base && !offset) return null
    if (!base) return offset
    if (!offset) return base
    return {
        x: base.x + offset.x,
        y: base.y + offset.y,
        w: base.w + offset.w,
        h: base.h + offset.h,
    }
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return !!value && typeof value === 'object' && !Array.isArray(value)
}

function collectRecognitionRois(source: unknown, bucket: RectResponse[]) {
    if (!source) return

    if (Array.isArray(source)) {
        source.forEach(item => collectRecognitionRois(item, bucket))
        return
    }

    if (!isRecord(source)) return

    const directRoi = toRectResponse(source.roi)
    const roiOffset = toRectResponse(source.roi_offset)
    const mergedRoi = mergeRectWithOffset(directRoi, roiOffset)
    if (mergedRoi && mergedRoi.w > 0 && mergedRoi.h > 0) {
        bucket.push(mergedRoi)
    }

    if (Array.isArray(source.all_of)) {
        source.all_of.forEach(item => collectRecognitionRois(item, bucket))
    }
    if (Array.isArray(source.any_of)) {
        source.any_of.forEach(item => collectRecognitionRois(item, bucket))
    }
    if (isRecord(source.param)) {
        collectRecognitionRois(source.param, bucket)
    }
}

function dedupeRois(rois: RectResponse[]): RectResponse[] {
    const seen = new Set<string>()
    return rois.filter((roi) => {
        const key = `${roi.x},${roi.y},${roi.w},${roi.h}`
        if (seen.has(key)) return false
        seen.add(key)
        return true
    })
}

const parsedRois = computed<RectResponse[]>(() => {
    const rois: RectResponse[] = []
    collectRecognitionRois(recognitionNodeJson.value, rois)
    return dedupeRois(rois)
})

const breadcrumbItems = computed(() => detailPath.value.map((item, index) => ({
    label: item.name,
    active: index === detailPath.value.length - 1,
})))

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

function resetRecoDetailState() {
    detail.value = null
    recognitionNodeJson.value = null
    cachedNodeJson.value = null
    nodeDataOpen.value = false
    isFullscreen.value = false
    imagePreviewOpen.value = false
    previewImageSrc.value = ''
    resetPreviewZoom()
}

function openSubRecoDetail(payload: { recoId: number, name: string }) {
    if (!payload.recoId) return
    resetRecoDetailState()
    selectedRecoId.value = payload.recoId

    const existsIndex = detailPath.value.findIndex(item => item.recoId === payload.recoId)
    if (existsIndex >= 0) {
        detailPath.value = detailPath.value.slice(0, existsIndex + 1)
        return
    }
    detailPath.value = [...detailPath.value, payload]
}

function openRecoFromBreadcrumb(index: number) {
    const target = detailPath.value[index]
    if (!target) return
    resetRecoDetailState()
    detailPath.value = detailPath.value.slice(0, index + 1)
    selectedRecoId.value = target.recoId
}

function resetBreadcrumbToRoot() {
    selectedRecoId.value = props.recoId
    detailPath.value = props.recoId == null ? [] : [{ recoId: props.recoId, name: props.nodeName ?? t('common.loading') }]
}

watch(() => props.recoId, (id) => {
    selectedRecoId.value = id
    detailPath.value = id == null ? [] : [{ recoId: id, name: props.nodeName ?? t('common.loading') }]
}, { immediate: true })

watch(open, (isOpen) => {
    if (!isOpen) {
        nodeDataOpen.value = false
        resetBreadcrumbToRoot()
    }
})

watch([selectedRecoId, open], async ([id, isOpen]) => {
    if (!isOpen || id == null) {
        detail.value = null
        recognitionNodeJson.value = null
        cachedNodeJson.value = null
        return
    }
    loading.value = true
    try {
        const [recoDetail, nodeData] = await Promise.all([
            getRecoDetailById(id),
            props.nodeName ? getNodeData(props.nodeName, { recoId: id }) : Promise.resolve(null),
        ])
        detail.value = recoDetail

        if (recoDetail && detailPath.value.length > 0 && detailPath.value[detailPath.value.length - 1]?.recoId === id) {
            detailPath.value = detailPath.value.map((item, index, arr) => {
                if (index !== arr.length - 1) return item
                return {
                    ...item,
                    name: recoDetail.name,
                }
            })
        }

        cachedNodeJson.value = nodeData?.node_json ?? null

        if (nodeData?.node_json) {
            try {
                const parsedNodeData = JSON.parse(nodeData.node_json) as Record<string, unknown>
                recognitionNodeJson.value = parsedNodeData.recognition ?? null
            } catch {
                recognitionNodeJson.value = null
            }
        } else {
            recognitionNodeJson.value = null
        }
    } catch {
        detail.value = null
        recognitionNodeJson.value = null
        cachedNodeJson.value = null
    } finally {
        loading.value = false
    }
})
</script>
