<template>
    <UCard class="w-full max-h-[calc(100vh-8rem)] xl:h-[calc(100vh-8rem)]" size="xl"
        :ui="{ root: 'h-full flex flex-col', body: 'flex flex-1 min-h-0 flex-col overflow-hidden' }">
        <template #header>
            <div class="flex min-h-10 items-center gap-2">
                <span class="font-bold">Task Detail</span>
                <div class="min-w-0 flex-1" />
                <div v-if="allTasks.length > 0" class="flex min-w-0 flex-wrap items-center justify-end gap-2">
                    <USelect :model-value="activeIndex" :items="taskSelectItems" value-key="value"
                        class="w-52 min-w-0 max-w-[40vw]" size="sm" arrow @update:model-value="onTaskSelect" />
                    <USelect :model-value="selectedNodeId ?? undefined" :items="nodeSelectItems" value-key="value"
                        class="w-56 min-w-0 max-w-[44vw]" size="sm" arrow :disabled="nodeSelectItems.length === 0"
                        @update:model-value="onNodeSelect" />
                    <UButton v-if="isHistoryMode" size="xs" color="primary" variant="soft"
                        icon="i-lucide-arrow-down-to-line" @click="goToLatestPage">
                        Latest
                    </UButton>
                    <UButton size="xs" variant="ghost" color="neutral" icon="i-lucide-trash-2" @click="resetGraph" />
                </div>
            </div>
        </template>

        <template #default>
            <div class="flex h-full min-h-0 flex-col gap-3">
                <div v-if="allTasks.length === 0" class="flex min-h-80 items-center">
                    <UEmpty icon="i-lucide:list-checks" title="No Task Details"
                        class="w-full rounded-xl border border-dashed border-default bg-default/25 py-10" />
                </div>

                <template v-else-if="activeTask">
                    <TaskDetailDisplayPane :active-task="activeTask" :active-index="activeIndex"
                        :displayed-nodes="displayedNodes" :selected-node-id="selectedNodeId"
                        :entry-node-id="entryNodeId" :current-page="currentPage" :total-pages="totalPages"
                        :total-node-count="totalNodeCount" :reverse-node-order="reverseNodeOrder"
                        :is-history-mode="isHistoryMode" :scroll-request-key="scrollRequestKey" @go-page="goToPage"
                        @go-latest="goToLatestPage" @request-detail="onRequestDetail"
                        @request-action-detail="onRequestActionDetail" />
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
import {
    activeTaskIndex,
    followLatestTask,
    selectedNodeId as taskDetailSelectedNodeId,
    viewMode,
    livePage,
    historyPage,
    historySnapshotNodes,
    resetTaskDetailState,
} from '@/stores/taskDetail'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import TaskDetailDisplayPane from './taskDetail/TaskDetailDisplayPane.vue'
import RecoDetailModal from './taskDetail/RecoDetailModal.vue'
import ActionDetailModal from './taskDetail/ActionDetailModal.vue'
import { clearCache } from '@/api/http'
import { findRecoNameInTasks } from './taskDetail/scopeTree'

const taskDetailSettingsStore = useTaskDetailSettingsStore()

const allTasks = computed(() => launchGraph.value.childs)
const activeIndex = activeTaskIndex
const followLatest = followLatestTask
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
const totalNodeCount = computed(() => effectiveOrderedNodes.value.length)

const latestPage = computed(() => reverseNodeOrder.value ? 1 : totalPages.value)
const currentPage = computed(() => isHistoryMode.value ? historyPage.value : livePage.value)

const displayedNodes = computed(() => {
    const start = (currentPage.value - 1) * nodePageSize.value
    const end = start + nodePageSize.value
    return effectiveOrderedNodes.value.slice(start, end)
})

const taskSelectItems = computed(() =>
    allTasks.value.map((task, index) => ({
        label: `#${index + 1} ${task.msg.entry}`,
        value: index,
    })),
)
const nodeSelectItems = computed(() =>
    displayedNodes.value.map((node) => ({
        label: node.msg.name,
        value: node.msg.node_id,
        muted: `#${node.msg.node_id} · ${node.status}`,
    })),
)

const entryNodeId = computed(() => activeTask.value?.entryNodeId ?? null)
const scrollRequestKey = ref(0)

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
    setLivePage(latestPage.value)
}

function selectTask(index: number) {
    activeIndex.value = index
}

function selectNode(nodeId: number) {
    selectedNodeId.value = nodeId
    scrollRequestKey.value += 1
}

function onTaskSelect(value: string | number | undefined) {
    if (typeof value !== 'number') return
    selectTask(value)
}

function onNodeSelect(value: string | number | undefined) {
    if (typeof value !== 'number') return
    selectNode(value)
}

watch(() => allTasks.value.length, (newLen, oldLen) => {
    if (newLen > (oldLen ?? 0) && followLatest.value) {
        activeIndex.value = newLen - 1
    }
})

watch(activeIndex, (idx, prevIdx) => {
    followLatest.value = idx === allTasks.value.length - 1
    if (idx === prevIdx) return
    viewMode.value = 'live'
    historySnapshotNodes.value = []
    selectedNodeId.value = null
    setLivePage(latestPage.value)
})

watch([() => activeTask.value?.childs.length, reverseNodeOrder, nodePageSize], () => {
    if (isHistoryMode.value) {
        clampPages()
        return
    }
    if (followLatest.value) {
        setLivePage(latestPage.value)
        return
    }
    clampPages()
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

watch(() => allTasks.value.length, (length) => {
    if (length === 0) {
        resetTaskDetailState()
    }
})

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
    resetTaskDetailState()
    await clearCache()
}
</script>
