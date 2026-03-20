<template>
    <div class="flex flex-col gap-3 min-h-0">
        <div class="flex flex-wrap items-center gap-2 text-sm min-w-0">
            <UBadge color="neutral" variant="outline" size="sm" class="shrink-0">
                #{{ activeIndex + 1 }}
            </UBadge>
            <UBadge
                :color="activeTask.status === 'success' ? 'success' : activeTask.status === 'failed' ? 'error' : 'info'"
                variant="subtle" class="capitalize shrink-0"
            >
                {{ activeTask.status }}
            </UBadge>
            <UTooltip :text="activeTask.msg.entry">
                <span class="text-dimmed min-w-0 flex-1 truncate block">{{ activeTask.msg.entry }}</span>
            </UTooltip>
            <UBadge color="neutral" variant="soft" size="sm" class="shrink-0">
                {{ displayedNodes.length }} / {{ activeTask.childs.length }} nodes
            </UBadge>
            <UBadge v-if="isHistoryMode" color="warning" variant="soft" size="sm" class="shrink-0">
                Browsing history
            </UBadge>
        </div>

        <div v-if="activeTask.childs.length > 0" class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2 text-xs text-dimmed">
                <span>Page {{ currentPage }} / {{ totalPages }}</span>
                <span>•</span>
                <span>{{ reverseNodeOrder ? 'Newest first' : 'Oldest first' }}</span>
                <span>•</span>
                <span>{{ isHistoryMode ? 'History mode' : 'Live mode' }}</span>
            </div>
            <div class="flex flex-wrap items-center justify-end gap-2">
                <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevrons-left"
                    :disabled="currentPage <= 1" @click="$emit('goPage', 1)">
                    First
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevron-left"
                    :disabled="currentPage <= 1" @click="$emit('goPage', currentPage - 1)">
                    Prev
                </UButton>
                <UButton v-if="isHistoryMode" size="xs" color="primary" variant="soft"
                    icon="i-lucide-arrow-down-to-line" @click="$emit('goLatest')">
                    Latest
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevron-right"
                    :disabled="currentPage >= totalPages" @click="$emit('goPage', currentPage + 1)">
                    Next
                </UButton>
                <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevrons-right"
                    :disabled="currentPage >= totalPages" @click="$emit('goPage', totalPages)">
                    Last
                </UButton>
            </div>
        </div>

        <div v-if="selectedNode" class="rounded-xl border border-default bg-default/40 p-3">
            <div class="mb-3 flex items-center gap-2">
                <UIcon name="i-lucide-inspection-panel" class="size-4 text-dimmed" />
                <span class="text-sm font-semibold">Selected node</span>
            </div>
            <PipelineNodeItem
                :key="`${selectedNode.msg.name}-${selectedNode.msg.node_id}`"
                :node="selectedNode"
                :is-entry="selectedNode.msg.node_id === entryNodeId"
                :default-expanded="true"
                @request-detail="$emit('requestDetail', $event)"
                @request-action-detail="$emit('requestActionDetail', $event)"
            />
        </div>
        <div v-else-if="displayedNodes.length > 0" class="text-xs text-dimmed italic pl-2">
            Select a node from the navigator to inspect its details.
        </div>
        <div v-else class="text-xs text-dimmed italic pl-2">
            No pipeline nodes
        </div>
    </div>
</template>

<script setup lang="ts">
import type { PipelineNodeScope, TaskScope } from './types'
import PipelineNodeItem from './PipelineNodeItem.vue'

defineProps<{
    activeTask: TaskScope
    activeIndex: number
    displayedNodes: PipelineNodeScope[]
    selectedNode: PipelineNodeScope | null
    entryNodeId: number | null
    currentPage: number
    totalPages: number
    reverseNodeOrder: boolean
    isHistoryMode: boolean
}>()

defineEmits<{
    goPage: [page: number]
    goLatest: []
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()
</script>
