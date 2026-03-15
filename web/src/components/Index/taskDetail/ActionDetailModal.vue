<template>
    <UModal v-model:open="open" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
        <template #header>
            <div v-if="detail" class="flex flex-row items-center gap-2 flex-wrap">
                <span class="text-sm text-highlighted font-semibold">{{ detail.name }}</span>
                <UBadge :color="detail.success ? 'success' : 'error'" variant="subtle"
                    :label="detail.success ? 'Success' : 'Failed'" />
                <UBadge color="info" variant="subtle" :label="detail.action" />
                <UButton color="neutral" variant="ghost" size="xs" icon="i-lucide-file-json"
                    label="NodeData" @click="nodeDataOpen = true" />
            </div>
        </template>

        <template #body>
            <div v-if="loading" class="flex items-center justify-center p-8">
                <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
            </div>
            <div v-else-if="detail" class="flex flex-col gap-3">
                <ActionDrawCanvas v-if="hasCoords && rawImage" :detail="detail" :raw-image="rawImage" />
                <ActionDetailItem v-else :detail="detail" />
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                No action detail available
            </div>
        </template>
    </UModal>
    <NodeDataModal v-model:open="nodeDataOpen" :node-name="detail?.name ?? null" :action-id="props.actionId" />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getActionDetailById, getTaskImageUrl } from '@/api/http'
import type { ActionDetailResponse } from './types'
import { actionHasCoords } from './types'
import ActionDetailItem from './ActionDetailItem.vue'
import ActionDrawCanvas from './ActionDrawCanvas.vue'
import NodeDataModal from './NodeDataModal.vue'

const props = defineProps<{
    actionId: number | null
}>()

const open = defineModel<boolean>('open', { default: false })
const loading = ref(false)
const detail = ref<ActionDetailResponse | null>(null)
const rawImage = ref<string | null>(null)
const nodeDataOpen = ref(false)

const hasCoords = computed(() => actionHasCoords(detail.value?.result))

watch(open, (isOpen) => {
    if (!isOpen) {
        nodeDataOpen.value = false
    }
})

watch([() => props.actionId, open], async ([id, isOpen]) => {
    if (!isOpen || id == null) {
        detail.value = null
        rawImage.value = null
        return
    }
    loading.value = true
    try {
        const d = await getActionDetailById(id)
        detail.value = d
        rawImage.value = d?.raw_image ? (d.raw_image.url || getTaskImageUrl(d.raw_image.id)) : null
    } catch {
        detail.value = null
        rawImage.value = null
    } finally {
        loading.value = false
    }
})
</script>
