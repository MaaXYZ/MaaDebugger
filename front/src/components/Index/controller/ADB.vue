<template>
    <div class="flex flex-col gap-3">
        <!-- Action Buttons Row -->
        <div class="flex flex-row gap-2">
            <UTooltip text="Detect">
                <UButton color="success" variant="outline" icon="i-lucide-search" size="xl" @click="onDetect" />
            </UTooltip>

            <UTooltip text="Edit">
                <UButton color="neutral" variant="outline" icon="i-lucide-square-pen" size="xl"
                    @click="editModalOpen = true" />
            </UTooltip>

            <UTooltip text="Disconnect">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- Device Select -->
        <USelect v-model="selectedDevice" value-key="value" :items="deviceItems" placeholder="Select a device..."
            icon="i-lucide-smartphone" class="w-full" size="xl" arrow />

        <!-- Edit Modal -->
        <UModal v-model:open="editModalOpen" title="ADB Configuration" description="Configure ADB connection settings">
            <template #body>
                <div class="flex flex-col gap-4">
                    <UFormField name="adb_path" label="ADB Path">
                        <UInput v-model="config.adb_path" placeholder="/path/to/adb" icon="i-lucide-folder"
                            class="w-full" />
                    </UFormField>

                    <UFormField name="adb_address" label="ADB Address">
                        <UInput v-model="config.adb_address" placeholder="127.0.0.1:5555" icon="i-lucide-network"
                            class="w-full" />
                    </UFormField>

                    <UFormField name="screencap" label="Screencap Method">
                        <USelect v-model="config.screencap_method" :items="screencapMethods" class="w-full" />
                    </UFormField>

                    <UFormField name="input" label="Input Method">
                        <USelect v-model="config.input_method" :items="inputMethods" class="w-full" />
                    </UFormField>

                    <UFormField name="extra" label="Extra Config">
                        <UButton color="neutral" variant="outline" icon="i-lucide-file-json" label="Edit JSON"
                            class="w-full" @click="onEditExtra" />
                    </UFormField>
                </div>
            </template>

            <template #footer>
                <div class="flex justify-end gap-2 w-full">
                    <UButton color="neutral" variant="ghost" label="Cancel" @click="editModalOpen = false" />
                    <UButton color="primary" label="Save" @click="onSave" />
                </div>
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

// --- Device Select ---
interface DeviceItem {
    label: string
    value: string
}

const deviceItems = ref<DeviceItem[]>([])
const selectedDevice = ref<string>('')

/**
 * Update device list from backend data.
 */
function updateDevices(devices: DeviceItem[]) {
    deviceItems.value = devices
    const first = devices[0]
    if (first && !selectedDevice.value) {
        selectedDevice.value = first.value
    }
}

// --- Edit Modal ---
const editModalOpen = ref(false)

const screencapMethods = [
    { label: 'Default', value: 0 },
    { label: 'EncodeToFileAndPull', value: 1 },       // 1
    { label: 'Encode', value: 1 << 1 },               // 2
    { label: 'RawWithGzip', value: 1 << 2 },           // 4
    { label: 'RawByNetcat', value: 1 << 3 },           // 8
    { label: 'MinicapDirect', value: 1 << 4 },         // 16
    { label: 'MinicapStream', value: 1 << 5 },         // 32
    { label: 'EmulatorExtras', value: 1 << 6 },        // 64
]

const inputMethods = [
    { label: 'Default', value: 0 },
    { label: 'AdbShell', value: 1 },                   // 1
    { label: 'MinitouchAndAdbKey', value: 1 << 1 },    // 2
    { label: 'Maatouch', value: 1 << 2 },              // 4
    { label: 'EmulatorExtras', value: 1 << 3 },        // 8
]

const config = reactive({
    adb_path: '',
    adb_address: '',
    screencap_method: 0,
    input_method: 0,
})

// --- Actions ---
function onDetect() {
    // TODO: call backend to detect devices
}

function onDisconnect() {
    // TODO: call backend to disconnect
}

function onSave() {
    editModalOpen.value = false
    // TODO: save config to backend
}

function onEditExtra() {
    // TODO: open JSONC editor
}

// Expose for parent component
defineExpose({
    updateDevices,
    config,
    selectedDevice,
})
</script>
