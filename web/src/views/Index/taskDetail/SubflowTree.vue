<template>
    <div v-if="nodes.length > 0" class="flex flex-col gap-2">
        <div class="flex items-center gap-2 text-xs text-dimmed flex-wrap">
            <UIcon name="i-lucide-workflow" class="size-3.5" />
            <span class="font-medium text-default">Internal custom flow</span>
            <UBadge size="xs" color="info" variant="subtle">Custom</UBadge>
            <UBadge v-if="contextLabel" size="xs" color="neutral" variant="subtle">{{ contextLabel }}</UBadge>
            <UBadge size="xs" variant="subtle" :color="summaryColor">
                {{ nodes.length }} node{{ nodes.length > 1 ? 's' : '' }}
            </UBadge>
        </div>

        <div class="rounded-lg border border-default bg-muted/30 px-3 py-2">
            <div class="flex flex-col gap-2 border-l border-default pl-3">
                <SubflowNode v-for="(node, index) in nodes" :key="nodeStableKey(node, index)" :node="node"
                    @request-detail="$emit('requestDetail', $event)"
                    @request-action-detail="$emit('requestActionDetail', $event)" />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AnyNodeScope } from '@/types/taskDetail'
import SubflowNode from './SubflowNode.vue'
import { nodeStableKey, summarizeAnyNodesStatus } from './scopeTree'

const props = withDefaults(defineProps<{
    nodes: AnyNodeScope[]
    kind?: 'reco' | 'action'
}>(), {
    kind: undefined,
})

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

const contextLabel = computed(() => {
    switch (props.kind) {
        case 'reco':
            return 'Inside Reco'
        case 'action':
            return 'Inside Action'
        default:
            return ''
    }
})

const summaryStatus = computed(() => summarizeAnyNodesStatus(props.nodes))
const summaryColor = computed(() => {
    switch (summaryStatus.value) {
        case 'running':
            return 'info' as const
        case 'failed':
            return 'error' as const
        default:
            return 'success' as const
    }
})
</script>
