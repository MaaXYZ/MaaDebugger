<template>
    <div class="w-full min-h-full flex flex-col gap-4 p-4 lg:p-6">
        <div class="flex items-center justify-between gap-3">
            <UButton color="neutral" variant="ghost" size="sm"
                :icon="showLeftTabs ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'"
                :label="showLeftTabs ? 'Hide Setup' : 'Show Setup'" @click="toggleLeftTabs" />
        </div>

        <div class="flex flex-col xl:flex-row xl:items-start">
            <div class="order-3 min-w-0 overflow-hidden motion-reduce:transition-none xl:order-0 xl:shrink-0 xl:w-[clamp(260px,28vw,32rem)]"
                :class="showLeftTabs
                    ? 'mb-4 max-h-800 max-w-full opacity-100 transition-[max-height,opacity,margin-bottom] duration-200 ease-out xl:mb-0 xl:mr-4 xl:max-h-none xl:max-w-lg xl:transition-[max-width,opacity,margin-right]'
                    : 'pointer-events-none mb-0 max-h-0 max-w-0 opacity-0 transition-[max-height,opacity,margin-bottom] duration-150 ease-in xl:mr-0 xl:max-h-none xl:transition-[max-width,opacity,margin-right]'">
                <LeftTabs />
            </div>

            <div class="order-1 min-w-0 flex-1 grid gap-4 xl:order-0 xl:grid-cols-2 xl:items-start">
                <div class="w-full min-w-0 flex flex-col gap-4 xl:self-start">
                    <TaskCard />
                </div>

                <div class="w-full min-w-0 flex flex-col gap-4 xl:self-start">
                    <TaskDetailCard />
                </div>
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
