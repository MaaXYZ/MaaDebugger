<template>
    <UModal v-model:open="open" title="Recognition Detail" :ui="{ width: 'sm:max-w-2xl' }">
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
                <div v-if="detail.combined_result && detail.combined_result.length > 0" class="flex flex-col gap-2">
                    <span class="text-sm font-medium text-dimmed">Combined ({{ detail.algorithm }}):</span>
                    <div class="pl-3 border-l-2 border-default flex flex-col gap-2">
                        <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub"
                            :depth="1" />
                    </div>
                </div>

                <!-- Detail JSON -->
                <div v-if="detail.detail_json" class="flex flex-col gap-1">
                    <span class="text-xs text-dimmed">Detail JSON:</span>
                    <pre
                        class="text-xs bg-elevated rounded p-2 overflow-auto max-h-40">{{ JSON.stringify(detail.detail_json, null, 2) }}</pre>
                </div>

                <!-- Draw images -->
                <div v-if="detail.draw_images && detail.draw_images.length > 0" class="flex flex-col gap-2">
                    <span class="text-xs text-dimmed">Draw:</span>
                    <div class="flex flex-row flex-wrap gap-2">
                        <img v-for="(img, idx) in detail.draw_images" :key="idx" :src="img"
                            class="max-w-full rounded border border-default" />
                    </div>
                </div>
            </div>
            <div v-else class="text-sm text-dimmed p-4 text-center">
                No detail available
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { getNodeDetail } from '@/api/http'
import type { RecoDetailResponse } from './types'
import RecoDetailItem from './RecoDetailItem.vue'

const props = defineProps<{
    nodeName: string | null
}>()

const open = defineModel<boolean>('open', { default: false })
const loading = ref(false)
const detail = ref<RecoDetailResponse | null>(null)

watch([() => props.nodeName, open], async ([name, isOpen]) => {
    if (!isOpen || !name) {
        detail.value = null
        return
    }
    loading.value = true
    try {
        const nodeDetail = await getNodeDetail(name)
        detail.value = nodeDetail?.recognition ?? null
    } catch {
        detail.value = null
    } finally {
        loading.value = false
    }
})
</script>
