<script setup lang="ts">
import AboutCard from "@/views/Settings/AboutCard.vue";
import DebugSettingsCard from "@/views/Settings/DebugSettingsCard.vue";
import EditorSettingsCard from "@/views/Settings/EditorSettingsCard.vue";
import KeyboardShortcutsCard from "@/views/Settings/KeyboardShortcutsCard.vue";
import LanguageCard from "@/views/Settings/LanguageCard.vue";
import TaskDetailSettingsCard from "@/views/Settings/TaskDetailSettingsCard.vue";
</script>

<template>
  <UContainer class="py-2">
    <div class="mt-8 flex flex-col gap-6">
      <LanguageCard />
      <DebugSettingsCard />
      <EditorSettingsCard />
      <TaskDetailSettingsCard />
      <KeyboardShortcutsCard />
      <AboutCard />
    </div>
  </UContainer>
</template>
