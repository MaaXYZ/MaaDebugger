<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import type { NavigationMenuItem } from '@nuxt/ui'
import { wsClient } from '@/api/ws'
import { getStatusSnapshot, getScreenshotStatus, getUACStatus } from '@/api/http'
import { useStatusStore } from '@/stores/status'
import { handleTaskEvent } from '@/stores/launchGraph'
import { latestAgentUpdate } from '@/api/agentEvents'
import { latestFrame, screenshotRunning, screenshotPaused, screenshotFps, screenshotError, screenshotOverlayState, screenshotOverlayMessage } from '@/stores/screenshot'

const BACKEND_DISCONNECT_TOAST_ID = 'backend-disconnected'
const PING_INTERVAL_MS = 5000

const selectTheme = { trailingIcon: 'transition-transform ease-in-out duration-200 group-data-[state=open]:rotate-180' }

const headerNavigationMenuItems = computed<NavigationMenuItem[]>(() => [
    {
        label: 'Home',
        icon: "i-lucide:home",
        to: '/',
    },
    {
        label: 'Running',
        icon: "i-lucide:loader",
        to: '/running',
    },
    {
        label: "Tools",
        icon: "i-lucide:box",
        children: [
            {
                label: "Screenshot",
                icon: "i-lucide:camera",
                description: "Placeholder"
            }]
    }
])

const statusStore = useStatusStore()
const toast = useToast()
const backendConnected = ref(true)
const isUAC = ref(false)
let pingTimer: ReturnType<typeof setInterval> | null = null

async function syncScreenshotStatus() {
    const ss = await getScreenshotStatus()
    if (ss) {
        screenshotRunning.value = ss.running
        screenshotPaused.value = ss.paused
        screenshotFps.value = ss.fps
        screenshotOverlayState.value = ss.overlay_state
        screenshotOverlayMessage.value = ss.overlay_message
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
    await (Promise.all([getStatusSnapshot(), getUACStatus()])).then(
        ([snapshotStatus, uacStatus]) => {
            if (snapshotStatus) {
                statusStore.updateStatus(snapshotStatus)
            }
            isUAC.value = uacStatus
        }
    )
    await syncScreenshotStatus()

    wsClient.connect({
        onStatusUpdate(status) {
            statusStore.updateStatus(status)
            syncScreenshotStatus()
        },
        onTaskEvent(event) {
            handleTaskEvent(event)
        },
        onTaskCompleted(result) {
            if (result.stopped) {
                // 用户主动停止，不需要额外提示
                return
            }
            if (result.success) {
                toast.add({
                    id: 'task-toast',
                    title: 'Task Completed',
                    description: result.entry ? `"${result.entry}" finished successfully` : 'Task finished successfully',
                    icon: 'i-lucide-circle-check',
                    color: 'success',
                })
            } else {
                toast.add({
                    id: 'task-toast',
                    title: 'Task Failed',
                    description: result.error || 'Unknown error',
                    icon: 'i-lucide-circle-x',
                    color: 'error',
                })
            }
        },
        onAgentUpdate(agents) {
            latestAgentUpdate.value = agents
        },
        onScreenshotFrame(data) {
            if (screenshotPaused.value || screenshotOverlayState.value === 'disconnected' || screenshotOverlayState.value === 'failed') {
                return
            }
            screenshotError.value = ''
            screenshotOverlayState.value = 'none'
            screenshotOverlayMessage.value = ''
            latestFrame.value = data
        },
        onScreenshotError(reason) {
            screenshotRunning.value = false
            screenshotError.value = reason
            screenshotOverlayState.value = 'failed'
            screenshotOverlayMessage.value = reason
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
    <UApp
        :toaster="{ position: 'bottom-right', duration: 3000, class: 'whitespace-pre-wrap break-words [overflow-wrap:anywhere]' }">
        <UTheme :ui="{
            select: selectTheme,
        }">
            <UHeader :ui="{ toggle: 'hidden' }">
                <template #left>
                    <div class="flex items-end gap-2">
                        <a href="/" aria-label="MaaDebugger"
                            class="group inline-flex items-end gap-0.5 shrink-0 text-2xl font-black tracking-tight transition-all duration-200 hover:opacity-90 focus-visible:outline-primary">
                            <span class="text-primary">Maa</span>
                            <span class="text-highlighted">Debugger</span>
                        </a>
                        <UBadge v-if="isUAC" label="UAC" color="warning" variant="subtle" size="xs" class="mb-1" />
                    </div>
                </template>

                <template #default>
                    <UNavigationMenu :items="headerNavigationMenuItems" class="w-full justify-center"
                        content-orientation="vertical" highlight />
                </template>

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

            <UMain>
                <RouterView />
            </UMain>
        </UTheme>
    </UApp>
</template>
