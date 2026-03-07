<script setup lang="ts">
import { ref } from 'vue'
import { useShortcutsStore, eventToShortcut, formatShortcut } from '@/stores/shortcuts'
import type { ShortcutAction } from '@/stores/shortcuts'
import AboutCard from '@/components/Settings/AboutCard.vue'

const shortcutsStore = useShortcutsStore()

// --- Recording state ---
const recordingAction = ref<ShortcutAction | null>(null)

function startRecording(action: ShortcutAction) {
    recordingAction.value = action
}

function stopRecording() {
    recordingAction.value = null
}

function onRecordKeydown(e: KeyboardEvent) {
    if (!recordingAction.value) return

    e.preventDefault()
    e.stopPropagation()

    // Escape cancels recording
    if (e.key === 'Escape') {
        stopRecording()
        return
    }

    const shortcut = eventToShortcut(e)
    if (shortcut) {
        shortcutsStore.setBinding(recordingAction.value, shortcut)
        stopRecording()
    }
}

function clearBinding(action: ShortcutAction) {
    shortcutsStore.setBinding(action, null)
    stopRecording()
}

function resetBinding(action: ShortcutAction) {
    shortcutsStore.resetBinding(action)
    stopRecording()
}
</script>

<template>
    <UContainer class="py-2">
        <div class="mt-8 flex flex-col gap-6">
            <!-- Keyboard Shortcuts Section -->
            <UCard size="xl">
                <template #header>
                    <div class="flex flex-row items-center justify-between">
                        <span class="font-bold">Keyboard Shortcuts</span>
                        <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" label="Reset All" size="xs"
                            @click="shortcutsStore.resetAll()" />
                    </div>
                </template>

                <div class="flex flex-col gap-3">
                    <div v-for="item in shortcutsStore.allShortcuts" :key="item.action"
                        class="flex flex-row items-center justify-between gap-4 rounded-lg border border-default p-3">
                        <!-- Action label -->
                        <span class="text-sm font-medium">{{ item.label }}</span>

                        <!-- Shortcut display / record area -->
                        <div class="flex flex-row items-center gap-2">
                            <!-- Current binding display or recording prompt -->
                            <button v-if="recordingAction === item.action"
                                class="flex items-center gap-1 rounded-md border-2 border-primary px-3 py-1.5 text-sm animate-pulse"
                                autofocus @keydown="onRecordKeydown" @blur="stopRecording">
                                <UIcon name="i-lucide-keyboard" class="size-4" />
                                <span>Press a key...</span>
                            </button>

                            <button v-else
                                class="flex items-center gap-1 rounded-md border border-default px-3 py-1.5 text-sm cursor-pointer hover:bg-elevated transition-colors"
                                @click="startRecording(item.action)">
                                <template v-if="item.binding">
                                    <UKbd v-for="k in formatShortcut(item.binding)" :key="k" :value="k" />
                                </template>
                                <span v-else class="text-dimmed italic">Not bound</span>
                            </button>

                            <!-- Clear binding -->
                            <UTooltip text="Clear binding">
                                <UButton color="neutral" variant="ghost" icon="i-lucide-x" size="xs"
                                    :disabled="!item.binding" @click="clearBinding(item.action)" />
                            </UTooltip>

                            <!-- Reset to default -->
                            <UTooltip text="Reset to default">
                                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" size="xs"
                                    @click="resetBinding(item.action)" />
                            </UTooltip>
                        </div>
                    </div>
                </div>
            </UCard>

            <AboutCard />
        </div>
    </UContainer>
</template>
