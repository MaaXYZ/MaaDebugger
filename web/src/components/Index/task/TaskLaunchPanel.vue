<template>
    <div class="flex flex-col gap-3">
        <div v-if="showTaskModeTabs" class="rounded-lg border border-default bg-elevated/30 p-2">
            <UTabs key="value" v-model="localTaskLaunchMode" :items="taskModeTabItems" class="w-full" />
        </div>

        <div v-if="localTaskLaunchMode === 'manual' || !hasInterfaceTasks" class="flex flex-col gap-3">
            <div class="flex flex-row items-center gap-2">
                <div class="flex-1 min-w-0">
                    <UTooltip :text="selectedEntry">
                        <USelectMenu v-model="localSelectedEntry" v-model:search-term="localEntrySearchTerm" virtualize
                            :items="entrySelectItems" ignore-filter placeholder="Select task entry..." :search-input="{
                                placeholder: 'Filter...',
                                icon: 'i-lucide-search'
                            }" :ui="{ base: 'w-full', content: '!w-auto min-w-(--entry-content-min-w) max-w-[80vw]' }"
                            class="w-full" size="xl" value-key="value" :disabled="isRunning" arrow />
                    </UTooltip>
                </div>
                <UTooltip text="Edit task override">
                    <UButton color="neutral" variant="outline" icon="i-lucide-file-edit" size="xl"
                        :loading="isPreparingOverrideEditor" :disabled="isPreparingOverrideEditor"
                        @click="emit('edit-override')" />
                </UTooltip>
                <UButton v-if="!isRunning" color="success" variant="soft" icon="i-lucide-play" size="xl"
                    :disabled="!canStart" @click="emit('start')">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
                <UButton v-else color="error" variant="soft" icon="i-lucide-square" size="xl" :loading="isStopping"
                    :disabled="isStopping" @click="emit('stop')">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
            </div>
        </div>

        <div v-else-if="localTaskLaunchMode === 'interface' && hasInterfaceTasks" class="flex flex-col gap-3">
            <div v-if="selectedInterfaceTask" class="rounded-lg border border-default bg-elevated/40 p-3 text-sm"
                @click="emit('open-interface-task-modal')">
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <div class="font-medium text-default break-all">
                            {{ selectedInterfaceTask.name }}
                        </div>
                        <div class="text-xs text-dimmed break-all">
                            Entry: {{ effectiveEntry || selectedInterfaceTask.entry || '-' }}
                        </div>
                    </div>
                    <UBadge color="primary" variant="subtle">
                        {{ selectedTaskOptionSelections.length }} option(s)
                    </UBadge>
                </div>
            </div>

            <div class="flex flex-row items-center gap-2">
                <div class="flex-1 min-w-0">
                    <UTooltip :text="selectedInterfaceTask?.name || ''">
                        <USelect :model-value="selectedInterfaceTask?.name || ''" :items="interfaceTaskItems"
                            value-key="value" placeholder="Select interface task..." class="w-full" size="xl" arrow
                            :disabled="isRunning" @update:model-value="onInterfaceTaskSelected" />
                    </UTooltip>
                </div>
                <UTooltip text="Edit task override">
                    <UButton color="neutral" variant="outline" icon="i-lucide-file-edit" size="xl"
                        :loading="isPreparingOverrideEditor" :disabled="isPreparingOverrideEditor"
                        @click="emit('edit-override')" />
                </UTooltip>
                <UButton v-if="!isRunning" color="success" variant="soft" icon="i-lucide-play" size="xl"
                    :disabled="!canStart" @click="emit('start')">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
                <UButton v-else color="error" variant="soft" icon="i-lucide-square" size="xl" :loading="isStopping"
                    :disabled="isStopping" @click="emit('stop')">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { InterfaceTaskCandidate } from '@/types/interface'

interface SelectItem {
    label: string
    value: string
}

interface TaskOptionSelection {
    optionName: string
    caseName: string
}

const props = defineProps<{
    selectedEntry: string
    entrySearchTerm: string
    entrySelectItems: SelectItem[]
    isRunning: boolean
    canStart: boolean
    isStopping: boolean
    startStopKeys: string[]
    isPreparingOverrideEditor: boolean
    hasInterfaceTasks: boolean
    interfaceTaskItems: SelectItem[]
    selectedInterfaceTask: InterfaceTaskCandidate | null
    selectedTaskOptionSelections: TaskOptionSelection[]
    taskLaunchMode: 'manual' | 'interface'
    effectiveEntry: string
}>()

const emit = defineEmits<{
    'update:selectedEntry': [value: string]
    'update:entrySearchTerm': [value: string]
    'update:taskLaunchMode': [value: 'manual' | 'interface']
    'interface-task-selected': [value: string]
    'open-interface-task-modal': []
    'edit-override': []
    start: []
    stop: []
}>()

const showTaskModeTabs = computed(() => props.hasInterfaceTasks)
const taskModeTabItems = [
    {
        label: 'Manual Entry',
        value: 'manual',
        icon: 'i-lucide-list',
    },
    {
        label: 'Interface Task',
        value: 'interface',
        icon: 'i-lucide-list-tree',
    },
]

const localSelectedEntry = computed({
    get: () => props.selectedEntry,
    set: (value: string) => emit('update:selectedEntry', value),
})

const localEntrySearchTerm = computed({
    get: () => props.entrySearchTerm,
    set: (value: string) => emit('update:entrySearchTerm', value),
})

const localTaskLaunchMode = computed({
    get: () => props.taskLaunchMode,
    set: (value: 'manual' | 'interface') => emit('update:taskLaunchMode', value),
})

function onInterfaceTaskSelected(value: string | undefined) {
    if (!value) return
    emit('interface-task-selected', value)
}
</script>
