<template>
    <UCard class="w-full max-h-screen flex-1" size="xl">
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
                <UEmpty v-if="allTasks.length === 0" icon="i-material-symbols:checklist-rounded"
                        title="No Task Details" />

                <template v-else-if="activeTask">
                    <div class="flex flex-wrap items-center gap-2 text-sm min-w-0">
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
                                     :disabled="currentPage <= 1" @click="goToPage(1)">
                                First
                            </UButton>
                            <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-chevron-left"
                                     :disabled="currentPage <= 1" @click="goToPage(currentPage - 1)">
                                Prev
                            </UButton>
                            <UButton v-if="isHistoryMode" size="xs" color="primary" variant="soft"
                                     icon="i-lucide-arrow-down-to-line" @click="goToLatestPage">
                                Latest
                            </UButton>
                            <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevron-right"
                                     :disabled="currentPage >= totalPages" @click="goToPage(currentPage + 1)">
                                Next
                            </UButton>
                            <UButton size="xs" variant="ghost" color="neutral" trailing-icon="i-lucide-chevrons-right"
                                     :disabled="currentPage >= totalPages" @click="goToPage(totalPages)">
                                Last
                            </UButton>
                        </div>
                    </div>

                    <div v-if="displayedNodes.length > 0" ref="scrollContainerRef"
                         class="node-list max-h-screen overflow-y-auto pr-1">
                        <div class="flex flex-col gap-2">
                            <PipelineNodeItem v-for="(node, idx) in displayedNodes"
                                              :key="`${node.msg.name}-${node.msg.node_id}`" :node="node"
                                              :default-expanded="defaultExpandedIndex === idx" @request-detail="onRequestDetail"
                                              @request-action-detail="onRequestActionDetail" />
                        </div>
                    </div>
                    <div v-else class="text-xs text-dimmed italic pl-2">
                        No pipeline nodes
                    </div>
                </template>
            </div>
        </template>
    </UCard>

    <RecoDetailModal v-model:open="modalOpen" :reco-id="selectedRecoId" :node-name="selectedRecoName" />
    <ActionDetailModal v-model:open="actionModalOpen" :action-id="selectedActionId" />
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { launchGraph, resetLaunchGraph } from '@/stores/launchGraph'
import { taskDetailActiveIndex, taskDetailFollowLatest } from '@/stores/taskDetail'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import type { PipelineNodeScope } from './taskDetail/types'
import PipelineNodeItem from './taskDetail/PipelineNodeItem.vue'
import RecoDetailModal from './taskDetail/RecoDetailModal.vue'
import ActionDetailModal from './taskDetail/ActionDetailModal.vue'
import { clearCache } from '@/api/http'

const route = useRoute()
const taskDetailSettingsStore = useTaskDetailSettingsStore()
const showOpenAsPage = computed(() => route.path !== '/TaskDetail')

const scrollContainerRef = ref<HTMLElement | null>(null)
const livePage = ref(1)
const historyPage = ref(1)
const viewMode = ref<'live' | 'history'>('live')
const historySnapshotNodes = ref<PipelineNodeScope[]>([])

const allTasks = computed(() => launchGraph.value.childs)
const activeIndex = taskDetailActiveIndex
const followLatest = taskDetailFollowLatest
const reverseNodeOrder = computed(() => taskDetailSettingsStore.reverseNodeOrder)
const nodePageSize = computed(() => Math.max(1, taskDetailSettingsStore.nodePageSize))

const activeTask = computed(() => {
    if (allTasks.value.length === 0) return null
    const idx = Math.min(activeIndex.value, allTasks.value.length - 1)
    return allTasks.value[idx] ?? null
})

const orderedNodes = computed(() => {
    const nodes = activeTask.value?.childs ?? []
    if (!reverseNodeOrder.value) return nodes
    return [...nodes].reverse()
})

