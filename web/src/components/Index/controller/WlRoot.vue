<template>
    <div class="flex flex-col gap-3 h-full">
        <!-- Action Buttons Row -->
        <div class="flex flex-row gap-2">
            <UTooltip text="Connect">
                <UButton color="primary" variant="outline" icon="i-lucide-link" size="xl" :loading="connecting"
                    :disabled="!socketPath.trim() || connecting" @click="onConnect" />
            </UTooltip>

            <UTooltip text="Disconnect">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- WlRoot Configuration -->
        <UFormField name="wlroot_socket_path" label="Socket Path">
            <UInput v-model="socketPath" placeholder="/tmp/wlroot.sock" icon="i-lucide-server" class="w-full" />
        </UFormField>
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { connectController, disconnectController } from '@/api/http'
import type { ConnectControllerRequest } from '@/types/api'
import { useControllerStore } from '@/stores/controller'

const emit = defineEmits<{
    (e: 'connected'): void
}>()

const toast = useToast()
const controllerStore = useControllerStore()

// connecting 使用 store 中的全局状态
const connecting = computed({
    get: () => controllerStore.connecting,
    set: (v: boolean) => { controllerStore.connecting = v },
})

const socketPath = computed({
    get: () => controllerStore.wlrootSocketPath ?? '',
    set: (value: string) => {
        controllerStore.updateWlRootConfig({ socketPath: value })
    },
})
/**
 * 连接 WlRoot 设备
 */
async function onConnect() {
    const normalizedSocketPath = socketPath.value.trim()
    if (!normalizedSocketPath) {
        toast.add({
            id: 'ctrl-toast',
            title: 'Controller Connect Failed',
            description: 'Socket Path is required',
            icon: 'i-lucide-circle-x',
            color: 'error',
        })
        return
    }

    connecting.value = true
    try {
        const params: ConnectControllerRequest = {
            type: 'wlroot',
            wlroot_socket_path: normalizedSocketPath,
        }

        const result = await connectController(params)
        if (!result.succeed) {
            console.error('[WlRoot] Connect failed:', result.msg)
            toast.add({
                id: 'ctrl-toast',
                title: 'Controller Connect Failed',
                description: result.msg || 'Unknown error',
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
            return
        }

        toast.add({
            id: 'ctrl-toast',
            title: 'Controller Connected',
            icon: 'i-lucide-check-circle',
            color: 'success',
        })

        emit('connected')

        // 连接成功 → 持久化
        controllerStore.updateWlRootConfig({
            socketPath: normalizedSocketPath,
        })
    } catch (err) {
        console.error('[WlRoot] Connect failed:', err)
        toast.add({
            id: 'ctrl-toast',
            title: 'Controller Connect Failed',
            description: err instanceof Error ? err.message : 'Unknown error',
            icon: 'i-lucide-circle-x',
            color: 'error',
        })
    } finally {
        connecting.value = false
    }
}

/**
 * 断开 Controller
 */
async function onDisconnect() {
    try {
        const result = await disconnectController()
        if (result && !result.succeed) {
            toast.add({
                id: 'ctrl-toast',
                title: 'Controller Disconnect Failed',
                description: result.msg,
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
        } else {
            toast.add({
                id: 'ctrl-toast',
                title: 'Controller Disconnected',
                icon: 'i-lucide-unlink',
                color: 'warning',
            })
        }
    } catch (err) {
        console.error('[WlRoot] Disconnect failed:', err)
        toast.add({
            id: 'ctrl-toast',
            title: 'Controller Disconnect Failed',
            description: err instanceof Error ? err.message : 'Unknown error',
            icon: 'i-lucide-circle-x',
            color: 'error',
        })
    }
}

// Expose for parent component
defineExpose({
    socketPath
})
</script>
