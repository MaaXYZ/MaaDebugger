<template>
    <UModal v-model:open="open" title="Action Detail" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
        <template #body>
            <div v-if="loading" class="flex items-center justify-center p-8">
                <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
            </div>
            <div v-else-if="actionDetail">
                <ActionDrawCanvas v-if="hasCoords && rawImage" :detail="actionDetail"
                                  :raw-image="rawImage" />
                <ActionDetailItem v-else :detail="actionDetail" />
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                No action detail available
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getNodeDetail } from '@/api/http'
import type { ActionDetailResponse } from './types'
import { actionHasCoords } from './types'
import ActionDetailItem from './ActionDetailItem.vue'
import ActionDrawCanvas from './ActionDrawCanvas.vue'

const props = defineProps<{
    nodeName: string | null
}>()

const open = defineModel<boolean>('open', { default: false })
const loading = ref(false)
const actionDetail = ref<ActionDetailResponse | null>(null)
const rawImage = ref<string | null>(null)

const hasCoords = computed(() => actionHasCoords(actionDetail.value?.result))

watch([() => props.nodeName, open], async ([name, isOpen]) => {
    if (!isOpen || !name) {
        actionDetail.value = null
        rawImage.value = null
        return
    }
    loading.value = true
    try {
        const nodeDetail = await getNodeDetail(name)
        actionDetail.value = nodeDetail?.action ?? null
        rawImage.value = nodeDetail?.recognition?.raw_image ?? null
    } catch {
        actionDetail.value = null
        rawImage.value = null
    } finally {
        loading.value = false
    }
})
</script>
