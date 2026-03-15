<template>
    <UCard class="w-full max-w-xl transition-opacity duration-200"
        :class="{ 'opacity-50 pointer-events-none': isTaskRunning }" size="xl" :ui="{ body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Controller</span>
                        <UBadge :color="statusColor" variant="subtle" size="sm" class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="statusStore.controllerStatus === 'connecting'"
                                    class="absolute inline-flex size-full animate-ping rounded-full bg-warning opacity-75"></span>
                                <span class="relative inline-flex size-2 rounded-full" :class="dotClass"></span>
                            </span>
                            {{ capitalizedStatus }}
                        </UBadge>
                    </div>
                    <div class="flex flex-row items-center gap-2">
                        <USelect v-model="controllerValue" value-key="value" :items="controllerItems"
                            :icon="controllerIcon" class="min-w-40" size="xl" arrow />
                    </div>
                </div>
            </div>
        </template>

        <div class="p-4 sm:p-6 min-h-36">
            <!-- ADB -->
            <ADB v-show="controllerValue === 'adb'" ref="adbRef" />

            <!-- PlayCover -->
            <PlayCover v-show="controllerValue === 'playcover'" ref="playcoverRef" />

            <!-- Win32 / Gamepad: 共享 WindowSearch + screencap + 各自独有配置 -->
            <div v-show="isDesktopType" class="flex flex-col gap-3 h-full">
                <!-- Action Buttons Row -->
                <div class="flex flex-row gap-2">
                    <UTooltip text="Search Windows">
                        <UButton color="success" variant="outline" icon="i-lucide-search" size="xl"
                            :loading="windowSearchRef?.searching" @click="windowSearchRef?.onSearch()" />
                    </UTooltip>

                    <UTooltip text="Connect">
                        <UButton color="primary" variant="outline" icon="i-lucide-link" size="xl"
                            :loading="controllerStore.connecting"
                            :disabled="!windowSearchRef?.selectedHwnd || controllerStore.connecting"
                            @click="onConnect" />
                    </UTooltip>

                    <UTooltip text="Disconnect">
                        <UButton color="error" variant="outline" icon="i-lucide-unlink" size="xl"
                            @click="onDisconnect" />
                    </UTooltip>
                </div>

                <!-- Shared WindowSearch -->
                <div class="flex flex-1 items-center gap-2">
                    <WindowSearch ref="windowSearchRef" />
                </div>

                <!-- Shared Screencap Method -->
                <UFormField name="screencap" label="Screencap Method">
                    <USelect v-model="desktopScreencap" :items="screencapMethods" class="w-full" arrow />
                </UFormField>

                <!-- Win32 独有：Mouse + Keyboard -->
                <template v-if="controllerValue === 'win32'">
                    <UFormField name="mouse" label="Mouse Method">
                        <USelect v-model="win32Config.mouse_method" :items="inputMethods" class="w-full" arrow :ui="{
                            trailingIcon: 'group-data-[state=open]:rotate-180 transition-transform duration-200'
                        }" />
                    </UFormField>

                    <UFormField name="keyboard" label="Keyboard Method">
                        <USelect v-model="win32Config.keyboard_method" :items="inputMethods" class="w-full" arrow />
                    </UFormField>
                </template>

                <!-- Gamepad 独有：Gamepad Type -->
                <template v-if="controllerValue === 'gamepad'">
                    <UFormField name="gamepad_type" label="Gamepad Type">
                        <USelect v-model="gamepadConfig.gamepad_type" :items="gamepadTypes"
                            :icon="gamepadConfig.gamepad_icon" class="w-full" arrow />
                    </UFormField>
                </template>
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref, reactive, watch, onMounted } from 'vue'
import ADB from './controller/ADB.vue'
import PlayCover from './controller/PlayCover.vue'
import WindowSearch from './controller/WindowSearch.vue'
import { useStatusStore } from '@/stores/status'
import {
    useControllerStore,
    DEFAULT_DESKTOP_SCREENCAP,
    DEFAULT_WIN32_MOUSE,
    DEFAULT_WIN32_KEYBOARD,
    DEFAULT_GAMEPAD_TYPE,
    DEFAULT_GAMEPAD_ICON,
} from '@/stores/controller'
import { connectController, disconnectController, getControllerMethod } from '@/api/http'
import { type MethodItems, type ConnectControllerRequest } from '@/types/api'

const toast = useToast()
const statusStore = useStatusStore()
const controllerStore = useControllerStore()
const isTaskRunning = computed(() => statusStore.taskStatus === 'running')
const windowSearchRef = ref<InstanceType<typeof WindowSearch> | null>(null)

