<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { eventToShortcut, formatShortcut, useShortcutsStore } from '@/stores/shortcuts'
import type { ShortcutAction } from '@/stores/shortcuts'

const { t } = useI18n()
const shortcutsStore = useShortcutsStore()
const recordingAction = ref<ShortcutAction | null>(null)

function shortcutLabel(action: ShortcutAction): string {
    switch (action) {
        case 'task.startStop':
            return t('settings.shortcuts.taskStartStop')
        default:
            return action
    }
}

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
    <UCard size="xl">
        <template #header>
            <div id="keyboard" class="flex flex-row items-center justify-between">
                <span class="font-bold">{{ t('settings.shortcuts.title') }}</span>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" :label="t('settings.shortcuts.resetAll')" size="xs"
                    @click="shortcutsStore.resetAll()" />
            </div>
        </template>

        <div class="flex flex-col gap-3">
            <div v-for="item in shortcutsStore.allShortcuts" :key="item.action"
                class="flex flex-row items-center justify-between gap-4 rounded-lg border border-default p-3">
                <span class="text-sm font-medium">{{ shortcutLabel(item.action) }}</span>

                <div class="flex flex-row items-center gap-2">
                    <button v-if="recordingAction === item.action" autofocus
                        class="flex animate-pulse items-center gap-1 rounded-md border-2 border-primary px-3 py-1.5 text-sm"
                        @keydown="onRecordKeydown" @blur="stopRecording">
                        <UIcon name="i-lucide-keyboard" class="size-4" />
                        <span>{{ t('settings.shortcuts.pressKey') }}</span>
                    </button>

                    <button v-else
                        class="cursor-pointer flex items-center gap-1 rounded-md border border-default px-3 py-1.5 text-sm transition-colors hover:bg-elevated"
                        @click="startRecording(item.action)">
                        <template v-if="item.binding">
                            <UKbd v-for="k in formatShortcut(item.binding)" :key="k" :value="k" />
                        </template>
                        <span v-else class="italic text-dimmed">{{ t('settings.shortcuts.notBound') }}</span>
                    </button>

                    <UTooltip :text="t('settings.shortcuts.clearBinding')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-x" size="xs" :disabled="!item.binding"
                            @click="clearBinding(item.action)" />
                    </UTooltip>

                    <UTooltip :text="t('settings.shortcuts.resetToDefault')">
                        <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" size="xs"
                            @click="resetBinding(item.action)" />
                    </UTooltip>
                </div>
            </div>
        </div>
    </UCard>
</template>
