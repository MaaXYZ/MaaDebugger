<template>
    <UButton size="sm" variant="outline" :color="btnColor" :icon="btnIcon" :loading="reco.status === 'running'"
        class="max-w-full min-w-0 justify-start overflow-hidden font-medium"
        @click="$emit('requestDetail', reco.msg.reco_id)">
        <template #default>
            <span class="flex max-w-full min-w-0 items-center gap-1.5 text-left overflow-hidden">
                <span v-if="primaryLabel" class="block min-w-0 truncate">{{ primaryLabel }}</span>
                <span v-for="meta in metaItems" :key="meta" class="shrink-0 text-[11px] text-dimmed">{{ meta
                    }}</span>
            </span>
        </template>
    </UButton>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTaskDetailSettingsStore } from '@/stores/taskDetailSettings'
import type { RecoScope, NextListItem } from '@/types/taskDetail'

const props = defineProps<{
    reco: RecoScope
    info?: NextListItem
    useWarning?: boolean
    algorithmType?: 'And' | 'Or'
}>()

const { t } = useI18n()

defineEmits<{
    requestDetail: [recoId: number]
}>()

const taskDetailSettingsStore = useTaskDetailSettingsStore()

const primaryLabel = computed(() => {
    const label = props.info?.label?.trim()
    if (label) return label
    if (props.info?.anchor) return t('taskDetail.anchor')
    if (props.info?.jump_back) return t('taskDetail.jumpBack')
    return ''
})

const metaItems = computed(() => {
    const items: string[] = []
    if (props.algorithmType) {
        items.push(props.algorithmType)
    }
    if (props.info?.anchor && primaryLabel.value !== t('taskDetail.anchor')) {
        items.push(t('taskDetail.anchor'))
    }
    if (props.info?.jump_back && primaryLabel.value !== t('taskDetail.jumpBack')) {
        items.push(t('taskDetail.jumpBack'))
    }
    if (taskDetailSettingsStore.showRecoId) {
        items.push(`#${props.reco.msg.reco_id}`)
    }
    return items
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
