<script setup lang="ts">
import AboutCard from '@/components/Settings/AboutCard.vue'
import EditorSettingsCard from '@/components/Settings/EditorSettingsCard.vue'
import KeyboardShortcutsCard from '@/components/Settings/KeyboardShortcutsCard.vue'
import TaskDetailSettingsCard from '@/components/Settings/TaskDetailSettingsCard.vue'
</script>

<template>
    <UContainer class="py-2">
        <div class="mt-8 flex flex-col gap-6">
            <EditorSettingsCard />
            <TaskDetailSettingsCard />
            <KeyboardShortcutsCard />
            <AboutCard />
        </div>
    </UContainer>
</template>
