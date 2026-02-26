<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import Index from '@/router/Index.vue'
import { wsClient } from '@/api/ws'
import { getStatusSnapshot } from '@/api/http'
import { useStatusStore } from '@/stores/status'
import { handleTaskEvent } from '@/stores/launchGraph'
import { latestAgentUpdate } from '@/api/agentEvents'

const statusStore = useStatusStore()

onMounted(async () => {
  const snapshot = await getStatusSnapshot()
  if (snapshot) {
    statusStore.updateStatus(snapshot)
  }

  wsClient.connect({
    onStatusUpdate(status) {
      statusStore.updateStatus(status)
    },
    onTaskEvent(event) {
      handleTaskEvent(event)
    },
    onAgentUpdate(agents) {
      latestAgentUpdate.value = agents
    },
  })
})

onUnmounted(() => {
  wsClient.disconnect()
})
</script>

<template>
  <UApp>
    <UMain>
      <UHeader title="MaaDebugger" :ui="{ toggle: 'hidden' }">
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
