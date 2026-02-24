<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Task</span>
                <div class="flex-1" />
                <TaskStatusBadge :status="taskStatus" />
            </div>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <!-- Task Entry SelectMenu + Edit Override -->
                <div class="flex flex-col items-start gap-4">
                    <USelectMenu v-model="selectedEntry" :items="entrySelectItems" placeholder="Select task entry..."
                        :search-input="{
                            placeholder: 'Filter...',
                            icon: 'i-lucide-search'
                        }" class="flex-1 w-full" size="xl" value-key="value" :disabled="isRunning" arrow />
                    <UTooltip text="Edit task override">
                        <UButton color="neutral" variant="outline" icon="i-lucide-file-edit" label="Edit Override"
                            @click="onEditOverride" />
                    </UTooltip>
                </div>
            </div>
        </template>

        <template #footer>
            <div class="flex flex-row gap-2">
                <!-- Start / Stop Button -->
                <UButton v-if="!isRunning" color="success" variant="soft" icon="i-lucide-play" label="Start" block
                    :disabled="!selectedEntry" @click="onStart">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
                <UButton v-else color="error" variant="soft" icon="i-lucide-square" label="Stop" block @click="onStop">
                    <template v-if="startStopKeys.length" #trailing>
                        <UKbd v-for="k in startStopKeys" :key="k" :value="k" />
                    </template>
                </UButton>
            </div>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import TaskStatusBadge from './task/TaskStatusBadge.vue'
import type { TaskStatus } from './task/types'
import { useShortcutsStore, formatShortcut } from '@/stores/shortcuts'

// --- Types ---
interface TaskEntry {
    label: string
    value: string
}

// --- State ---
const shortcutsStore = useShortcutsStore()
const entries = ref<TaskEntry[]>([])
const selectedEntry = ref<string>('')
const taskStatus = ref<TaskStatus>('idle')

// --- Computed ---
const entrySelectItems = computed(() => {
    return entries.value.map(e => ({
        label: e.label,
        value: e.value,
    }))
})

const isRunning = computed(() => {
    return taskStatus.value === 'running'
})

const startStopKeys = computed(() => formatShortcut(shortcutsStore.getBinding('task.startStop')))

// --- Actions ---
function onStart() {
    if (!selectedEntry.value) return
    taskStatus.value = 'running'
}

function onStop() {
    // TODO: Replace with actual task stop logic
    taskStatus.value = 'idle'
}

function onEditOverride() {
    // TODO: Open editor for task override
}

// --- Keyboard Shortcut ---
function onKeydown(e: KeyboardEvent) {
    // Ignore if typing in an input/textarea/select
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
    if ((e.target as HTMLElement)?.isContentEditable) return

    if (shortcutsStore.matches(e, 'task.startStop')) {
        e.preventDefault()
        if (isRunning.value) {
            onStop()
        } else {
            onStart()
        }
    }
}

onMounted(() => {
    window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
    window.removeEventListener('keydown', onKeydown)
})

// --- Public API ---
function getSelectedEntry() {
    return selectedEntry.value
}

function getTaskStatus() {
    return taskStatus.value
}

function setEntries(newEntries: TaskEntry[]) {
    entries.value = newEntries
}

defineExpose({
    selectedEntry,
    taskStatus,
    entries,
    getSelectedEntry,
    getTaskStatus,
    setEntries,
})
</script>
