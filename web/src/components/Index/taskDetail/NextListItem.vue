<template>
    <div class="flex flex-col gap-1.5 w-full min-w-0">
        <!-- Already resolved reco scopes -->
        <template v-for="(reco, idx) in nextList.childs" :key="`reco-${idx}`">
            <RecoButton :reco="reco" :info="nextList.msg.list?.[idx]" use-warning
                        @request-detail="$emit('requestDetail', $event)" />
        </template>
        <!-- Pending (not yet started) items -->
        <template v-for="(item, idx) in pendingItems" :key="`wait-${idx}`">
            <UButton size="sm" variant="outline" color="neutral" disabled>
                {{ formatItemLabel(item) }}
            </UButton>
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { NextListScope, NextListItem } from './types'
import RecoButton from './RecoButton.vue'

const props = defineProps<{
    nextList: NextListScope
}>()

defineEmits<{
    requestDetail: [recoId: number]
}>()

const pendingItems = computed(() => {
    const list = props.nextList.msg.list ?? []
    return list.slice(props.nextList.childs.length)
})

function formatItemLabel(item: NextListItem): string {
    let label = item.name
    if (item.anchor) label = `[Anchor] ${label}`
    if (item.jump_back) label = `[JumpBack] ${label}`
    return label
}
</script>
