<template>
    <UModal :open="open" title="Interface Task" :dismissible="false" :ui="{ content: 'sm:max-w-2xl' }"
        @update:open="emit('update:open', $event)">
        <template #body>
            <div class="w-full flex flex-col gap-3">
                <div v-if="selectedTask" class="rounded-lg border border-default bg-elevated/50 p-3 space-y-3">
                    <div>
                        <div class="text-sm font-medium text-default flex flex-wrap items-center gap-x-2 gap-y-1">
                            <span>{{ taskTitle }}</span>
                            <span v-if="taskTitleMuted" class="text-xs text-muted font-normal break-all">{{
                                taskTitleMuted }}</span>
                        </div>
                        <div class="text-xs text-dimmed break-all">
                            Entry: {{ selectedTask.entry || 'n/a' }}
                        </div>
                        <div v-if="taskDescription" class="mt-1 text-xs text-dimmed whitespace-pre-wrap">
                            {{ taskDescription }}
                        </div>
                    </div>
                </div>

                <div v-if="selectedTask && optionDefs.length"
                    class="rounded-lg border border-default bg-elevated/50 p-3 space-y-3">
                    <div class="text-xs font-medium text-default">Options</div>
                    <div v-for="optionDef in optionDefs" :key="optionDef.name" class="space-y-2">
                        <div class="flex items-start justify-between gap-3">
                            <div class="min-w-0">
                                <div
                                    class="text-xs font-medium text-default break-all flex flex-wrap items-center gap-x-2 gap-y-1">
                                    <span>{{ getOptionTitle(optionDef).title }}</span>
                                    <span v-if="getOptionTitle(optionDef).muted"
                                        class="text-[11px] text-muted font-normal break-all">{{
                                        getOptionTitle(optionDef).muted }}</span>
                                </div>
                                <div v-if="getOptionDescription(optionDef)"
                                    class="text-xs text-dimmed whitespace-pre-wrap break-all">
                                    {{ getOptionDescription(optionDef) }}
                                </div>
                            </div>
                            <UBadge color="neutral" variant="subtle" size="md">
                                {{ optionDef.type || 'option' }}
                            </UBadge>
                        </div>

                        <div v-if="optionDef.type === 'switch'"
                            class="rounded-md border border-default bg-default/30 px-3 py-2">
                            <USwitch :model-value="isSwitchChecked(optionDef)"
                                @update:model-value="(value) => onSwitchUpdated(optionDef, value)">
                                <template #label>
                                    <div class="text-sm text-default flex flex-wrap items-center gap-x-2 gap-y-1">
                                        <span>{{ getSwitchLabel(optionDef).title }}</span>
                                        <span v-if="getSwitchLabel(optionDef).muted" class="text-xs text-muted">{{
                                            getSwitchLabel(optionDef).muted }}</span>
                                    </div>
                                </template>
                            </USwitch>
                        </div>

                        <USelectMenu v-else :model-value="selectedCaseMap[optionDef.name]"
                            :items="buildOptionCaseItems(optionDef)" value-key="value" class="w-full" arrow
                            @update:model-value="(value) => onCaseSelected(optionDef.name, value)" />
                    </div>
                </div>
            </div>
        </template>
        <template #footer>
            <div class="flex w-full justify-end gap-2">
                <UButton color="neutral" variant="ghost" @click="emit('cancel')">Cancel</UButton>
                <UButton color="primary" @click="emit('confirm')">Confirm</UButton>
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTaskStore } from '@/stores/task'
import type { InterfaceTaskCandidate, InterfaceTaskOptionCase, InterfaceTaskOptionDefinition } from '@/types/interface'

const props = defineProps<{
    open: boolean
    selectedTask: InterfaceTaskCandidate | null
    optionDefs: InterfaceTaskOptionDefinition[]
    selectedCaseMap: Record<string, string>
}>()

const emit = defineEmits<{
    'update:open': [value: boolean]
    'select-case': [optionName: string, caseName: string]
    cancel: []
    confirm: []
}>()

const taskStore = useTaskStore()

const taskTitle = computed(() => taskStore.getDisplayName(props.selectedTask?.name, props.selectedTask?.label))
const taskTitleMuted = computed(() => {
    const displayName = taskTitle.value.trim()
    const rawName = props.selectedTask?.name?.trim() || ''
    return displayName && displayName !== rawName ? rawName : ''
})
const taskDescription = computed(() => taskStore.resolveInterfaceText(props.selectedTask?.description))

function buildDisplayTitle(name?: string, label?: string) {
    const rawName = (name ?? '').trim()
    const displayName = taskStore.getDisplayName(name, label).trim()
    return {
        title: displayName || rawName,
        muted: displayName && displayName !== rawName ? rawName : '',
    }
}

function getOptionTitle(optionDef: InterfaceTaskOptionDefinition) {
    return buildDisplayTitle(optionDef.name, optionDef.label)
}

function getOptionDescription(optionDef: InterfaceTaskOptionDefinition) {
    return taskStore.resolveInterfaceText(optionDef.description)
}

function getCaseTitle(item: InterfaceTaskOptionCase) {
    return buildDisplayTitle(item.name, item.label)
}

function buildOptionCaseItems(optionDef: InterfaceTaskOptionDefinition) {
    return (optionDef.cases ?? []).map((item) => {
        const caseTitle = getCaseTitle(item)
        const description = taskStore.resolveInterfaceText(item.description)
        return {
            label: caseTitle.title,
            value: item.name,
            description,
            muted: caseTitle.muted,
        }
    })
}

function normalizeCaseName(value: string | undefined) {
    return (value ?? '').trim().toLowerCase()
}

function resolveSwitchCases(optionDef: InterfaceTaskOptionDefinition) {
    const cases = optionDef.cases ?? []
    const yesCase = cases.find((item) => ['yes', 'y'].includes(normalizeCaseName(item.name))) ?? cases[0]
    const noCase = cases.find((item) => ['no', 'n'].includes(normalizeCaseName(item.name))) ?? cases[1]
    return { yesCase, noCase }
}

function getSwitchLabel(optionDef: InterfaceTaskOptionDefinition) {
    const { yesCase, noCase } = resolveSwitchCases(optionDef)
    const selectedName = props.selectedCaseMap[optionDef.name]
    if (yesCase && selectedName === yesCase.name) return getCaseTitle(yesCase)
    if (noCase && selectedName === noCase.name) return getCaseTitle(noCase)
    return getCaseTitle(yesCase ?? { name: optionDef.name })
}

function isSwitchChecked(optionDef: InterfaceTaskOptionDefinition) {
    const { yesCase } = resolveSwitchCases(optionDef)
    return props.selectedCaseMap[optionDef.name] === yesCase?.name
}

function onCaseSelected(optionName: string, value: string | undefined) {
    if (!value) return
    emit('select-case', optionName, value)
}

function onSwitchUpdated(optionDef: InterfaceTaskOptionDefinition, value: boolean | 'indeterminate') {
    const { yesCase, noCase } = resolveSwitchCases(optionDef)
    const targetCase: InterfaceTaskOptionCase | undefined = value === true ? yesCase : noCase
    if (!targetCase?.name) return
    emit('select-case', optionDef.name, targetCase.name)
}
</script>
