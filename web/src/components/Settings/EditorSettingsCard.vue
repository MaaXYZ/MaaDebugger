<script setup lang="ts">
import { useEditorSettingsStore } from '@/stores/editorSettings'

const editorSettingsStore = useEditorSettingsStore()
</script>

<template>
    <UCard size="xl">
        <template #header>
            <div class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">Editor</span>
                    <span class="text-sm text-dimmed">Configure JSON editor size and readability.</span>
                </div>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" label="Reset" size="xs"
                    @click="editorSettingsStore.reset()" />
            </div>
        </template>

        <div class="flex flex-col gap-6">
            <div class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">Editor font size</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedFontSize }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedFontSize" :min="12" :max="24" :step="1"
                    @update:model-value="editorSettingsStore.setFontSize(Number($event))" />
            </div>

            <div class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">Minimum editor height</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedMinHeight }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedMinHeight" :min="360" :max="960" :step="20"
                    @update:model-value="editorSettingsStore.setMinHeight(Number($event))" />
            </div>

            <div class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">Maximum editor height</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedMaxHeight }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedMaxHeight" :min="420" :max="1200" :step="20"
                    @update:model-value="editorSettingsStore.setMaxHeight(Number($event))" />
            </div>
        </div>
    </UCard>
</template>
