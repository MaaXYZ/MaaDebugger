<template>
    <div class="flex flex-col gap-2 py-1">
        <div class="flex items-start gap-2 min-w-0">
            <div class="mt-1 flex size-5 shrink-0 items-center justify-center rounded-md border border-default bg-default/40">
                <UIcon :name="iconName" class="size-3 text-dimmed" />
            </div>

            <div class="min-w-0 flex-1 rounded-md border border-default bg-default/60 px-2.5 py-2">
                <div class="flex items-start gap-2 min-w-0">
                    <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2 min-w-0 flex-wrap">
                            <span class="text-[11px] uppercase tracking-wide text-dimmed shrink-0">{{ kindLabel }}</span>
                            <span class="truncate text-sm font-medium text-highlighted">{{ title }}</span>
                        </div>
                    </div>
                    <StatusIcon :status="node.status" class="shrink-0" />
                </div>

                <div v-if="node.type === 'pipeline_node'" class="mt-2 flex flex-col gap-2">
                    <div v-if="node.reco.length > 0" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5 text-xs text-dimmed">
                            <UIcon name="i-lucide-scan-search" class="size-3.5" />
                            <span>Reco</span>
                        </div>
                        <div class="pl-4 flex flex-wrap items-start gap-1.5">
                            <NextListItem
                                v-for="(nextList, index) in node.reco"
                                :key="nextListStableKey(nextList, index)"
                                :next-list="nextList"
                                @request-detail="$emit('requestDetail', $event)"
                                @request-action-detail="$emit('requestActionDetail', $event)"
                            />
                        </div>
                    </div>

                    <div v-if="node.action" class="flex flex-col gap-1.5">
                        <div class="flex items-center gap-1.5 text-xs text-dimmed">
                            <UIcon name="i-lucide-play" class="size-3.5" />
                            <span>Action</span>
                        </div>
                        <div class="pl-4 flex flex-col gap-2">
                            <NodeStatusButton
                                :status="node.action.status"
                                :label="node.action.msg.name"
                                :action-id="node.action.msg.action_id"
                                size="sm"
                                @click="$emit('requestActionDetail', node.action.msg.action_id)"
                            />
                            <SubflowDisclosure
                                v-if="node.action.childs.length > 0"
                                label="Internal flow"
                                kind="action"
                                :count="countActionSubflowNodes(node.action)"
                            >
                                <SubflowTree
                                    :nodes="node.action.childs"
                                    kind="action"
                                    @request-detail="$emit('requestDetail', $event)"
                                    @request-action-detail="$emit('requestActionDetail', $event)"
                                />
                            </SubflowDisclosure>
                        </div>
                    </div>
                </div>

                <div v-else-if="node.type === 'reco_node'" class="mt-2 pl-4 flex flex-col gap-2">
                    <RecoButton
                        v-if="node.reco"
                        :reco="node.reco"
                        @request-detail="$emit('requestDetail', $event)"
                    />
                    <SubflowDisclosure
                        v-if="node.reco && node.reco.childs.length > 0"
                        label="Internal flow"
                        kind="reco"
                        :count="countRecoSubflowNodes(node.reco)"
                    >
                        <SubflowTree
                            :nodes="node.reco.childs"
                            kind="reco"
                            @request-detail="$emit('requestDetail', $event)"
                            @request-action-detail="$emit('requestActionDetail', $event)"
                        />
                    </SubflowDisclosure>
                </div>

                <div v-else-if="node.type === 'act_node'" class="mt-2 pl-4 flex flex-col gap-2">
                    <NodeStatusButton
                        v-if="node.action"
                        :status="node.action.status"
                        :label="node.action.msg.name"
                        :action-id="node.action.msg.action_id"
                        size="sm"
                        @click="$emit('requestActionDetail', node.action.msg.action_id)"
                    />
                    <SubflowDisclosure
                        v-if="node.action && node.action.childs.length > 0"
                        label="Internal flow"
                        kind="action"
                        :count="countActionSubflowNodes(node.action)"
                    >
                        <SubflowTree
                            :nodes="node.action.childs"
                            kind="action"
                            @request-detail="$emit('requestDetail', $event)"
                            @request-action-detail="$emit('requestActionDetail', $event)"
                        />
                    </SubflowDisclosure>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AnyNodeScope } from './types'
import StatusIcon from './StatusIcon.vue'
import NodeStatusButton from './NodeStatusButton.vue'
import NextListItem from './NextListItem.vue'
import RecoButton from './RecoButton.vue'
import SubflowTree from './SubflowTree.vue'
import SubflowDisclosure from './SubflowDisclosure.vue'
import {
    countActionSubflowNodes,
    countRecoSubflowNodes,
    nextListStableKey,
} from './scopeTree'

const props = defineProps<{
    node: AnyNodeScope
}>()

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

const kindLabel = computed(() => {
    switch (props.node.type) {
    case 'pipeline_node':
        return 'Pipeline'
    case 'reco_node':
        return 'Reco Node'
    case 'act_node':
        return 'Action Node'
    }
})

const title = computed(() => props.node.msg.name)

const iconName = computed(() => {
    switch (props.node.type) {
    case 'pipeline_node':
        return 'i-lucide-workflow'
    case 'reco_node':
        return 'i-lucide-scan-search'
    case 'act_node':
        return 'i-lucide-play'
    }
})
</script>
