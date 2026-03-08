<template>
    <UModal v-model:open="open" title="Action Detail" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
        <template #body>
            <div v-if="loading" class="flex items-center justify-center p-8">
                <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
            </div>
            <div v-else-if="actionDetail" class="flex flex-col gap-3">
                <div class="flex items-center justify-end">
                    <UButton color="neutral" variant="ghost" size="xs" icon="i-lucide-file-json"
                        @click="nodeDataOpen = true">
                        NodeData
                    </UButton>
                </div>
                <ActionDrawCanvas v-if="hasCoords && rawImage" :detail="actionDetail" :raw-image="rawImage" />
                <ActionDetailItem v-else :detail="actionDetail" />
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                No action detail available
            </div>
        </template>
    </UModal>
    <NodeDataModal v-model:open="nodeDataOpen" :node-name="actionDetail?.name ?? null" />
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
const actionDetail = ref<ActionDetailResponse | null>(null)
const rawImage = ref<string | null>(null)
const nodeDataOpen = ref(false)

const hasCoords = computed(() => actionHasCoords(actionDetail.value?.result))

watch(open, (isOpen) => {
    if (!isOpen) {
        nodeDataOpen.value = false
    }
})

watch([() => props.actionId, open], async ([id, isOpen]) => {
    if (!isOpen || id == null) {
        actionDetail.value = null
        rawImage.value = null
        return
    }
    loading.value = true
    try {
        const detail = await getActionDetailById(id)
        actionDetail.value = detail
        rawImage.value = detail?.raw_image ? (detail.raw_image.url || getTaskImageUrl(detail.raw_image.id)) : null
    } catch {
        actionDetail.value = null
        rawImage.value = null
    } finally {
        loading.value = false
    }
})
</script>
