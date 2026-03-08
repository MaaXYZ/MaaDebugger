<script setup lang="ts">
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'

const taskDetailSettingsStore = useTaskDetailSettingsStore()
</script>

<template>
    <UCard size="xl">
        <template #header>
            <div class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">Task Detail</span>
                    <span class="text-sm text-dimmed">Configure whether reco and action identifiers are shown.</span>
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
        </div>
    </UCard>
</template>
