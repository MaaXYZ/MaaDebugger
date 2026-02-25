<template>
    <UCard class="w-full h-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task Detail</span>
                <UButton v-if="currentTask" size="xs" variant="ghost" color="neutral" icon="i-lucide-trash-2"
                    @click="resetGraph" />
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <!-- Empty state -->
                <div v-if="!currentTask"
                    class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-6 text-dimmed gap-2">
                    <UIcon name="i-lucide-list-checks" class="size-5" />
                    <span class="text-sm">No task running</span>
                </div>

                <!-- Task active but no pipeline nodes yet -->
                <template v-else>
                    <!-- Task status header -->
                    <div class="flex flex-row items-center gap-2 text-sm">
                        <UBadge
                            :color="currentTask.status === 'success' ? 'success' : currentTask.status === 'failed' ? 'error' : 'info'"
                            variant="subtle" class="capitalize">
                            {{ currentTask.status }}
                        </UBadge>
                        <span class="text-dimmed">{{ currentTask.msg.entry }}</span>
                    </div>

                    <!-- Pipeline nodes -->
                    <div v-if="currentTask.childs.length > 0" ref="scrollContainerRef"
                        class="node-list max-h-[60vh] overflow-y-auto pr-1">
                        <div class="flex flex-col gap-2">
                            <PipelineNodeItem v-for="(node, idx) in currentTask.childs" :key="idx" :node="node"
                                @request-detail="onRequestDetail" />
                        </div>
                    </div>
                    <div v-else class="text-xs text-dimmed italic pl-2">
                        No pipeline nodes
                    </div>
                </template>
            </div>
        </template>
    </UCard>

    <!-- Reco Detail Modal -->
    <RecoDetailModal v-model:open="modalOpen" :node-name="selectedNodeName" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { launchGraph, resetLaunchGraph } from '@/stores/launchGraph'
import PipelineNodeItem from './taskDetail/PipelineNodeItem.vue'
import RecoDetailModal from './taskDetail/RecoDetailModal.vue'

const scrollContainerRef = ref<HTMLElement | null>(null)

const currentTask = computed(() => {
    const graph = launchGraph.value
    return graph.childs.length > 0 ? graph.childs[graph.childs.length - 1] : null
})

// Auto-scroll on new nodes
watch(
    () => currentTask.value?.childs.length,
    () => {
        nextTick(() => {
            if (scrollContainerRef.value) {
                scrollContainerRef.value.scrollTop = scrollContainerRef.value.scrollHeight
            }
        })
    },
)

// --- Modal ---
const modalOpen = ref(false)
const selectedNodeName = ref<string | null>(null)

function onRequestDetail(name: string) {
    selectedNodeName.value = name
    modalOpen.value = true
}

function resetGraph() {
    resetLaunchGraph()
}
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
