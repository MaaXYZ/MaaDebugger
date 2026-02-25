<template>
    <div class="flex flex-col gap-3 h-full">
        <!-- Action Buttons Row -->
        <div class="flex flex-row gap-2">
            <UTooltip text="Connect">
                <UButton color="primary" variant="outline" icon="i-lucide-link" size="xl" :loading="connecting"
                    :disabled="!config.address || connecting" @click="onConnect" />
            </UTooltip>

            <UTooltip text="Disconnect">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- PlayCover Configuration -->
        <UFormField name="playcover_address" label="Address">
            <UInput v-model="config.address" placeholder="192.168.1.100" icon="i-lucide-network" class="w-full" />
        </UFormField>

        <UFormField name="playcover_uuid" label="UUID (Optional)">
            <UInput v-model="config.uuid" placeholder="Device UUID (optional)" icon="i-lucide-fingerprint"
                class="w-full" />
        </UFormField>
    </div>
</template>

<script setup lang="ts">
import { reactive, watch, computed } from 'vue'
import { connectController, disconnectController } from '@/api/http'
import type { ConnectControllerRequest } from '@shared/types/api'
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

// 编辑状态
const config = reactive({
    address: '',
    uuid: '',
})

// 持久化异步恢复后，同步到本地 UI 状态
watch(
    () => [controllerStore.playcoverAddress, controllerStore.playcoverUuid],
    ([address, uuid]) => {
        config.address = address ?? ''
        config.uuid = uuid ?? ''
    },
    { immediate: true },
)

// config 变化时自动保存到 store
watch(
    () => [config.address, config.uuid] as const,
    ([address, uuid]) => {
        controllerStore.updatePlayCoverConfig({ address, uuid })
    },
)

/**
 * 连接 PlayCover 设备
 */
async function onConnect() {
    if (!config.address) return

    connecting.value = true
    try {
        const params: ConnectControllerRequest = {
            type: 'playcover',
            playcover_address: config.address,
            playcover_uuid: config.uuid,
        }

        const result = await connectController(params)
        if (!result.succeed) {
            console.error('[PlayCover] Connect failed:', result.msg)
            toast.add({
                title: 'Controller Connect Failed',
                description: result.msg || 'Unknown error',
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
            return
        }

        toast.add({
            title: 'Controller Connected',
            icon: 'i-lucide-check-circle',
            color: 'success',
        })

        emit('connected')

        // 连接成功 → 持久化
        controllerStore.updatePlayCoverConfig({
            address: config.address,
            uuid: config.uuid,
        })
    } catch (err) {
        console.error('[PlayCover] Connect failed:', err)
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
                title: 'Controller Disconnect Failed',
                description: result.msg,
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
        } else {
            toast.add({
                title: 'Controller Disconnected',
                icon: 'i-lucide-unlink',
                color: 'warning',
            })
        }
    } catch (err) {
        console.error('[PlayCover] Disconnect failed:', err)
    }
}

// Expose for parent component
defineExpose({
    config,
})
</script>
