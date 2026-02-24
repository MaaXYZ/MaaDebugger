<template>
    <div class="flex flex-col gap-3 h-full">
        <!-- Action Buttons Row -->
        <div class="flex flex-row gap-2">
            <UTooltip text="Detect">
                <UButton color="success" variant="outline" icon="i-lucide-search" size="xl" :loading="detecting"
                    @click="onDetect" />
            </UTooltip>

            <UTooltip text="Connect">
                <UButton color="primary" variant="outline" icon="i-lucide-link" size="xl" :loading="connecting"
                    :disabled="!selectedDevice || connecting" @click="onConnect" />
            </UTooltip>

            <UTooltip text="Edit">
                <UButton color="neutral" variant="outline" icon="i-lucide-square-pen" size="xl" @click="onOpenEdit" />
            </UTooltip>

            <UTooltip text="Disconnect">
                <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl" @click="onDisconnect" />
            </UTooltip>
        </div>

        <!-- Device Select -->
        <div class="flex flex-1 items-center gap-2">
            <USelectMenu v-model="selectedDevice" value-key="value" :items="deviceItems"
                placeholder="Select a device..." icon="i-lucide-smartphone" class="w-full" size="xl" />
        </div>



        <!-- Edit Modal -->
        <UModal v-model:open="editModalOpen" title="ADB Configuration" description="Configure ADB connection settings">
            <template #body>
                <div class="flex flex-col gap-4">
                    <UFormField name="adb_path" label="ADB Path">
                        <UInput v-model="config.adb_path" placeholder="/path/to/adb" icon="i-lucide-folder"
                            class="w-full" />
                    </UFormField>

                    <UFormField name="adb_address" label="ADB Address">
                        <UInput v-model="config.adb_address" placeholder="127.0.0.1:5555" icon="i-lucide-network"
                            class="w-full" />
                    </UFormField>

                    <UFormField name="screencap" label="Screencap Method">
                        <USelect v-model="config.screencap_method" :items="screencapMethods" class="w-full" />
                    </UFormField>

                    <UFormField name="input" label="Input Method">
                        <USelect v-model="config.input_method" :items="inputMethods" class="w-full" />
                    </UFormField>

                    <UFormField name="extra" label="Extra Config">
                        <UButton color="neutral" variant="outline" icon="i-lucide-file-json" label="Edit JSON"
                            class="w-full" @click="onEditExtra" />
                    </UFormField>
                </div>
            </template>

            <template #footer>
                <div class="flex justify-end gap-2 w-full">
                    <UButton color="neutral" variant="ghost" label="Cancel" @click="editModalOpen = false" />
                    <UButton color="primary" icon="i-lucide-link" label="Connect" :loading="connecting"
                        @click="onConnectFromEdit" />
                </div>
            </template>
        </UModal>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { detectAdbDevices, connectController, disconnectController } from '@/api/http'
import type { AdbDeviceInfo, ConnectControllerRequest } from '@shared/types/api'
import { useControllerStore, DEFAULT_SCREENCAP_METHOD, DEFAULT_INPUT_METHOD } from '@/stores/controller'

/** maa-node All（Uint64 字符串） */
const ALL_METHODS = '18446744073709551615'

// --- Device Select ---
interface DeviceItem {
    label: string
    value: string
}

const deviceItems = ref<DeviceItem[]>([])
const deviceMap = ref<Record<string, AdbDeviceInfo>>({})
const controllerStore = useControllerStore()

const selectedDevice = ref<string>(controllerStore.selectedAdbDevice)
const detecting = ref(false)
// connecting 使用 store 中的全局状态，以便 TaskCard 等组件检查
const connecting = computed({
    get: () => controllerStore.connecting,
    set: (v: boolean) => { controllerStore.connecting = v },
})

watch(selectedDevice, (value) => {
    controllerStore.selectedAdbDevice = value
    ensureSelectedItemVisible(value)
})

