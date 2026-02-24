<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import Index from '@/router/Index.vue'
import { wsClient } from '@/api/ws'
import { getStatusSnapshot } from '@/api/http'
import { useStatusStore } from '@/stores/status'

const statusStore = useStatusStore()

onMounted(async () => {
  // 先获取一次当前状态快照
  const snapshot = await getStatusSnapshot()
  if (snapshot) {
    statusStore.updateStatus(snapshot)
  }

  // 启动 WebSocket 连接，后续状态通过 WS 实时更新
  wsClient.connect({
    onStatusUpdate(status) {
      statusStore.updateStatus(status)
    },
  })
})

onUnmounted(() => {
  wsClient.disconnect()
})
</script>

<template>
  <UToaster>
    <UApp :toaster="{ position: 'button-right', progress: false, duration: 8000 }">
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
  </uToaster>
</template>
