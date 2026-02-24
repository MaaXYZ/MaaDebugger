<template>
    <UCard class="w-full" size="xl" :ui="{ body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Controller</span>
                        <UBadge :color="statusColor" :label="statusStore.controllerStatus" variant="subtle" size="sm" />
                    </div>
                    <div class="flex flex-row items-center gap-2">
                        <USelect v-model="controllerValue" value-key="value" :items="controllerItems"
                            :icon="controllerIcon" class="min-w-40" size="xl" arrow
                            :disabled="statusStore.controllerStatus !== 'disconnected'" />
                        <UButton variant="outline" color="neutral" trailing-icon="i-lucide-chevron-down"
                            :data-state="showFullCard ? 'open' : 'closed'" @click="showFullCard = !showFullCard" />
                    </div>
                </div>
                <div v-show="!showFullCard" class="text-sm text-dimmed">
                    {{ summaryText }}
                </div>
            </div>
        </template>

        <template #default>
            <div v-show="showFullCard" class="p-4 sm:p-6 min-h-36">
                <ADB v-show="controllerValue === 'adb'" ref="adbRef" />
                <Win32 v-show="controllerValue === 'win32'" />
                <Gamepad v-show="controllerValue === 'gamepad'" />
            </div>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ADB from './controller/ADB.vue'
import Win32 from './controller/Win32.vue'
import Gamepad from './controller/Gamepad.vue'
import { useStatusStore } from '@/stores/status'
import { useControllerStore } from '@/stores/controller'

const statusStore = useStatusStore()
const controllerStore = useControllerStore()
const showFullCard = ref(true)
const adbRef = ref<InstanceType<typeof ADB> | null>(null)

const statusColor = computed(() => {
    switch (statusStore.controllerStatus) {
        case 'connected':
            return 'success' as const
        case 'connecting':
            return 'warning' as const
        case 'disconnected':
        default:
            return 'neutral' as const
    }
})

interface ControllerItem {
    label: string
    value: string
    icon: string
}

const controllerItems: ControllerItem[] = [
    { label: 'ADB', value: 'adb', icon: 'i-material-symbols:android' },
    { label: 'Win32', value: 'win32', icon: 'i-material-symbols:desktop-windows-outline' },
    { label: 'Gamepad', value: 'gamepad', icon: 'i-material-symbols:gamepad-outline-rounded' },
    { label: 'Custom', value: 'custom', icon: 'i-material-symbols:upload-rounded' },
]

// 双向同步 store.controllerType ↔ controllerValue
const controllerValue = ref<string>(controllerStore.controllerType)

watch(controllerValue, (v) => {
    controllerStore.controllerType = v
})
watch(
    () => controllerStore.controllerType,
    (v) => { if (v !== controllerValue.value) controllerValue.value = v },
    { immediate: true },
)

const controllerIcon = computed<string>(
    () => controllerItems.find((item) => item.value === controllerValue.value)?.icon ?? 'i-material-symbols:android'
)
const controllerLabel = computed<string>(
    () => controllerItems.find((item) => item.value === controllerValue.value)?.label ?? 'ADB'
)

/**
 * 折叠时显示的摘要文本：选中设备的 name (address)
 */
const summaryText = computed(() => {
    if (controllerValue.value === 'adb' && adbRef.value?.selectedDevice) {
        return adbRef.value.selectedDevice
    }
    return controllerLabel.value
})
</script>