function capitalize(s: string): string {
    return s.charAt(0).toUpperCase() + s.slice(1)
}

/** 是否曾经尝试过连接（用于区分 Idle 和 Disconnected） */
const hasAttemptedConnection = ref(false)

const capitalizedStatus = computed(() => {
    switch (statusStore.controllerStatus) {
        case 'connected':
            return 'Connected'
        case 'connecting':
            return 'Connecting'
        case 'disconnected':
            return hasAttemptedConnection.value ? 'Disconnected' : 'Idle'
        default:
            return capitalize(statusStore.controllerStatus)
    }
})

const statusColor = computed(() => {
    switch (statusStore.controllerStatus) {
        case 'connected':
            return 'success' as const
        case 'connecting':
            return 'warning' as const
        case 'disconnected':
            return hasAttemptedConnection.value ? 'error' as const : 'neutral' as const
        default:
            return 'neutral' as const
    }
})

const dotClass = computed(() => {
    switch (statusStore.controllerStatus) {
        case 'connected':
            return 'bg-success'
        case 'connecting':
            return 'bg-warning'
        case 'disconnected':
            return hasAttemptedConnection.value ? 'bg-error' : 'bg-gray-400 dark:bg-gray-500'
        default:
            return 'bg-gray-400 dark:bg-gray-500'
    }
})

// 状态变化时记录是否尝试过连接
watch(() => statusStore.controllerStatus, (newStatus, oldStatus) => {
    if (!oldStatus || newStatus === oldStatus) return

    // 标记曾经尝试连接
    if (newStatus === 'connecting' || newStatus === 'connected') {
        hasAttemptedConnection.value = true
    }
})

interface ControllerItem {
    label: string
    value: string
    icon: string
}

const controllerItems: ControllerItem[] = [
    { label: 'ADB', value: 'adb', icon: 'i-simple-icons:android' },
    { label: 'Win32', value: 'win32', icon: 'i-simple-icons:windows' },
    { label: 'Gamepad', value: 'gamepad', icon: 'i-lucide:gamepad-directional' },
    { label: "PlayCover", value: "playcover", icon: "i-simple-icons:apple" },
    { label: 'Custom', value: 'custom', icon: 'i-lucide:upload' },
]

// 双向同步 store.controllerType ↔ controllerValue
const controllerValue = ref<string>(controllerStore.controllerType)

// 是否为桌面类型（Win32 / Gamepad）
const isDesktopType = computed(() =>
    controllerValue.value === 'win32' || controllerValue.value === 'gamepad'
)

watch(controllerValue, (v, oldV) => {
    controllerStore.controllerType = v
    // 切换控制器类型时，如果当前已连接则自动断连
    if (oldV && v !== oldV && statusStore.controllerStatus !== 'disconnected') {
        disconnectController().catch((err) => {
            console.error('[Controller] Auto-disconnect on type switch failed:', err)
        })
    }
})
watch(
    () => controllerStore.controllerType,
    (v) => { if (v !== controllerValue.value) controllerValue.value = v },
    { immediate: true },
)

const controllerIcon = computed<string>(
    () => controllerItems.find((item) => item.value === controllerValue.value)?.icon ?? 'i-simple-icons:android'
)
// --- 桌面共享配置 ---
const desktopScreencap = ref(DEFAULT_DESKTOP_SCREENCAP)

// 从 store 恢复截图方法
watch(
    () => controllerStore.desktopScreencapMethod,
    (v) => { desktopScreencap.value = v ?? DEFAULT_DESKTOP_SCREENCAP },
    { immediate: true },
)

// 截图方法变化时保存到 store
watch(desktopScreencap, (v) => {
    controllerStore.desktopScreencapMethod = v
})

// 从后端获取 Methods
const screencapMethods = ref<MethodItems[]>([])
const inputMethods = ref<MethodItems[]>([])
const gamepadTypes = ref<MethodItems[]>([])
onMounted(async () => {
    try {
        const [screencapResp, inputResp, gamepadResp] = await Promise.all([
            getControllerMethod('window_screencap'),
            getControllerMethod('win32_input'),
            getControllerMethod('gamepad_type'),
        ])
        screencapMethods.value = screencapResp.data ?? []
        inputMethods.value = inputResp.data ?? []
        gamepadTypes.value = gamepadResp.data ?? []
    } catch (err) {
        console.error('[ADB] Load controller methods failed:', err)
        screencapMethods.value = []
        inputMethods.value = []
        gamepadTypes.value = []
    }
})


// --- Win独有配置 ---
const win32Config = reactive({
    mouse_method: DEFAULT_WIN32_MOUSE,
    keyboard_method: DEFAULT_WIN32_KEYBOARD,
})

