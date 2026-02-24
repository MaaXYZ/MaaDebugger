<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task Detail</span>
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <!-- Empty state -->
                <div v-if="nodes.length === 0"
                    class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-6 text-dimmed gap-2">
                    <UIcon name="i-lucide-list-checks" class="size-5" />
                    <span class="text-sm">No task nodes</span>
                </div>

                <!-- Node list -->
                <div v-else class="node-list flex flex-col gap-3 max-h-[60vh] overflow-y-auto pr-1">
                    <div v-for="node in nodes" :key="node.nodeId"
                        class="flex flex-col gap-2 rounded-lg border border-default p-3 transition-colors hover:bg-elevated">

                        <!-- Row 1: Node Name -->
                        <div class="flex flex-row items-center gap-2 min-w-0">
                            <UIcon name="i-lucide-workflow" class="size-4 shrink-0 text-dimmed" />
                            <UTooltip :text="node.name">
                                <span class="text-sm font-medium truncate block max-w-full">{{ node.name }}</span>
                            </UTooltip>
                        </div>

                        <!-- Row 2: Recognition Status (可能有多个, 每个 reco 旁边显示 reco_id) -->
                        <div class="flex flex-col gap-1.5">
                            <div class="flex flex-row items-center gap-1.5">
                                <UIcon name="i-lucide-scan-search" class="size-3.5 shrink-0 text-dimmed" />
                                <span class="text-xs text-dimmed">Reco</span>
                            </div>
                            <div class="flex flex-row flex-wrap gap-2 pl-5">
                                <div v-if="node.recoList.length === 0" class="text-xs text-dimmed italic">
                                    Empty
                                </div>
                                <NodeStatusButton v-for="reco in node.recoList" :key="reco.recoId" :status="reco.status"
                                    :label="reco.name" size="md" />
                            </div>
                        </div>

                        <!-- Row 3: Action Status -->
                        <div class="flex flex-col gap-1.5">
                            <div class="flex flex-row items-center gap-1.5">
                                <UIcon name="i-lucide-play" class="size-3.5 shrink-0 text-dimmed" />
                                <span class="text-xs text-dimmed">Action</span>
                            </div>
                            <div class="flex flex-row flex-wrap gap-2 pl-5">
                                <NodeStatusButton :status="node.actionStatus" :label="node.name" size="md" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import NodeStatusButton from './taskDetail/NodeStatusButton.vue'
import type { NodeDetail } from './taskDetail/types'

// --- State ---
const nodes = ref<NodeDetail[]>([
    {
        nodeId: 1,
        name: 'TestOCR',
        recoList: [
            { recoId: 101, name: 'TestOCR', status: 'success' },
        ],
        actionStatus: 'failed',
    },
])

// --- Public API ---
function getNodes() {
    return nodes.value
}

function setNodes(newNodes: NodeDetail[]) {
    nodes.value = newNodes
}

function addNode(node: NodeDetail) {
    nodes.value.push(node)
}

function clearNodes() {
    nodes.value = []
}

defineExpose({
    nodes,
    getNodes,
    setNodes,
    addNode,
    clearNodes,
})
</script>

<style scoped>
.node-list::-webkit-scrollbar {
    width: 4px;
}

.node-list::-webkit-scrollbar-track {
    background: transparent;
}

.node-list::-webkit-scrollbar-thumb {
    background-color: rgba(128, 128, 128, 0.3);
    border-radius: 2px;
}

.node-list::-webkit-scrollbar-thumb:hover {
    background-color: rgba(128, 128, 128, 0.5);
}

.node-list {
    scrollbar-width: thin;
    scrollbar-color: rgba(128, 128, 128, 0.3) transparent;
}
</style>
