<template>
    <div class="flex flex-col gap-2">
        <UButton
            color="neutral"
            variant="ghost"
            size="xs"
            class="w-fit px-0"
            :icon="open ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
            @click="toggleOpen"
        >
            <template #default>
                <span class="flex items-center gap-2">
                    <UIcon name="i-lucide-workflow" class="size-3.5 text-dimmed" />
                    <span>{{ label }}</span>
                    <UBadge size="xs" color="info" variant="subtle">{{ t('taskDetail.custom') }}</UBadge>
                    <UBadge v-if="contextLabel" size="xs" color="neutral" variant="subtle">{{ contextLabel }}</UBadge>
                    <UBadge size="xs" color="neutral" variant="subtle">{{ formatNodeCount(count) }}</UBadge>
                </span>
            </template>
        </UButton>

        <UCollapsible v-model:open="open" :unmount-on-hide="true">
            <template #content>
                <div class="pl-1">
                    <slot />
                </div>
            </template>
        </UCollapsible>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
    label?: string
    kind?: 'reco' | 'action'
    count: number
}>(), {
    label: 'Internal flow',
    kind: undefined,
})

const { t } = useI18n()

function formatNodeCount(count: number): string {
    return count > 1
        ? t('taskDetail.nodesLabels', { count })
        : t('taskDetail.nodesLabel', { count })
}

const open = ref(false)
const contextLabel = computed(() => {
    switch (props.kind) {
    case 'reco':
        return t('taskDetail.insideReco')
    case 'action':
        return t('taskDetail.insideAction')
    default:
        return ''
    }
})

function toggleOpen() {
    open.value = !open.value
}
</script>
