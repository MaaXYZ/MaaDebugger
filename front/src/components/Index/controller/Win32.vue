<template>
    <div class="flex flex-col gap-3 h-full">
        <!-- Action Buttons Row: Search + Edit + Disconnect -->
        <div class="flex flex-row gap-2">
            <UTooltip text="Search Windows">
                <UButton color="success" variant="outline" icon="i-lucide-search" size="xl"
                    @click="windowSearchRef?.onSearch()" />
            </UTooltip>

            <UTooltip text="Edit">
                <UButton color="neutral" variant="outline" icon="i-lucide-square-pen" size="xl"
                    @click="editModalOpen = true" />
            </UTooltip>

            <UTooltip text="Disconnect">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- Window Search (shared component) -->
        <WindowSearch ref="windowSearchRef" />

        <!-- Edit Modal -->
        <UModal v-model:open="editModalOpen" title="Win32 Configuration"
            description="Configure Win32 controller settings">
            <template #body>
                <div class="flex flex-col gap-4">
                    <UFormField name="screencap" label="Screencap Method">
                        <USelect v-model="config.screencap_method" :items="screencapMethods" class="w-full" />
                    </UFormField>

                    <UFormField name="input" label="Input Method">
                        <USelect v-model="config.input_method" :items="inputMethods" class="w-full" />
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
import WindowSearch from './WindowSearch.vue'

const windowSearchRef = ref<InstanceType<typeof WindowSearch> | null>(null)

// --- Edit Modal ---
const editModalOpen = ref(false)

const screencapMethods = [
    { label: 'GDI', value: 1 },                          // 1
    { label: 'FramePool', value: 1 << 1 },               // 2
    { label: 'DXGI_DesktopDup', value: 1 << 2 },         // 4
    { label: 'DXGI_DesktopDup_Window', value: 1 << 3 },  // 8
    { label: 'PrintWindow', value: 1 << 4 },              // 16
    { label: 'ScreenDC', value: 1 << 5 },                 // 32
]

const inputMethods = [
    { label: 'Seize', value: 1 },                                // 1
    { label: 'SendMessage', value: 1 << 1 },                     // 2
    { label: 'PostMessage', value: 1 << 2 },                     // 4
    { label: 'LegacyEvent', value: 1 << 3 },                     // 8
    { label: 'PostThreadMessage', value: 1 << 4 },               // 16
    { label: 'SendMessageWithCursorPos', value: 1 << 5 },        // 32
    { label: 'PostMessageWithCursorPos', value: 1 << 6 },        // 64
    { label: 'SendMessageWithWindowPos', value: 1 << 7 },        // 128
    { label: 'PostMessageWithWindowPos', value: 1 << 8 },        // 256
]

const config = reactive({
    screencap_method: 1,          // default: GDI
    input_method: 1,              // default: Seize
})

// --- Actions ---
function onDisconnect() {
    // TODO: call backend to disconnect
}

function onSave() {
    editModalOpen.value = false
    // TODO: save config to backend
}

// Expose for parent component
defineExpose({
    windowSearchRef,
    config,
})
</script>
