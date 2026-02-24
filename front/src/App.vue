<script setup lang="ts">
import Index from '@/router/Index.vue';
</script>

<template>
  <UApp :toaster="{ position: 'button-right', progress: false, duration: 8000 }">
    <UMain>
      <UHeader title="MaaDebugger" :ui="{ container: '!px-4 sm:!px-4 lg:!px-0', toggle: 'hidden' }">
        <template #right>
          <UColorModeButton />

          <UTooltip text="Settings">
            <UButton color="neutral" variant="ghost" to="/settings" icon="i-lucide-settings" aria-label="Settings" />
          </UTooltip>

          <UTooltip text="Open on GitHub">
            <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaDebugger" target="_blank"
              icon="i-simple-icons:github" aria-label="GitHub" />
          </UTooltip>
        </template>
      </UHeader>
      <Index />
    </UMain>
  </UApp>
</template>