watch(
    () => [controllerStore.win32MouseMethod, controllerStore.win32KeyboardMethod],
    ([mouse, keyboard]) => {
        win32Config.mouse_method = mouse ?? DEFAULT_WIN32_MOUSE
        win32Config.keyboard_method = keyboard ?? DEFAULT_WIN32_KEYBOARD
    },
    { immediate: true },
)

// Win32 config 变化时保存到 store
watch(
    () => [win32Config.mouse_method, win32Config.keyboard_method] as const,
    ([mouse, keyboard]) => {
        controllerStore.updateWin32Input({ mouse_method: mouse, keyboard_method: keyboard })
    },
)

// --- Gamepad 独有配置 ---
const gamepadConfig = reactive({
    gamepad_type: DEFAULT_GAMEPAD_TYPE,
    gamepad_icon: DEFAULT_GAMEPAD_ICON
})

watch(
    () => controllerStore.gamepadType,
    (gamepadType) => {
        gamepadConfig.gamepad_type = gamepadType ?? DEFAULT_GAMEPAD_TYPE
        switch (gamepadType) {
            case "0":
                gamepadConfig.gamepad_icon = "i-simple-icons:xbox"
                break
            case "1":
                gamepadConfig.gamepad_icon = "i-simple-icons:playstation"
                break
            default:
                gamepadConfig.gamepad_icon = ""
        }
    },
    { immediate: true },
)

// Gamepad config 变化时保存到 store
watch(
    () => gamepadConfig.gamepad_type,
    (gamepadType) => {
        controllerStore.updateGamepadInput({ gamepad_type: gamepadType })
    },
)

// --- 恢复持久化 / 导入的窗口选项 ---
watch(
    [
        () => controllerStore.desktopHwnd,
        () => controllerStore.desktopClassFilter,
        () => controllerStore.desktopWindowRegex,
        windowSearchRef,
    ],
    ([hwnd, classFilter, windowRegex, ref]) => {
        if (ref) {
            ref.searchFilter.className = classFilter ?? ''
            ref.searchFilter.windowRegex = windowRegex ?? ''
            if (hwnd) {
                ref.initFromPersisted(hwnd, controllerStore.desktopWindowName)
            }
        }
    },
    { immediate: true },
)

// 搜索过滤条件变化时保存到 store
watch(
    () => [windowSearchRef.value?.searchFilter.className, windowSearchRef.value?.searchFilter.windowRegex] as const,
    ([className, windowRegex]) => {
        if (className !== undefined) {
            controllerStore.desktopClassFilter = className ?? ''
        }
        if (windowRegex !== undefined) {
            controllerStore.desktopWindowRegex = windowRegex ?? ''
        }
    },
)

// --- 连接 / 断连 ---

async function onConnect() {
    const hwnd = windowSearchRef.value?.selectedHwnd
    if (!hwnd) return

    controllerStore.connecting = true
    try {
        const type = controllerValue.value as 'win32' | 'gamepad'
        const params: ConnectControllerRequest = type === 'win32'
            ? {
                type,
                hwnd,
                win32_screencap_method: desktopScreencap.value,
                win32_mouse_method: win32Config.mouse_method,
                win32_keyboard_method: win32Config.keyboard_method,
            }
            : {
                type,
                hwnd,
                gamepad_screencap_method: desktopScreencap.value,
                gamepad_type: gamepadConfig.gamepad_type,
            }

        const result = await connectController(params)
        if (!result.succeed) {
            console.error(`[${type}] Connect failed:`, result.msg)
            toast.add({
                id: 'ctrl-toast',
                title: 'Controller Connect Failed',
                description: result.msg || 'Unknown error',
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
            return
        }

        // 连接时不自动收起 方便用户切换连接方式
        // showFullCard.value = false
        toast.add({
            id: 'ctrl-toast',
            title: 'Controller Connected',
            icon: 'i-lucide-check-circle',
            color: 'success',
        })

        // 持久化共享桌面配置
        const windowInfo = windowSearchRef.value?.windowMap?.[hwnd]
        controllerStore.updateDesktopConfig({
            hwnd,
            class_name: windowInfo?.class_name ?? controllerStore.desktopClassName,
            window_name: windowInfo?.window_name ?? controllerStore.desktopWindowName,
            screencap_method: desktopScreencap.value,
        })
    } catch (err) {
        console.error('[Desktop] Connect failed:', err)
    } finally {
        controllerStore.connecting = false
    }
}

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
        console.error('[Desktop] Disconnect failed:', err)
    }
}
</script>
