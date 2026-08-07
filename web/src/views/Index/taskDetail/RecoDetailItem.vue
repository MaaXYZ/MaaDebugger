<template>
    <div class="flex flex-col gap-1.5 rounded-lg border border-default p-2 transition-colors"
        :class="isClickable ? 'cursor-pointer hover:bg-elevated/60' : ''" @click="handleOpenDetail">
        <!-- Header -->
        <div class="flex flex-row items-center gap-2 flex-wrap">
            <UBadge :color="detail.hit ? 'success' : 'error'" variant="subtle" size="xs"
                :label="detail.hit ? t('taskDetail.hit') : t('taskDetail.miss')" />
            <UBadge color="info" variant="subtle" size="xs" :label="detail.algorithm" />
            <span class="text-xs font-medium">{{ detail.name }}</span>
        </div>

        <!-- Box -->
        <div v-if="detail.box" class="text-xs text-dimmed">
            {{ t('taskDetail.box', { x: detail.box.x, y: detail.box.y, w: detail.box.w, h: detail.box.h }) }}
        </div>

        <!-- Nested Combined Result (recursive And/Or) -->
        <div v-if="detail.combined_result && detail.combined_result.length > 0 && depth < 10"
            class="flex flex-col gap-1.5 mt-1" @click.stop>
            <span class="text-xs text-dimmed">{{ t('taskDetail.combined', { algorithm: detail.algorithm }) }}:</span>
            <div class="pl-2 border-l-2 border-default flex flex-col gap-1.5">
                <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub" :depth="depth + 1"
                    @request-detail="$emit('requestDetail', $event)" />
            </div>
        </div>

    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RecoDetailResponse } from '@/types/taskDetail'

const props = defineProps<{
    detail: RecoDetailResponse
    depth: number
}>()

const { t } = useI18n()

const emit = defineEmits<{
    requestDetail: [payload: { recoId: number, name: string }]
}>()

const isClickable = computed(() => !!props.detail.reco_id)

function handleOpenDetail() {
    if (!props.detail.reco_id) return
    emit('requestDetail', {
        recoId: props.detail.reco_id,
        name: props.detail.name,
    })
}
</script>
