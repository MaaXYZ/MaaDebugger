<template>
    <UIcon v-if="status === 'running'" name="i-lucide-loader" class="size-3.5 animate-spin text-info" />
    <UIcon v-else-if="status === 'success'" name="i-lucide-check" class="size-3.5 text-success" />
    <UIcon v-else-if="status === 'failed'" name="i-lucide-x" class="size-3.5 text-error" />
</template>

<script setup lang="ts">
import type { GeneralStatus } from './types'

defineProps<{
    status: GeneralStatus
}>()
</script>
