<template>
    <UBadge :color="badgeColor" :variant="badgeVariant" class="gap-1.5">
        <span class="relative flex size-2">
            <span v-if="isPulsing" class="absolute inline-flex size-full animate-ping rounded-full opacity-75"
                  :class="pulseColorClass"></span>
            <span class="relative inline-flex size-2 rounded-full" :class="dotColorClass"></span>
        </span>
        {{ statusLabel }}
    </UBadge>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ConnectionStatus } from './types'

const props = defineProps<{
    status: ConnectionStatus
}>()

const { t } = useI18n()

const isPulsing = computed(() => props.status === 'connecting')

const statusLabel = computed(() => {
    switch (props.status) {
    case 'idle': return t('common.idle')
    case 'connecting': return t('controller.connecting')
    case 'connected': return t('controller.connected')
    case 'failed': return t('common.failed')
    default: return t('common.unknown')
    }
})

const badgeColor = computed(() => {
    switch (props.status) {
    case 'connected': return 'success' as const
    case 'failed': return 'error' as const
    case 'connecting': return 'info' as const
    default: return 'neutral' as const
    }
})

const badgeVariant = computed(() => 'subtle' as const)

const dotColorClass = computed(() => {
    switch (props.status) {
    case 'connected': return 'bg-success'
    case 'failed': return 'bg-error'
    case 'connecting': return 'bg-info'
    default: return 'bg-gray-400 dark:bg-gray-500'
    }
})

const pulseColorClass = computed(() => {
    switch (props.status) {
    case 'connecting': return 'bg-info'
    default: return ''
    }
})
</script>
