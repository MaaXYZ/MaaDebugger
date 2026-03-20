<template>
    <div v-if="nodes.length > 0" class="flex flex-col gap-2">
        <div class="flex items-center gap-2 text-[11px] text-dimmed uppercase tracking-wide">
            <UIcon name="i-lucide-git-branch-plus" class="size-3.5" />
            <span>Subflow</span>
            <UBadge size="xs" variant="subtle" :color="summaryColor">
                {{ nodes.length }}
            </UBadge>
        </div>

        <div class="rounded-lg border border-default bg-muted/30 px-3 py-2">
            <div class="flex flex-col gap-2 border-l border-default pl-3">
                <SubflowNode
                    v-for="(node, index) in nodes"
                    :key="nodeStableKey(node, index)"
                    :node="node"
                    @request-detail="$emit('requestDetail', $event)"
                    @request-action-detail="$emit('requestActionDetail', $event)"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AnyNodeScope } from './types'
import SubflowNode from './SubflowNode.vue'
import { nodeStableKey, summarizeAnyNodesStatus } from './scopeTree'

const props = defineProps<{
    nodes: AnyNodeScope[]
}>()

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

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