/**
 * 确保当前选中值即使未检测到设备，也能在下拉列表里可见
 */
function ensureSelectedItemVisible(value: string) {
    if (!value) return
    if (deviceItems.value.some(item => item.value === value)) return
    deviceItems.value = [{ label: value, value }, ...deviceItems.value]
}

// 初始恢复持久化选中值时，先放入下拉选项，避免出现"已选中但列表 No data"
ensureSelectedItemVisible(selectedDevice.value)

/**
 * 从后端检测 ADB 设备
 */
async function onDetect() {
    detecting.value = true
    try {
        const devices = await detectAdbDevices()
        // 过滤掉 address 为空的设备，因为 SelectItem 不允许空字符串 value
        const validDevices = devices.filter(d => d.address)
        deviceMap.value = Object.fromEntries(validDevices.map(d => [`${d.name} (${d.address})`, d]))
        deviceItems.value = validDevices.map(d => ({
            label: `${d.name} (${d.address})`,
            value: `${d.name} (${d.address})`,
        }))

        // 恢复持久化选中项：优先使用持久化的 label，其次按持久化的 adb_address/adb_path 匹配
        if (deviceItems.value.length === 0) {
            selectedDevice.value = ''
            controllerStore.selectedAdbDevice = ''
        } else if (deviceMap.value[controllerStore.selectedAdbDevice]) {
            selectedDevice.value = controllerStore.selectedAdbDevice
        } else {
            const matched = validDevices.find((d) => {
                if (!controllerStore.adbAddress) return false
                return d.address === controllerStore.adbAddress
                    && (!controllerStore.adbPath || d.adb_path === controllerStore.adbPath)
            })

            if (matched) {
                const label = `${matched.name} (${matched.address})`
                selectedDevice.value = label
                controllerStore.selectedAdbDevice = label
            } else if (deviceMap.value[selectedDevice.value]) {
                // 保留当前选中
            } else {
                selectedDevice.value = deviceItems.value[0]!.value
                controllerStore.selectedAdbDevice = selectedDevice.value
            }
        }
    } catch (err) {
        console.error('[ADB] Detect failed:', err)
    } finally {
        detecting.value = false
        ensureSelectedItemVisible(selectedDevice.value)
    }
}

// --- 连接核心逻辑 ---

/**
 * 内部统一连接函数，避免 onConnect / onConnectFromEdit 重复
 */
async function doConnect(params: ConnectControllerRequest): Promise<boolean> {
    if (!params.adb_address) {
        console.error('[ADB] Connect failed: adb_address is empty')
        return false
    }

    connecting.value = true
    try {
        const result = await connectController(params)
        if (!result.success) {
            console.error('[ADB] Connect failed:', result.error)
            return false
        }

        // 连接成功 → 持久化
        controllerStore.selectedAdbDevice = selectedDevice.value
        controllerStore.updateAdbConfig({
            adb_path: params.adb_path ?? '',
            adb_address: params.adb_address ?? '',
            screencap_method: params.adb_screencap_method ?? DEFAULT_SCREENCAP_METHOD,
            input_method: params.adb_input_method ?? DEFAULT_INPUT_METHOD,
            adb_config: params.adb_config,
        })
        return true
    } catch (err) {
        console.error('[ADB] Connect failed:', err)
        return false
    } finally {
        connecting.value = false
    }
}

/**
 * 连接选中的 ADB 设备（工具栏按钮）
 */
async function onConnect() {
    if (!selectedDevice.value) return

    const device = deviceMap.value[selectedDevice.value]

    // 回退链：config（编辑值）→ device（已检测）→ store（持久化）
    await doConnect({
        type: 'adb',
        adb_path: config.adb_path || device?.adb_path || controllerStore.adbPath,
        adb_address: config.adb_address || device?.address || controllerStore.adbAddress,
        adb_screencap_method: config.screencap_method || device?.screencap_methods || controllerStore.screencapMethod,
        adb_input_method: config.input_method || device?.input_methods || controllerStore.inputMethod,
        adb_config: device?.config || controllerStore.adbConfig || '',
    })
}

