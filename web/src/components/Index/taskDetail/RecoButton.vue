<template>
    <UTooltip :text="reco.msg.name" class="inline-flex max-w-full min-w-0">
        <UButton size="sm" :variant="'outline'" :color="btnColor" :icon="btnIcon" :loading="reco.status === 'running'"
                 class="font-medium max-w-full min-w-0 justify-start overflow-hidden"
                 @click="$emit('requestDetail', reco.msg.reco_id)">
            <template #default>
                <span class="flex items-center gap-1 max-w-full min-w-0 text-left overflow-hidden">
                    <span class="truncate block min-w-0">{{ itemBrief }}</span>
                    <span class="text-[11px] text-dimmed shrink-0">#{{ reco.msg.reco_id }}</span>
                </span>
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
    requestDetail: [recoId: number]
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
