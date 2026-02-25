<template>
    <UTooltip :text="tooltipText">
        <UButton :color="btnColor" :variant="btnVariant" :icon="btnIcon" :size="size" class="font-medium max-w-48">
            <template #default>
                <span class="truncate">{{ label }}</span>
            </template>
        </UButton>
    </UTooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { NodeStatus } from './types'

const props = withDefaults(defineProps<{
    status: NodeStatus
    label: string
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}>(), {
    size: 'sm',
    badgeText: undefined
})

const tooltipText = computed(() => {
    return props.label
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
