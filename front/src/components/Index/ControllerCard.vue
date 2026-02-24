<template>
    <UCard class="" w-full size="xl">
        <template #header>
            <span class="font-bold">Controller</span>
        </template>

        <template #default>
            <USelect v-model="controllerValue" value-key="value" :items="controllerItems" :icon="controllerIcon"
                class="w-full" size="xl" arrow />
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface ControllerItem {
    label: string
    value: string
    icon: string
}

const controllerItems = ref<ControllerItem[]>([
    {
        label: 'ADB',
        value: 'adb',
        icon: 'i-material-symbols:android'
    },
    {
        label: 'Win32',
        value: 'win32',
        icon: 'i-material-symbols:desktop-windows-outline'
    },
    {
        label: 'Gamepad',
        value: 'gamepad',
        icon: 'i-material-symbols:gamepad-outline-rounded'
    },
    {
        label: 'Custom',
        value: 'custom',
        icon: 'i-material-symbols:upload-rounded'
    }
])
const controllerValue = ref<string>(controllerItems.value[0]?.value ?? 'adb')
const controllerIcon = computed<string>(
    () => controllerItems.value.find((item) => item.value === controllerValue.value)?.icon ?? 'i-material-symbols:android'
)
</script>
