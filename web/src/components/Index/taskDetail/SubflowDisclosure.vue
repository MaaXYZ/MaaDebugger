<template>
    <div class="flex flex-col gap-2">
        <UButton
            color="neutral"
            variant="ghost"
            size="xs"
            class="w-fit px-0"
            :icon="open ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right'"
            @click="open = !open"
        >
            <template #default>
                <span class="flex items-center gap-2">
                    <span>{{ label }}</span>
                    <UBadge size="xs" color="neutral" variant="subtle">{{ count }}</UBadge>
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
import { ref, watch } from 'vue'

const props = withDefaults(defineProps<{
    label?: string
    count: number
    defaultOpen?: boolean
}>(), {
    label: 'Subflow',
    defaultOpen: false,
})

const open = ref(props.defaultOpen)

watch(() => props.defaultOpen, (value) => {
    if (value) open.value = true
})
</script>