const isHistoryMode = computed(() => viewMode.value === 'history')
const effectiveOrderedNodes = computed(() => isHistoryMode.value ? historySnapshotNodes.value : orderedNodes.value)
const totalPages = computed(() => {
    if (effectiveOrderedNodes.value.length === 0) return 1
    return Math.ceil(effectiveOrderedNodes.value.length / nodePageSize.value)
})

const latestPage = computed(() => reverseNodeOrder.value ? 1 : totalPages.value)
const currentPage = computed(() => isHistoryMode.value ? historyPage.value : livePage.value)

const displayedNodes = computed(() => {
    const start = (currentPage.value - 1) * nodePageSize.value
    const end = start + nodePageSize.value
    return effectiveOrderedNodes.value.slice(start, end)
})

const defaultExpandedIndex = computed(() => {
    if (displayedNodes.value.length === 0) return -1
    return reverseNodeOrder.value ? 0 : displayedNodes.value.length - 1
})

function setLivePage(page: number) {
    livePage.value = Math.min(Math.max(page, 1), totalPages.value)
}

function setHistoryPage(page: number) {
    historyPage.value = Math.min(Math.max(page, 1), totalPages.value)
}

function clampPages() {
    const maxPage = Math.max(1, totalPages.value)
    livePage.value = Math.min(Math.max(livePage.value, 1), maxPage)
    historyPage.value = Math.min(Math.max(historyPage.value, 1), maxPage)
}

function enterHistoryMode(targetPage = livePage.value) {
    historySnapshotNodes.value = [...orderedNodes.value]
    viewMode.value = 'history'
    historyPage.value = targetPage
    clampPages()
}

function setCurrentPage(page: number) {
    if (isHistoryMode.value) {
        setHistoryPage(page)
        return
    }
    enterHistoryMode(page)
}

function goToPage(page: number) {
    setCurrentPage(page)
}

function goToLatestPage() {
    viewMode.value = 'live'
    historySnapshotNodes.value = []
    setLivePage(reverseNodeOrder.value ? 1 : totalPages.value)
}

function scrollNodeListToLatestPosition() {
    nextTick(() => {
        if (!scrollContainerRef.value) return
        scrollContainerRef.value.scrollTop = reverseNodeOrder.value ? 0 : scrollContainerRef.value.scrollHeight
    })
}

watch(() => allTasks.value.length, (newLen, oldLen) => {
    if (newLen > (oldLen ?? 0) && followLatest.value) {
        activeIndex.value = newLen - 1
        viewMode.value = 'live'
        setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
        scrollNodeListToLatestPosition()
    }
})

watch(activeIndex, (idx) => {
    followLatest.value = idx === allTasks.value.length - 1
    viewMode.value = 'live'
    historySnapshotNodes.value = []
    setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
    scrollNodeListToLatestPosition()
})

watch([() => activeTask.value?.childs.length, reverseNodeOrder, nodePageSize], () => {
    if (isHistoryMode.value) {
        clampPages()
        return
    }
    setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
    scrollNodeListToLatestPosition()
})

const modalOpen = ref(false)
const selectedRecoId = ref<number | null>(null)
const selectedRecoName = ref<string | null>(null)

function findRecoNameById(recoId: number): string | null {
    for (const task of allTasks.value) {
        for (const pipelineNode of task.childs) {
            for (const nextList of pipelineNode.reco) {
                for (const reco of nextList.childs) {
                    if (reco.msg.reco_id === recoId) {
                        return reco.msg.name
                    }
                }
            }
        }
    }
    return null
}

function onRequestDetail(recoId: number) {
    selectedRecoId.value = recoId
    selectedRecoName.value = findRecoNameById(recoId)
    modalOpen.value = true
}

const actionModalOpen = ref(false)
const selectedActionId = ref<number | null>(null)

function onRequestActionDetail(actionId: number) {
    selectedActionId.value = actionId
    actionModalOpen.value = true
}

async function resetGraph() {
    resetLaunchGraph()
    activeIndex.value = 0
    livePage.value = 1
    historyPage.value = 1
    viewMode.value = 'live'
    followLatest.value = true
    historySnapshotNodes.value = []
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
