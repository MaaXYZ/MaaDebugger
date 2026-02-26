<template>
    <div class="flex flex-col gap-1.5 rounded-lg border border-default p-2">
        <!-- Header -->
        <div class="flex flex-row items-center gap-2 flex-wrap">
            <UBadge :color="detail.hit ? 'success' : 'error'" variant="subtle" size="xs">
                {{ detail.hit ? 'Hit' : 'Miss' }}
            </UBadge>
            <UBadge color="info" variant="subtle" size="xs">{{ detail.algorithm }}</UBadge>
            <span class="text-xs font-medium">{{ detail.name }}</span>
        </div>

        <!-- Box -->
        <div v-if="detail.box" class="text-xs text-dimmed">
            Box: [{{ detail.box.x }}, {{ detail.box.y }}, {{ detail.box.w }}, {{ detail.box.h }}]
        </div>

        <!-- Nested Combined Result (recursive And/Or) -->
        <div v-if="detail.combined_result && detail.combined_result.length > 0 && depth < 10"
             class="flex flex-col gap-1.5 mt-1">
            <span class="text-xs text-dimmed">Combined ({{ detail.algorithm }}):</span>
            <div class="pl-2 border-l-2 border-default flex flex-col gap-1.5">
                <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub"
                                :depth="depth + 1" />
            </div>
        </div>

    </div>
</template>

<script setup lang="ts">
import type { RecoDetailResponse } from './types'

defineProps<{
    detail: RecoDetailResponse
    depth: number
}>()
</script>
