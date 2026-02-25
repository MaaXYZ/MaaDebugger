<template>
    <UBadge :color="badgeColor" :variant="badgeVariant" class="gap-1.5">
        <span class="relative flex size-2">
            <span v-if="isPulsing" class="absolute inline-flex size-full animate-ping rounded-full opacity-75"
                :class="pulseColorClass" />
            <span class="relative inline-flex size-2 rounded-full" :class="dotColorClass" />
        </span>
        {{ statusLabel }}
    </UBadge>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TaskStatus } from './types'

const props = defineProps<{
    status: TaskStatus
}>()

const isPulsing = computed(() => props.status === 'running')

const statusLabel = computed(() => {
    switch (props.status) {
        case 'idle': return 'Idle'
        case 'running': return 'Running'
        case 'success': return 'Success'
        case 'failed': return 'Failed'
        case 'stopped': return 'Stopped'
        default: return 'Unknown'
    }
})

const badgeColor = computed(() => {
    switch (props.status) {
        case 'running': return 'info' as const
        case 'success': return 'success' as const
        case 'failed': return 'error' as const
        case 'stopped': return 'warning' as const
        default: return 'neutral' as const
    }
})

const badgeVariant = computed(() => 'subtle' as const)

const dotColorClass = computed(() => {
    switch (props.status) {
        case 'running': return 'bg-info'
        case 'success': return 'bg-success'
        case 'failed': return 'bg-error'
        case 'stopped': return 'bg-warning'
        default: return 'bg-gray-400 dark:bg-gray-500'
    }
})

const pulseColorClass = computed(() => {
    switch (props.status) {
        case 'running': return 'bg-info'
        default: return ''
    }
})
</script>
