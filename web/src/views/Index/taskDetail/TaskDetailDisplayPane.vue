<template>
    <div class="flex min-h-0 flex-col gap-3">
        <div class="flex min-w-0 flex-wrap items-start gap-2 text-sm">
            <UBadge color="neutral" variant="outline" size="sm" class="shrink-0">
                #{{ activeIndex + 1 }}
            </UBadge>
            <UBadge
                :color="activeTask.status === 'success' ? 'success' : activeTask.status === 'failed' ? 'error' : 'info'"
                variant="subtle" class="shrink-0 capitalize">
                {{ activeTask.status }}
            </UBadge>
            <UTooltip :text="activeTask.msg.entry">
                <span class="block min-w-0 flex-1 break-all text-dimmed">{{ activeTask.msg.entry }}</span>
            </UTooltip>
            <UBadge color="neutral" variant="soft" size="sm" class="shrink-0">
                {{ t('taskDetail.nodeCount', { displayed: displayedNodes.length, total: totalNodeCount }) }}
            </UBadge>
            <UBadge v-if="isHistoryMode" color="warning" variant="soft" size="sm" class="shrink-0">
                {{ t('taskDetail.browsingHistory') }}
            </UBadge>
        </div>

        <div v-if="totalNodeCount > 0" class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2 text-xs text-dimmed">
                <span>{{ t('taskDetail.page', { current: currentPage, total: totalPages }) }}</span>
                <span>•</span>
                <span>{{ reverseNodeOrder ? t('taskDetail.newestFirst') : t('taskDetail.oldestFirst') }}</span>
                <span>•</span>
                <span>{{ isHistoryMode ? t('taskDetail.historyMode') : t('taskDetail.liveMode') }}</span>
            </div>
            <div class="flex flex-wrap items-center justify-end gap-2">
                <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevrons-left"
                    :disabled="currentPage <= 1" @click="$emit('goPage', 1)">
                    {{ t('taskDetail.first') }}
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevron-left"
                    :disabled="currentPage <= 1" @click="$emit('goPage', currentPage - 1)">
                    {{ t('taskDetail.prev') }}
                </UButton>
                <UButton v-if="isHistoryMode" size="xs" color="primary" variant="soft"
                    icon="i-lucide-arrow-down-to-line" @click="$emit('goLatest')">
                    {{ t('taskDetail.latest') }}
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevron-right"
                    :disabled="currentPage >= totalPages" @click="$emit('goPage', currentPage + 1)">
                    {{ t('taskDetail.next') }}
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevrons-right"
                    :disabled="currentPage >= totalPages" @click="$emit('goPage', totalPages)">
                    {{ t('taskDetail.last') }}
                </UButton>
            </div>
        </div>

        <div v-if="displayedNodes.length > 0" ref="feedContainer" class="min-h-0 flex-1 overflow-y-auto pr-1">
            <div class="flex flex-col gap-3">
                <div v-for="node in displayedNodes" :key="`${activeTask.msg.uuid}-${currentPage}-${node.msg.node_id}`"
                    :ref="(el) => setNodeElement(node.msg.node_id, el)" class="scroll-mt-3"
                    :data-node-id="node.msg.node_id">
                    <PipelineNodeItem :node="node" :is-entry="node.msg.node_id === entryNodeId"
                        :default-expanded="false" :highlighted="node.msg.node_id === selectedNodeId"
                        @request-detail="$emit('requestDetail', $event)"
                        @request-action-detail="$emit('requestActionDetail', $event)" />
                </div>
            </div>
        </div>
        <div v-else class="pl-2 text-xs italic text-dimmed">
            {{ t('taskDetail.noPipelineNodes') }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ComponentPublicInstance } from 'vue'
import type { PipelineNodeScope, TaskScope } from '@/types/taskDetail'
import PipelineNodeItem from './PipelineNodeItem.vue'

const props = defineProps<{
    activeTask: TaskScope
    activeIndex: number
    displayedNodes: PipelineNodeScope[]
    selectedNodeId: number | null
    currentPage: number
    totalPages: number
    totalNodeCount: number
    entryNodeId: number | null
    reverseNodeOrder: boolean
    isHistoryMode: boolean
    scrollRequestKey: number
}>()

const { t } = useI18n()

defineEmits<{
    goPage: [page: number]
    goLatest: []
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

const feedContainer = ref<HTMLElement | null>(null)
const nodeElements = new Map<number, HTMLElement>()

function setNodeElement(nodeId: number, el: Element | ComponentPublicInstance | null) {
    if (el instanceof HTMLElement) {
        nodeElements.set(nodeId, el)
        return
    }
    nodeElements.delete(nodeId)
}

async function scrollToNode(nodeId: number | null) {
    if (nodeId == null) return
    await nextTick()
    nodeElements.get(nodeId)?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

watch(() => props.scrollRequestKey, () => {
    void scrollToNode(props.selectedNodeId)
})

watch([() => props.activeTask.msg.uuid, () => props.currentPage], async () => {
    await nextTick()
    feedContainer.value?.scrollTo({ top: 0, behavior: 'auto' })
})
</script>
