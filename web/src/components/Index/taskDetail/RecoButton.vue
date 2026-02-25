<template>
    <UTooltip :text="reco.msg.name">
        <UButton size="sm" :variant="'outline'" :color="btnColor" :icon="btnIcon" :loading="reco.status === 'running'"
            class="font-medium max-w-48" @click="$emit('requestDetail', reco.msg.name)">
            <template #default>
                <span class="truncate">{{ itemBrief }}</span>
            </template>
        </UButton>
    </UTooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RecoScope, NextListItem } from './types'

const props = defineProps<{
    reco: RecoScope
    info?: NextListItem
    useWarning?: boolean
}>()

defineEmits<{
    requestDetail: [name: string]
}>()

const itemBrief = computed(() => {
    let result = props.reco.msg.name
    if (props.info) {
        if (props.info.anchor) result = `[Anchor] ${props.info.name} = ${result}`
        if (props.info.jump_back) result = `[JumpBack] ${result}`
    }
    return result
})

const btnColor = computed(() => {
    if (props.reco.status === 'success') return 'success' as const
    if (props.reco.status === 'failed') {
        return props.useWarning ? 'warning' as const : 'error' as const
    }
    return 'neutral' as const
})

const btnIcon = computed(() => {
    if (props.reco.status === 'success') return 'i-lucide-check'
    if (props.reco.status === 'failed') return 'i-lucide-x'
    return undefined
})
</script>
