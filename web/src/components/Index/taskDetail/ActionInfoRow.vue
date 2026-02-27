<template>
    <div class="flex gap-2 text-xs">
        <span class="text-dimmed shrink-0">{{ label }}</span>
        <span class="text-right break-all tabular-nums">{{ value }}</span>
    </div>
</template>

<script setup lang="ts">
defineProps<{
    label: string
    value: string
}>()
</script>
