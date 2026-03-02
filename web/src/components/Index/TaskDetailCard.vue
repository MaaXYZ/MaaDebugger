<template>
    <UCard class="w-full h-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task Detail</span>
                <div class="flex-1"></div>
                <template v-if="allTasks.length > 1">
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevron-left"
                        :disabled="activeIndex <= 0" @click="activeIndex--" />
                    <span class="text-xs tabular-nums text-dimmed min-w-12 text-center">
                        {{ activeIndex + 1 }} / {{ allTasks.length }}
                    </span>
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevron-right"
                        :disabled="activeIndex >= allTasks.length - 1" @click="activeIndex++" />
                </template>
                <UTooltip v-if="showOpenAsPage" text="Open as page">
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-arrow-up-right" to="/TaskDetail"
                        aria-label="Open Task Detail as page" />
                </UTooltip>
                <UButton v-if="allTasks.length > 0" size="xs" variant="ghost" color="neutral" icon="i-lucide-trash-2"
                    @click="resetGraph" />
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <!-- Empty state -->
                <UEmpty v-if="allTasks.length === 0" icon="i-material-symbols:checklist-rounded"
                    title="No Task Details" />

                <template v-else-if="activeTask">
                    <!-- Task status header -->
                    <div class="flex flex-row items-center gap-2 text-sm min-w-0">
                        <UBadge color="neutral" variant="outline" size="sm" class="shrink-0">
                            #{{ activeIndex + 1 }}
                        </UBadge>
                        <UBadge
                            :color="activeTask.status === 'success' ? 'success' : activeTask.status === 'failed' ? 'error' : 'info'"
                            variant="subtle" class="capitalize shrink-0">
                            {{ activeTask.status }}
                        </UBadge>
                        <UTooltip :text="activeTask.msg.entry">
                            <span class="text-dimmed min-w-0 flex-1 truncate block">{{ activeTask.msg.entry }}</span>
                        </UTooltip>
                    </div>

                    <!-- Pipeline nodes -->
                    <div v-if="activeTask.childs.length > 0" ref="scrollContainerRef"
                        class="node-list max-h-[60vh] overflow-y-auto pr-1">
                        <div class="flex flex-col gap-2">
                            <PipelineNodeItem v-for="(node, idx) in activeTask.childs"
                                :key="`${node.msg.name}-${node.msg.node_id}`" :node="node"
                                :default-expanded="idx === activeTask.childs.length - 1"
                                @request-detail="onRequestDetail" @request-action-detail="onRequestActionDetail" />
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
    <RecoDetailModal v-model:open="modalOpen" :reco-id="selectedRecoId" />

    <!-- Action Detail Modal -->
    <ActionDetailModal v-model:open="actionModalOpen" :action-id="selectedActionId" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { launchGraph, resetLaunchGraph } from '@/stores/launchGraph'
import { taskDetailActiveIndex, taskDetailFollowLatest } from '@/stores/taskDetail'
import PipelineNodeItem from './taskDetail/PipelineNodeItem.vue'
import RecoDetailModal from './taskDetail/RecoDetailModal.vue'
import ActionDetailModal from './taskDetail/ActionDetailModal.vue'
import { clearCache } from '@/api/http'

const route = useRoute()
const showOpenAsPage = computed(() => route.path !== '/TaskDetail')

const scrollContainerRef = ref<HTMLElement | null>(null)

const allTasks = computed(() => launchGraph.value.childs)
const activeIndex = taskDetailActiveIndex
const followLatest = taskDetailFollowLatest

const activeTask = computed(() => {
    if (allTasks.value.length === 0) return null
    const idx = Math.min(activeIndex.value, allTasks.value.length - 1)
    return allTasks.value[idx] ?? null
})

// When a new task arrives, auto-switch to it (unless user manually navigated away)
watch(() => allTasks.value.length, (newLen, oldLen) => {
    if (newLen > (oldLen ?? 0) && followLatest.value) {
        activeIndex.value = newLen - 1
    }
})

// Track whether user is on the latest task
watch(activeIndex, (idx) => {
    followLatest.value = idx === allTasks.value.length - 1
})

// Auto-scroll when new nodes appear in the active task
watch(
    () => activeTask.value?.childs.length,
    () => {
        nextTick(() => {
            if (scrollContainerRef.value) {
                scrollContainerRef.value.scrollTop = scrollContainerRef.value.scrollHeight
            }
        })
    },
)

// --- Reco Modal ---
const modalOpen = ref(false)
const selectedRecoId = ref<number | null>(null)

function onRequestDetail(recoId: number) {
    selectedRecoId.value = recoId
    modalOpen.value = true
}

// --- Action Modal ---
const actionModalOpen = ref(false)
const selectedActionId = ref<number | null>(null)

function onRequestActionDetail(actionId: number) {
    selectedActionId.value = actionId
    actionModalOpen.value = true
}

async function resetGraph() {
    resetLaunchGraph()
    activeIndex.value = 0
    followLatest.value = true
    await clearCache()
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
