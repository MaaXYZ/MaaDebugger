<template>
    <UCard class="w-full xl:min-h-[34rem]" size="xl" :ui="{ root: 'h-full flex flex-col', body: 'flex-1' }">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task Detail</span>
                <div class="flex-1"></div>
                <UButton v-if="allTasks.length > 0" size="xs" variant="ghost" color="neutral" icon="i-lucide-trash-2"
                    @click="resetGraph" />
            </div>
        </template>

        <template #default>
            <div class="flex h-full flex-col gap-3 min-h-0 xl:grid xl:grid-cols-[minmax(0,1.4fr)_minmax(280px,0.9fr)] xl:items-start">
                <div v-if="allTasks.length === 0" class="xl:col-span-2 flex min-h-[20rem] xl:min-h-[28rem] items-center">
                    <UEmpty icon="i-lucide:list-checks" title="No Task Details"
                        class="w-full rounded-xl border border-dashed border-default bg-default/25 py-10" />
                </div>

                <template v-else-if="activeTask">
                    <TaskDetailDisplayPane
                        :active-task="activeTask"
                        :active-index="activeIndex"
                        :displayed-nodes="displayedNodes"
                        :selected-node="selectedNode"
                        :entry-node-id="entryNodeId"
                        :current-page="currentPage"
                        :total-pages="totalPages"
                        :reverse-node-order="reverseNodeOrder"
                        :is-history-mode="isHistoryMode"
                        @go-page="goToPage"
                        @go-latest="goToLatestPage"
                        @request-detail="onRequestDetail"
                        @request-action-detail="onRequestActionDetail"
                    />

                    <TaskDetailNavigatorPane
                        :tasks="allTasks"
                        :active-task="activeTask"
                        :active-index="activeIndex"
                        :displayed-nodes="displayedNodes"
                        :selected-node-id="selectedNodeId"
                        :entry-node-id="entryNodeId"
                        :is-history-mode="isHistoryMode"
                        @select-task="selectTask"
                        @select-node="selectNode"
                        @go-latest="goToLatestPage"
                    />
                </template>
            </div>
        </template>
    </UCard>

    <RecoDetailModal v-model:open="modalOpen" :reco-id="selectedRecoId" :node-name="selectedRecoName" />
    <ActionDetailModal v-model:open="actionModalOpen" :action-id="selectedActionId" />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { launchGraph, resetLaunchGraph } from '@/stores/launchGraph'
import { taskDetailActiveIndex, taskDetailFollowLatest, taskDetailSelectedNodeId } from '@/stores/taskDetail'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import type { PipelineNodeScope } from './taskDetail/types'
import TaskDetailDisplayPane from './taskDetail/TaskDetailDisplayPane.vue'
import TaskDetailNavigatorPane from './taskDetail/TaskDetailNavigatorPane.vue'
import RecoDetailModal from './taskDetail/RecoDetailModal.vue'
import ActionDetailModal from './taskDetail/ActionDetailModal.vue'
import { clearCache } from '@/api/http'
import { findRecoNameInTasks } from './taskDetail/scopeTree'

const taskDetailSettingsStore = useTaskDetailSettingsStore()

const livePage = ref(1)
const historyPage = ref(1)
const viewMode = ref<'live' | 'history'>('live')
const historySnapshotNodes = ref<PipelineNodeScope[]>([])

const allTasks = computed(() => launchGraph.value.childs)
const activeIndex = taskDetailActiveIndex
const followLatest = taskDetailFollowLatest
const selectedNodeId = taskDetailSelectedNodeId
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

const entryNodeId = computed(() => activeTask.value?.entryNodeId ?? null)

const selectedNode = computed(() => {
    if (displayedNodes.value.length === 0) return null
    if (selectedNodeId.value == null) {
        return displayedNodes.value[0] ?? null
    }
    return displayedNodes.value.find((node) => node.msg.node_id === selectedNodeId.value) ?? displayedNodes.value[0] ?? null
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

function selectTask(index: number) {
    activeIndex.value = index
}

function selectNode(nodeId: number) {
    selectedNodeId.value = nodeId
}

watch(() => allTasks.value.length, (newLen, oldLen) => {
    if (newLen > (oldLen ?? 0) && followLatest.value) {
        activeIndex.value = newLen - 1
        viewMode.value = 'live'
        setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
    }
})

watch(activeIndex, (idx) => {
    followLatest.value = idx === allTasks.value.length - 1
    viewMode.value = 'live'
    historySnapshotNodes.value = []
    selectedNodeId.value = null
    setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
})

watch([() => activeTask.value?.childs.length, reverseNodeOrder, nodePageSize], () => {
    if (isHistoryMode.value) {
        clampPages()
        return
    }
    setLivePage(reverseNodeOrder.value ? 1 : latestPage.value)
})

watch(displayedNodes, (nodes) => {
    if (nodes.length === 0) {
        selectedNodeId.value = null
        return
    }
    if (selectedNodeId.value == null || !nodes.some((node) => node.msg.node_id === selectedNodeId.value)) {
        selectedNodeId.value = nodes[0]?.msg.node_id ?? null
    }
}, { immediate: true })

const modalOpen = ref(false)
const selectedRecoId = ref<number | null>(null)
const selectedRecoName = ref<string | null>(null)

function findRecoNameById(recoId: number): string | null {
    return findRecoNameInTasks(allTasks.value, recoId)
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
    selectedNodeId.value = null
    historySnapshotNodes.value = []
    await clearCache()
}
</script>
