<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import Index from '@/router/Index.vue'
import { wsClient } from '@/api/ws'
import { getStatusSnapshot, getScreenshotStatus } from '@/api/http'
import { useStatusStore } from '@/stores/status'
import { handleTaskEvent } from '@/stores/launchGraph'
import { latestAgentUpdate } from '@/api/agentEvents'
import { latestFrame, screenshotRunning, screenshotPaused, screenshotFps, screenshotError } from '@/stores/screenshot'

const BACKEND_DISCONNECT_TOAST_ID = 'backend-disconnected'
const PING_INTERVAL_MS = 5000

const statusStore = useStatusStore()
const toast = useToast()
const backendConnected = ref(true)
let pingTimer: ReturnType<typeof setInterval> | null = null

async function syncScreenshotStatus() {
    const ss = await getScreenshotStatus()
    if (ss) {
        screenshotRunning.value = ss.running
        screenshotPaused.value = ss.paused
        screenshotFps.value = ss.fps
    }
}

async function pingBackend() {
    const snapshot = await getStatusSnapshot()
    const connected = snapshot !== null
    if (!connected && backendConnected.value) {
        backendConnected.value = false
        toast.add({
            id: BACKEND_DISCONNECT_TOAST_ID,
            title: 'Disconnected',
            description: 'Please check the service status.',
            icon: 'i-lucide-wifi-off',
            color: 'error',
            duration: 0,
            close: false
        })
    } else if (connected && !backendConnected.value) {
        backendConnected.value = true
        toast.remove(BACKEND_DISCONNECT_TOAST_ID)
    }
}

onMounted(async () => {
    const snapshot = await getStatusSnapshot()
    if (snapshot) {
        statusStore.updateStatus(snapshot)
    }
    await syncScreenshotStatus()

    wsClient.connect({
        onStatusUpdate(status) {
            statusStore.updateStatus(status)
            syncScreenshotStatus()
        },
        onTaskEvent(event) {
            handleTaskEvent(event)
        },
        onAgentUpdate(agents) {
            latestAgentUpdate.value = agents
        },
        onScreenshotFrame(data) {
            screenshotError.value = ''
            latestFrame.value = data
        },
        onScreenshotError(reason) {
            screenshotRunning.value = false
            screenshotError.value = reason
            latestFrame.value = null
            toast.add({
                id: 'screenshot-error',
                title: 'Screenshot Stopped',
                description: reason,
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
        },
    })

    pingTimer = setInterval(pingBackend, PING_INTERVAL_MS)
})

onUnmounted(() => {
    if (pingTimer) {
        clearInterval(pingTimer)
        pingTimer = null
    }
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
                        <UButton color="neutral" variant="ghost" to="/settings" icon="i-lucide-settings"
                            aria-label="Settings" />
                    </UTooltip>

                    <UTooltip text="Open on GitHub">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaDebugger"
                            target="_blank" icon="i-simple-icons:github" aria-label="GitHub" />
                    </UTooltip>
                </template>
            </UHeader>
            <Index />
        </UMain>
    </UApp>

</template>
