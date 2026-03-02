<template>
    <div class="rounded-lg border border-default transition-colors hover:bg-elevated w-full flex-1">
        <!-- Header: always visible -->
        <div class="flex flex-row items-center gap-2 min-w-0 p-3 cursor-pointer select-none w-full"
            @click="expanded = !expanded">
            <UIcon :name="expanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                class="size-3.5 shrink-0 text-dimmed" />
            <UIcon name="i-lucide-workflow" class="size-4 shrink-0 text-dimmed" />
            <div class="min-w-0 flex-1">
                <span class="text-sm font-medium truncate block w-full" :title="node.msg.name">{{ node.msg.name
                    }}</span>
            </div>
            <StatusIcon :status="node.status" class="shrink-0" />
            <span v-if="!expanded && node.reco.length > 0" class="text-xs text-dimmed tabular-nums ml-auto shrink-0">
                {{ node.reco.length }} round{{ node.reco.length > 1 ? 's' : '' }}
            </span>
        </div>

        <UCollapsible v-model:open="expanded" :unmount-on-hide="true">
            <template #content>
                <div class="flex flex-col gap-2 px-3 pb-3">
                    <!-- Recognition section -->
                    <div v-if="node.reco.length > 0" class="flex flex-col gap-1.5">
                        <div class="flex flex-row items-center gap-1.5">
                            <UIcon name="i-lucide-scan-search" class="size-3.5 shrink-0 text-dimmed" />
                            <span class="text-xs text-dimmed">Reco</span>
                            <span class="text-xs text-dimmed tabular-nums">({{ node.reco.length }})</span>
                        </div>
                        <div class="pl-5 flex flex-col gap-1.5">
                            <UButton v-if="hasHiddenRounds" size="xs" variant="ghost" color="neutral" class="self-start"
                                @click.stop="showAllRounds = !showAllRounds">
                                {{ showAllRounds ? 'Collapse' : `Show ${hiddenCount} older round${hiddenCount > 1 ? 's'
                                    : ''}...` }}
                            </UButton>
                            <template v-for="(nextList, idx) in visibleReco" :key="recoOffset + idx">
                                <NextListItem :next-list="nextList" @request-detail="$emit('requestDetail', $event)" />
                            </template>
                        </div>
                    </div>

                    <!-- Action section -->
                    <div v-if="node.action" class="flex flex-col gap-1.5">
                        <div class="flex flex-row items-center gap-1.5">
                            <UIcon name="i-lucide-play" class="size-3.5 shrink-0 text-dimmed" />
                            <span class="text-xs text-dimmed">Action</span>
                        </div>
                        <div class="pl-5">
                            <NodeStatusButton :status="node.action.status" :label="node.action.msg.name" size="sm"
                                @click="$emit('requestActionDetail', node.action!.msg.action_id)" />
                        </div>
                    </div>
                </div>
            </template>
        </UCollapsible>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { PipelineNodeScope } from './types'
import StatusIcon from './StatusIcon.vue'
import NextListItem from './NextListItem.vue'
import NodeStatusButton from './NodeStatusButton.vue'

const VISIBLE_ROUNDS = 5

const props = defineProps<{
    node: PipelineNodeScope
    defaultExpanded?: boolean
}>()

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

const expanded = ref(props.defaultExpanded ?? true)
const showAllRounds = ref(false)

const hiddenCount = computed(() => Math.max(0, props.node.reco.length - VISIBLE_ROUNDS))
const hasHiddenRounds = computed(() => hiddenCount.value > 0)

const recoOffset = computed(() =>
    showAllRounds.value || !hasHiddenRounds.value ? 0 : hiddenCount.value
)

const visibleReco = computed(() => {
    if (showAllRounds.value || !hasHiddenRounds.value) {
        return props.node.reco
    }
    return props.node.reco.slice(-VISIBLE_ROUNDS)
})
</script>
