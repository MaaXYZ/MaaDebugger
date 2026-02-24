<template>
    <div class="flex flex-col gap-3">
        <!-- Search Filters -->
        <div class="flex flex-row gap-2">
            <UFormField name="class_name" label="Class Name" class="flex-1">
                <UInput v-model="searchFilter.className" placeholder="Window class name"
                    icon="i-lucide-text-cursor-input" class="w-full" />
            </UFormField>

            <UFormField name="window_regex" label="Window Name Regex" class="flex-1">
                <UInput v-model="searchFilter.windowRegex" placeholder=".*" icon="i-lucide-regex" class="w-full" />
            </UFormField>
        </div>

        <!-- Window Select -->
        <USelect v-model="selectedHwnd" value-key="value" :items="windowItems" placeholder="Select a window..."
            icon="i-lucide-app-window" class="w-full" size="xl" arrow />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

// --- Window Search Filter ---
const searchFilter = reactive({
    className: '',
    windowRegex: '',
})

// --- Window List ---
interface WindowItem {
    label: string
    value: string // hwnd as string
}

const windowItems = ref<WindowItem[]>([])
const selectedHwnd = ref<string>('')

/**
 * Update window list from backend data.
 */
function updateWindows(windows: WindowItem[]) {
    windowItems.value = windows
    const first = windows[0]
    if (first && !selectedHwnd.value) {
        selectedHwnd.value = first.value
    }
}

/**
 * Trigger search (called by parent component).
 */
function onSearch() {
    // TODO: call backend to search windows with className and windowRegex
}

// Expose for parent component
defineExpose({
    searchFilter,
    selectedHwnd,
    windowItems,
    updateWindows,
    onSearch,
})
</script>
