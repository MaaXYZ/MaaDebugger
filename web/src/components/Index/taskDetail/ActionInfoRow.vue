<template>
    <div class="flex items-baseline gap-2 text-xs">
        <span class="text-dimmed shrink-0">{{ label }}</span>
        <UKbd v-if="label === 'Keycode' && value !== ''" :value="getKey(Number(value), controller_type ?? '')" />
        <span class="text-right break-all tabular-nums">{{ value }}</span>
    </div>
</template>

<script setup lang="ts">
import getKey from './keymap';

defineProps<{
    label: string
    value: string,
    controller_type?: string
}>()
</script>
