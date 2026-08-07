<template>
    <div class="w-full flex-1 rounded-lg border transition-colors"
        :class="highlighted ? 'border-primary bg-primary/5 ring-1 ring-inset ring-primary/30' : 'border-default hover:bg-elevated'">
        <!-- Header: always visible -->
        <div class="flex flex-row items-center gap-2 min-w-0 p-3 cursor-pointer select-none w-full"
            @click="expanded = !expanded">
            <UIcon :name="expanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
                class="size-3.5 shrink-0 text-dimmed" />
            <UIcon name="i-lucide-workflow" class="size-4 shrink-0 text-dimmed" />
            <div class="min-w-0 flex-1">
                <div class="min-w-0 flex items-start gap-2">
                    <span class="block min-w-0 flex-1 break-all text-sm font-medium" :title="node.msg.name">{{
                        node.msg.name }}</span>
                    <UBadge v-if="isEntry" :label="t('taskDetail.entry')" color="primary" variant="soft" size="xs" class="shrink-0" />
                </div>
            </div>
            <StatusIcon :status="node.status" class="shrink-0" />
            <span v-if="!expanded && node.reco.length > 0" class="text-xs text-dimmed tabular-nums ml-auto shrink-0">
                {{ formatRounds(node.reco.length) }}
            </span>
        </div>

        <UCollapsible v-model:open="expanded" :unmount-on-hide="true">
            <template #content>
                <div class="flex flex-col gap-2 px-3 pb-3">
                    <!-- Recognition section -->
                    <div v-if="node.reco.length > 0" class="flex flex-col gap-1.5">
                        <div class="flex flex-row items-center gap-1.5">
                            <UIcon name="i-lucide-scan-search" class="size-3.5 shrink-0 text-dimmed" />
                            <span class="text-xs text-dimmed">{{ t('taskDetail.reco') }}</span>
                            <span class="text-xs text-dimmed tabular-nums">({{ node.reco.length }})</span>
                        </div>
                        <div class="pl-5 flex flex-wrap items-start gap-1.5">
                            <template v-for="(nextList, idx) in node.reco" :key="idx">
                                <NextListItem :next-list="nextList" @request-detail="$emit('requestDetail', $event)"
                                    @request-action-detail="$emit('requestActionDetail', $event)" />
                            </template>
                        </div>
                    </div>

                    <!-- Action section -->
                    <div v-if="node.action" class="flex flex-col gap-1.5">
                        <div class="flex flex-row items-center gap-1.5">
                            <UIcon name="i-lucide-play" class="size-3.5 shrink-0 text-dimmed" />
                            <span class="text-xs text-dimmed">{{ t('taskDetail.action') }}</span>
                        </div>
                        <div class="pl-5 flex flex-col gap-2">
                            <NodeStatusButton :status="node.action.status" :label="t('taskDetail.action')" :meta="[t('taskDetail.custom')]"
                                :action-id="node.action.msg.action_id" size="sm"
                                @click="$emit('requestActionDetail', node.action!.msg.action_id)" />
                            <SubflowDisclosure v-if="node.action.childs.length > 0" :label="t('taskDetail.internalFlow')" kind="action"
                                :count="countActionSubflowNodes(node.action)">
                                <SubflowTree :nodes="node.action.childs" kind="action"
                                    @request-detail="$emit('requestDetail', $event)"
                                    @request-action-detail="$emit('requestActionDetail', $event)" />
                            </SubflowDisclosure>
                        </div>
                    </div>
                </div>
            </template>
        </UCollapsible>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PipelineNodeScope } from '@/types/taskDetail'
import StatusIcon from './StatusIcon.vue'
import NextListItem from './NextListItem.vue'
import NodeStatusButton from './NodeStatusButton.vue'
import SubflowTree from './SubflowTree.vue'
import SubflowDisclosure from './SubflowDisclosure.vue'
import { countActionSubflowNodes } from './scopeTree'

const props = defineProps<{
    node: PipelineNodeScope
    isEntry?: boolean
    defaultExpanded?: boolean
    highlighted?: boolean
}>()

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

const { t } = useI18n()

function formatRounds(count: number): string {
    return count > 1
        ? t('taskDetail.roundss', { count })
        : t('taskDetail.rounds', { count })
}

const expanded = ref(props.defaultExpanded ?? true)
</script>
