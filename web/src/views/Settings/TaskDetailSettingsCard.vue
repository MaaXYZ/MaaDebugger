<script setup lang="ts">
import { computed } from 'vue'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'

const taskDetailSettingsStore = useTaskDetailSettingsStore()

const nodePageSizeInput = computed({
    get: () => String(taskDetailSettingsStore.nodePageSize),
    set: (value: string | number) => {
        const parsed = Number(value)
        taskDetailSettingsStore.setNodePageSize(parsed)
    },
})
</script>

<template>
    <UCard size="xl">
        <template #header>
            <div class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">Task Detail</span>
                    <span class="text-sm text-dimmed">Configure task detail display and node browsing behavior.</span>
                </div>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" label="Reset" size="xs"
                         @click="taskDetailSettingsStore.reset()" />
            </div>
        </template>

        <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Show Recognition ID</span>
                    <span class="text-sm text-dimmed">Display the ID beside recognition buttons like #400000001</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.showRecoId"
                         @update:model-value="taskDetailSettingsStore.setShowRecoId(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Show Action ID</span>
                    <span class="text-sm text-dimmed">Display the ID beside action buttons like #500000001</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.showActionId"
                         @update:model-value="taskDetailSettingsStore.setShowActionId(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Reverse Node Order</span>
                    <span class="text-sm text-dimmed">Show newer pipeline nodes above older ones by default</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.reverseNodeOrder"
                         @update:model-value="taskDetailSettingsStore.setReverseNodeOrder(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Nodes Per Page</span>
                    <span class="text-sm text-dimmed">Limit how many pipeline nodes are rendered per page</span>
                </div>
                <UInput v-model="nodePageSizeInput" type="number" min="1" step="1" class="w-24" />
            </div>
        </div>
    </UCard>
</template>
