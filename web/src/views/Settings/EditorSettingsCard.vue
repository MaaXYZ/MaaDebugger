<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useEditorSettingsStore } from '@/stores/editorSettings'

const { t } = useI18n()
const editorSettingsStore = useEditorSettingsStore()
</script>

<template>
    <UCard size="xl">
        <template #header>
            <div id="editor" class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">{{ t('settings.editor.title') }}</span>
                    <span class="text-sm text-dimmed">{{ t('settings.editor.description') }}</span>
                </div>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" :label="t('settings.editor.reset')" size="xs"
                    @click="editorSettingsStore.reset()" />
            </div>
        </template>

        <div class="flex flex-col gap-6">
            <div id="editor-fontSize" class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">{{ t('settings.editor.fontSize') }}</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedFontSize }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedFontSize" :min="12" :max="24" :step="1"
                    @update:model-value="editorSettingsStore.setFontSize(Number($event))" />
            </div>

            <div id="editor-minHeight" class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">{{ t('settings.editor.minHeight') }}</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedMinHeight }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedMinHeight" :min="360" :max="960" :step="20"
                    @update:model-value="editorSettingsStore.setMinHeight(Number($event))" />
            </div>

            <div id="editor-maxHeight" class="flex flex-col gap-2">
                <div class="flex items-center justify-between gap-3">
                    <span class="text-sm font-medium">{{ t('settings.editor.maxHeight') }}</span>
                    <span class="text-sm text-dimmed tabular-nums">{{ editorSettingsStore.normalizedMaxHeight }}
                        px</span>
                </div>
                <USlider :model-value="editorSettingsStore.normalizedMaxHeight" :min="420" :max="1200" :step="20"
                    @update:model-value="editorSettingsStore.setMaxHeight(Number($event))" />
            </div>
        </div>
    </UCard>
</template>
