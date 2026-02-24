<template>
    <UCard class="w-full" size="xl" :ui="{ body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <span class="font-bold">Controller</span>
                    <div class="flex flex-row items-center gap-2">
                        <USelect v-model="controllerValue" value-key="value" :items="controllerItems"
                            :icon="controllerIcon" class="min-w-40" size="xl" arrow />
                        <UButton variant="outline" color="neutral"
                            trailing-icon="i-lucide-chevron-down"
                            :data-state="showFullCard ? 'open' : 'closed'"
                            @click="showFullCard = !showFullCard" />
                    </div>
                </div>
                <div v-show="!showFullCard" class="text-sm text-dimmed">
                    {{ controllerLabel }}
                </div>
            </div>
        </template>

        <template #default>
            <UCollapsible :open="showFullCard">
                <template #content>
                    <div class="p-4 sm:p-6 min-h-36">
                        <ADB v-if="controllerValue === 'adb'" />
                        <Win32 v-else-if="controllerValue === 'win32'" />
                        <Gamepad v-else-if="controllerValue === 'gamepad'" />
                    </div>
                </template>
            </UCollapsible>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ADB from './controller/ADB.vue'
import Win32 from './controller/Win32.vue'
import Gamepad from './controller/Gamepad.vue'

const showFullCard = ref(true)

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
const controllerLabel = computed<string>(
    () => controllerItems.value.find((item) => item.value === controllerValue.value)?.label ?? 'ADB'
)
</script>
