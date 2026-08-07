<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'

const { t } = useI18n()
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
            <div id="task" class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">{{ t('settings.taskDetail.title') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.taskDetail.description') }}</span>
                </div>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" :label="t('settings.taskDetail.reset')" size="xs"
                    @click="taskDetailSettingsStore.reset()" />
            </div>
        </template>

        <div class="flex flex-col gap-3">
            <div id="task-showRecoID"
                class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">{{ t('settings.taskDetail.showRecoId') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.taskDetail.showRecoIdDescription') }}</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.showRecoId"
                    @update:model-value="taskDetailSettingsStore.setShowRecoId(Boolean($event))" />
            </div>

            <div id="task-showActionID"
                class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">{{ t('settings.taskDetail.showActionId') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.taskDetail.showActionIdDescription') }}</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.showActionId"
                    @update:model-value="taskDetailSettingsStore.setShowActionId(Boolean($event))" />
            </div>

            <div id="task-reverseNodeOrder"
                class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">{{ t('settings.taskDetail.reverseNodeOrder') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.taskDetail.reverseNodeOrderDescription') }}</span>
                </div>
                <USwitch :model-value="taskDetailSettingsStore.reverseNodeOrder"
                    @update:model-value="taskDetailSettingsStore.setReverseNodeOrder(Boolean($event))" />
            </div>

            <div id="task-nodePageSize"
                class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">{{ t('settings.taskDetail.nodesPerPage') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.taskDetail.nodesPerPageDescription') }}</span>
                </div>
                <UInput v-model="nodePageSizeInput" type="number" min="1" step="1" class="w-24" />
            </div>
        </div>
    </UCard>
</template>
