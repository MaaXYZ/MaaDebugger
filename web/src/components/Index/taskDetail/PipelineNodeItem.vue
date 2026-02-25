<template>
    <div class="flex flex-col gap-2 rounded-lg border border-default p-3 transition-colors hover:bg-elevated">
        <!-- Header: Node Name + Status -->
        <div class="flex flex-row items-center gap-2 min-w-0">
            <UIcon name="i-lucide-workflow" class="size-4 shrink-0 text-dimmed" />
            <UTooltip :text="node.msg.name">
                <span class="text-sm font-medium truncate block max-w-full">{{ node.msg.name }}</span>
            </UTooltip>
            <StatusIcon :status="node.status" />
        </div>

        <!-- Recognition section: NextList → Reco buttons -->
        <div v-if="node.reco.length > 0" class="flex flex-col gap-1.5">
            <div class="flex flex-row items-center gap-1.5">
                <UIcon name="i-lucide-scan-search" class="size-3.5 shrink-0 text-dimmed" />
                <span class="text-xs text-dimmed">Reco</span>
            </div>
            <div class="pl-5 flex flex-col gap-1.5">
                <NextListItem v-for="(nextList, idx) in node.reco" :key="idx" :next-list="nextList"
                    @request-detail="$emit('requestDetail', $event)" />
            </div>
        </div>

        <!-- Action section: only show when reco succeeded -->
        <div v-if="node.action" class="flex flex-col gap-1.5">
            <div class="flex flex-row items-center gap-1.5">
                <UIcon name="i-lucide-play" class="size-3.5 shrink-0 text-dimmed" />
                <span class="text-xs text-dimmed">Action</span>
            </div>
            <div class="pl-5">
                <NodeStatusButton :status="node.action.status" :label="node.action.msg.name" size="md"
                    @click="$emit('requestDetail', node.msg.name)" />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { PipelineNodeScope } from './types'
import StatusIcon from './StatusIcon.vue'
import NextListItem from './NextListItem.vue'
import NodeStatusButton from './NodeStatusButton.vue'

defineProps<{
    node: PipelineNodeScope
}>()

defineEmits<{
    requestDetail: [name: string]
}>()
</script>
