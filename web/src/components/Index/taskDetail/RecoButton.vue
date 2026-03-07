<template>
    <UTooltip :text="reco.msg.name" class="inline-flex max-w-full min-w-0">
        <UButton size="sm" :variant="'outline'" :color="btnColor" :icon="btnIcon" :loading="reco.status === 'running'"
            class="font-medium max-w-full min-w-0 justify-start overflow-hidden"
            @click="$emit('requestDetail', reco.msg.reco_id)">
            <template #default>
                <span class="flex items-center gap-1 max-w-full min-w-0 text-left overflow-hidden">
                    <span class="truncate block min-w-0">{{ itemBrief }}</span>
                    <UBadge v-if="algorithmType" size="xs" color="info" variant="subtle" class="shrink-0">
                        {{ algorithmType }}
                    </UBadge>
                    <span v-if="taskDetailSettingsStore.showRecoId" class="text-[11px] text-dimmed shrink-0">#{{
                        reco.msg.reco_id }}</span>
                </span>
            </template>
        </UButton>
    </UTooltip>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getRecoDetailById } from '@/api/http'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import type { RecoScope, NextListItem } from './types'

const props = defineProps<{
    reco: RecoScope
    info?: NextListItem
    useWarning?: boolean
}>()

defineEmits<{
    requestDetail: [recoId: number]
}>()

const taskDetailSettingsStore = useTaskDetailSettingsStore()
const algorithmType = ref<'And' | 'Or' | null>(null)

watch(
    () => props.reco.msg.reco_id,
    async (recoId) => {
        algorithmType.value = null
        try {
            const detail = await getRecoDetailById(recoId)
            if (detail?.algorithm === 'And' || detail?.algorithm === 'Or') {
                algorithmType.value = detail.algorithm
            }
        } catch {
            algorithmType.value = null
        }
    },
    { immediate: true },
)

const itemBrief = computed(() => {
    if (!props.info) return props.reco.msg.name

    const label = props.info.label?.trim() || props.info.name
    if (label === props.reco.msg.name) return label
    return `${label} = ${props.reco.msg.name}`
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