/**
 * 断开 Controller
 */
async function onDisconnect() {
    try {
        await disconnectController()
    } catch (err) {
        console.error('[ADB] Disconnect failed:', err)
    }
}

/**
 * Update device list from backend data (legacy API for parent).
 */
function updateDevices(devices: DeviceItem[]) {
    deviceItems.value = devices
    const first = devices[0]
    if (first && !selectedDevice.value) {
        selectedDevice.value = first.value
        controllerStore.selectedAdbDevice = first.value
    }
}

// --- Edit Modal ---
const editModalOpen = ref(false)

// 与 maa-node AdbScreencapMethod / AdbInputMethod 常量保持一致
const screencapMethods = [
    { label: 'Default', value: DEFAULT_SCREENCAP_METHOD },
    { label: 'All', value: ALL_METHODS },
    { label: 'EncodeToFileAndPull', value: '1' },
    { label: 'Encode', value: '2' },
    { label: 'RawWithGzip', value: '4' },
    { label: 'RawByNetcat', value: '8' },
    { label: 'MinicapDirect', value: '16' },
    { label: 'MinicapStream', value: '32' },
    { label: 'EmulatorExtras', value: '64' },
]

const inputMethods = [
    { label: 'Default', value: DEFAULT_INPUT_METHOD },
    { label: 'All', value: ALL_METHODS },
    { label: 'AdbShell', value: '1' },
    { label: 'MinitouchAndAdbKey', value: '2' },
    { label: 'Maatouch', value: '4' },
    { label: 'EmulatorExtras', value: '8' },
]

// 编辑弹窗的本地状态，watch immediate 会从 store 同步初始值
const config = reactive({
    adb_path: '',
    adb_address: '',
    screencap_method: DEFAULT_SCREENCAP_METHOD,
    input_method: DEFAULT_INPUT_METHOD,
})

// 持久化异步恢复后，同步到本地 UI 状态
watch(
    () => [
        controllerStore.adbPath,
        controllerStore.adbAddress,
        controllerStore.screencapMethod,
        controllerStore.inputMethod,
    ],
    ([adbPath, adbAddress, screencapMethod, inputMethod]) => {
        config.adb_path = adbPath ?? ''
        config.adb_address = adbAddress ?? ''
        config.screencap_method = screencapMethod ?? DEFAULT_SCREENCAP_METHOD
        config.input_method = inputMethod ?? DEFAULT_INPUT_METHOD
    },
    { immediate: true },
)

watch(
    () => controllerStore.selectedAdbDevice,
    (value) => {
        if (value !== selectedDevice.value) {
            selectedDevice.value = value
        }
    },
    { immediate: true },
)

// --- Actions ---

/**
 * 打开编辑弹窗，并将选中设备的信息填入 config
 */
function onOpenEdit() {
    const device = selectedDevice.value ? deviceMap.value[selectedDevice.value] : null
    if (device) {
        config.adb_path = device.adb_path
        config.adb_address = device.address
        config.screencap_method = device.screencap_methods
        config.input_method = device.input_methods
    }
    editModalOpen.value = true
}

/**
 * 从编辑弹窗连接 ADB，使用 config 中的手动配置值
 */
async function onConnectFromEdit() {
    const ok = await doConnect({
        type: 'adb',
        adb_path: config.adb_path,
        adb_address: config.adb_address,
        adb_screencap_method: config.screencap_method,
        adb_input_method: config.input_method,
        adb_config: '',
    })
    if (ok) {
        editModalOpen.value = false
    }
}

function onEditExtra() {
    // TODO: open JSONC editor
}

// Expose for parent component
defineExpose({
    updateDevices,
    config,
    selectedDevice,
})
</script>
