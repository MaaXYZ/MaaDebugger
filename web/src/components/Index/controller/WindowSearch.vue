<template>
    <div class="flex flex-col gap-3 w-full min-w-0">
        <!-- Search Filters -->
        <div class="flex flex-row gap-2 min-w-0">
            <UFormField name="class_name" label="Class Name" class="flex-1 min-w-0">
                <UInput v-model="searchFilter.className" placeholder="Window class name"
                    icon="i-lucide-text-cursor-input" class="w-full" :ui="{ base: 'truncate' }" />
            </UFormField>

            <UFormField name="window_regex" label="Window Name Regex" class="flex-1 min-w-0">
                <UInput v-model="searchFilter.windowRegex" placeholder=".*" icon="i-lucide-regex" class="w-full"
                    :ui="{ base: 'truncate' }" />
            </UFormField>
        </div>

        <!-- Window Select -->
        <USelectMenu v-model="selectedHwnd" value-key="value" :items="windowItems" placeholder="Select a window..."
            icon="i-lucide-app-window" class="w-full min-w-0" size="xl" :loading="searching" />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { detectDesktopWindows } from '@/api/http'
import type { Win32WindowInfo } from '@shared/types/api'

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
const windowMap = ref<Record<string, Win32WindowInfo>>({})
const selectedHwnd = ref<string>('')
const searching = ref(false)

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
 * 从持久化数据恢复选中状态。
 * 当 windowItems 为空且有持久化的 hwnd 时，注入一个占位选项，
 * 这样 USelectMenu 就不会显示 "No data"。
 */
function initFromPersisted(hwnd: string, windowName?: string) {
    if (!hwnd) return
    selectedHwnd.value = hwnd
    // 如果 windowItems 已有该项，则不需要再注入
    if (windowItems.value.some(item => item.value === hwnd)) return
    const label = windowName ? `${windowName} - ${hwnd}` : hwnd
    windowItems.value = [{ label, value: hwnd }]
}

/**
 * Trigger search: call backend to detect desktop windows with className/windowRegex filters.
 */
async function onSearch() {
    searching.value = true
    try {
        const windows = await detectDesktopWindows(
            searchFilter.className || undefined,
            searchFilter.windowRegex || undefined,
        )

        // 过滤掉没有 window_name 的窗口
        const validWindows = windows.filter((w) => w.window_name)

        windowMap.value = Object.fromEntries(
            validWindows.map((w) => [w.hwnd, w])
        )

        windowItems.value = validWindows.map((w) => ({
            label: `${w.window_name} - ${w.hwnd}`,
            value: w.hwnd,
        }))

        // 自动选中第一个，如果当前没有选中
        if (windowItems.value.length > 0 && !selectedHwnd.value) {
            selectedHwnd.value = windowItems.value[0]!.value
        }
        // 如果当前选中的不在新列表中，选中第一个
        if (selectedHwnd.value && !windowItems.value.some(item => item.value === selectedHwnd.value)) {
            selectedHwnd.value = windowItems.value[0]?.value ?? ''
        }
    } catch (err) {
        console.error('[WindowSearch] Search failed:', err)
    } finally {
        searching.value = false
    }
}

// Expose for parent component
defineExpose({
    searchFilter,
    selectedHwnd,
    windowItems,
    windowMap,
    updateWindows,
    initFromPersisted,
    onSearch,
    searching,
})
</script>
