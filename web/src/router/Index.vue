<template>
    <div class="w-full min-h-full flex flex-col gap-4 p-4 lg:p-6">
        <div class="flex items-center justify-between gap-3">
            <UButton color="neutral" variant="ghost" size="sm"
                :icon="showLeftTabs ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'"
                :label="showLeftTabs ? 'Hide Setup' : 'Show Setup'" @click="toggleLeftTabs" />
        </div>

        <div class="grid gap-4 xl:items-start"
            :class="showLeftTabs ? 'xl:grid-cols-[minmax(260px,0.78fr)_minmax(0,1.06fr)_minmax(0,1.06fr)]' : 'xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]'">
            <div v-if="showLeftTabs" class="w-full min-w-0 flex flex-col gap-4 order-3 xl:order-0">
                <LeftTabs />
            </div>

            <div class="w-full min-w-0 flex flex-col gap-4 order-2 xl:order-0 xl:self-start">
                <TaskCard />
            </div>

            <div class="w-full min-w-0 flex flex-col gap-4 order-1 xl:order-0 xl:self-start">
                <TaskDetailCard />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import LeftTabs from '@/components/Index/LeftTabs.vue'
import TaskCard from '@/components/Index/TaskCard.vue'
import TaskDetailCard from '@/components/Index/TaskDetailCard.vue'
import { useDebugWorkspaceSettingsStore } from '@/stores/debugWorkspaceSettings'
import { useStatusStore } from '@/stores/status'

const debugWorkspaceSettingsStore = useDebugWorkspaceSettingsStore()
const statusStore = useStatusStore()

const shouldAutoCollapseLeftTabs = computed(() =>
    debugWorkspaceSettingsStore.autoCollapseLeftTabsOnRunStart && statusStore.taskStatus === 'running',
)

const showLeftTabs = computed(() =>
    !debugWorkspaceSettingsStore.leftTabsCollapsed,
)

watch(shouldAutoCollapseLeftTabs, (shouldAutoCollapse, wasAutoCollapse) => {
    if (!shouldAutoCollapse || wasAutoCollapse || debugWorkspaceSettingsStore.leftTabsCollapsed) return
    debugWorkspaceSettingsStore.setLeftTabsCollapsed(true)
})

function toggleLeftTabs() {
    debugWorkspaceSettingsStore.setLeftTabsCollapsed(showLeftTabs.value)
}
</script>
