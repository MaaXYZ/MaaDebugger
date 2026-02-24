<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2">
                <span class="font-bold">Controller</span>
                <USelect v-model="controllerValue" value-key="value" :items="controllerItems" :icon="controllerIcon"
                    class="w-full" size="xl" arrow />
            </div>
        </template>

        <template #default>
            <ADB v-if="controllerValue === 'adb'" />
            <Win32 v-else-if="controllerValue === 'win32'" />
            <Gamepad v-else-if="controllerValue === 'gamepad'" />
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ADB from './controller/ADB.vue'
import Win32 from './controller/Win32.vue'
import Gamepad from './controller/Gamepad.vue'

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
