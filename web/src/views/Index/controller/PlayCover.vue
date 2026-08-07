<template>
    <div class="flex flex-col gap-3 h-full">
        <!-- Action Buttons Row -->
        <div class="flex flex-row gap-2">
            <UTooltip :text="t('common.connect')">
                <UButton color="primary" variant="outline" icon="i-lucide-link" size="xl" :loading="connecting"
                    :disabled="!playcoverAddress.trim() || connecting" @click="onConnect" />
            </UTooltip>

            <UTooltip :text="t('common.disconnect')">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- PlayCover Configuration -->
        <UFormField name="playcover_address" :label="t('playcover.address')">
            <UInput v-model="playcoverAddress" :placeholder="t('playcover.addressPlaceholder')" icon="i-lucide-network" class="w-full" />
        </UFormField>

        <UFormField name="playcover_uuid" :label="t('playcover.uuid')">
            <UInput v-model="playcoverUuid" :placeholder="t('playcover.uuidPlaceholder')" icon="i-lucide-fingerprint"
                class="w-full" />
        </UFormField>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { connectController, disconnectController } from '@/api/http'
import type { ConnectControllerRequest } from '@/types/api'
import { useControllerStore } from '@/stores/controller'

const emit = defineEmits<{
    (e: 'connected'): void
}>()

const toast = useToast()
const { t } = useI18n()
const controllerStore = useControllerStore()

// connecting 使用 store 中的全局状态
const connecting = computed({
    get: () => controllerStore.connecting,
    set: (v: boolean) => { controllerStore.connecting = v },
})

const config = computed({
    get: () => ({
        address: controllerStore.playcoverAddress ?? '',
        uuid: controllerStore.playcoverUuid ?? '',
    }),
    set: (value: { address: string, uuid: string }) => {
        controllerStore.updatePlayCoverConfig({
            address: value.address,
            uuid: value.uuid,
        })
    },
})

const playcoverAddress = computed({
    get: () => config.value.address,
    set: (value: string) => {
        config.value = {
            ...config.value,
            address: value,
        }
    },
})

const playcoverUuid = computed({
    get: () => config.value.uuid,
    set: (value: string) => {
        config.value = {
            ...config.value,
            uuid: value,
        }
    },
})

/**
 * 连接 PlayCover 设备
 */
async function onConnect() {
    const normalizedAddress = playcoverAddress.value.trim()
    if (!normalizedAddress) return

    connecting.value = true
    try {
        const params: ConnectControllerRequest = {
            type: 'playcover',
            playcover_address: normalizedAddress,
            playcover_uuid: playcoverUuid.value.trim(),
        }

        const result = await connectController(params)
        if (!result.succeed) {
            console.error('[PlayCover] Connect failed:', result.msg)
            toast.add({
                id: 'ctrl-toast',
                title: t('common.controllerConnectFailed'),
                description: result.msg || t('common.unknownErrorMsg'),
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
            return
        }

        toast.add({
            id: 'ctrl-toast',
            title: t('common.controllerConnected'),
            icon: 'i-lucide-check-circle',
            color: 'success',
        })

        emit('connected')

        // 连接成功 → 持久化
        controllerStore.updatePlayCoverConfig({
            address: normalizedAddress,
            uuid: playcoverUuid.value.trim(),
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
                id: 'ctrl-toast',
                title: t('common.controllerDisconnectFailed'),
                description: result.msg,
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
        } else {
            toast.add({
                id: 'ctrl-toast',
                title: t('common.controllerDisconnected'),
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
