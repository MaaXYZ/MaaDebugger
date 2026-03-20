<template>
    <UButton :color="btnColor" :variant="btnVariant" :icon="btnIcon" :size="size"
        class="max-w-full min-w-0 overflow-hidden font-medium"
        :class="fullWidth ? 'w-full justify-start' : 'w-fit justify-start'">
        <template #default>
            <span class="flex min-w-0 items-center gap-1.5" :class="fullWidth ? 'w-full' : 'max-w-full'">
                <span v-if="label" class="truncate">{{ label }}</span>
                <span v-for="metaItem in metaItems" :key="metaItem" class="shrink-0 text-[11px] text-dimmed">{{
                    metaItem }}</span>
            </span>
        </template>
    </UButton>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import type { NodeStatus } from './types'

const props = withDefaults(defineProps<{
    status: NodeStatus
    label?: string
    meta?: string[]
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
    actionId?: number
    fullWidth?: boolean
}>(), {
    label: '',
    tooltip: '',
    meta: () => [],
    size: 'sm',
    actionId: undefined,
    fullWidth: false,
})

const taskDetailSettingsStore = useTaskDetailSettingsStore()

const metaItems = computed(() => {
    const items = [...props.meta]
    if (taskDetailSettingsStore.showActionId && props.actionId !== undefined) {
        items.push(`#${props.actionId}`)
    }
    return items
})


const btnColor = computed(() => {
    switch (props.status) {
        case 'success': return 'success' as const
        case 'failed': return 'error' as const
        case 'running': return 'info' as const
        case 'skipped': return 'warning' as const
        default: return 'neutral' as const
    }
})

const btnVariant = computed(() => 'outline' as const)

const btnIcon = computed(() => {
    switch (props.status) {
        case 'success': return 'i-lucide-check'
        case 'failed': return 'i-lucide-x'
        case 'running': return 'i-lucide-loader'
        case 'skipped': return 'i-lucide-skip-forward'
        default: return undefined
    }
})
</script>
