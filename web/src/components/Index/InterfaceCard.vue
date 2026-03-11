<template>
    <UCard class="w-full max-w-xl transition-opacity duration-200"
        :class="{ 'opacity-50 pointer-events-none': isTaskRunning }" size="xl" :ui="{ body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Interface</span>
                        <UBadge :color="statusColor" variant="subtle" size="sm" class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="loading"
                                    class="absolute inline-flex size-full animate-ping rounded-full bg-warning opacity-75"></span>
                                <span class="relative inline-flex size-2 rounded-full" :class="dotClass"></span>
                            </span>
                            {{ statusLabel }}
                        </UBadge>

                        <span class="font-medium text-muted" v-if="loadedInterface">
                            Project: {{ loadedInterface.name || 'Unnamed interface' }}
                        </span>
                    </div>
                </div>
            </div>
        </template>

        <div class="p-4 sm:p-6 min-h-36 flex flex-col gap-3">
            <UFormField name="interfacePath" label="File Path" :error="pathError || undefined">
                <UInput v-model="interfacePath" class="w-full" placeholder="Enter interface.json file path..."
                    icon="i-lucide-file-json" size="xl" :color="pathError ? 'error' : 'neutral'" @blur="onPathBlur" />
            </UFormField>

            <div class="flex justify-end">
                <UButton color="primary" variant="soft" icon="i-lucide-folder-open" size="xl" :loading="loading"
                    :disabled="!canLoad" @click="onLoad" label="Load" />
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { checkPathExists, getStoreConfig, parseInterface, saveStoreConfig } from '@/api/http'
import { useStatusStore } from '@/stores/status'
import { useControllerStore } from '@/stores/controller'
import { useResourceStore } from '@/stores/resource'
import { useTaskStore } from '@/stores/task'
import type { InterfaceControllerCandidate, InterfaceParseResult } from '@/types/interface'

const toast = useToast()
const statusStore = useStatusStore()
const controllerStore = useControllerStore()
const resourceStore = useResourceStore()
const taskStore = useTaskStore()

const isTaskRunning = computed(() => statusStore.taskStatus === 'running')
const interfacePath = ref('')
const loading = ref(false)
const pathError = ref('')

const canLoad = computed(() => interfacePath.value.trim().length > 0 && !loading.value && !pathError.value)
const loadedInterface = ref<InterfaceParseResult | null>(null)
const statusLabel = computed(() => {
    if (loading.value) return 'Loading'
    if (loadedInterface.value) return 'Loaded'
    return 'Idle'
})
const statusColor = computed(() => {
    if (loading.value) return 'warning' as const
    if (loadedInterface.value) return 'success' as const
    return 'neutral' as const
})
const dotClass = computed(() => {
    if (loading.value) return 'bg-warning'
    if (loadedInterface.value) return 'bg-success'
    return 'bg-gray-400 dark:bg-gray-500'
})

watch(interfacePath, () => {
    if (pathError.value) {
        pathError.value = ''
    }
})

async function validatePath(): Promise<boolean> {
    const trimmed = interfacePath.value.trim()
    if (!trimmed) {
        pathError.value = ''
        return false
    }

    const result = await checkPathExists(trimmed, 'file')
    const exists = Boolean(result.succeed && result.data?.exists)

    if (!exists) {
        pathError.value = result.succeed ? 'File does not exist' : (result.msg || 'Path validation failed')
        return false
    }

    pathError.value = ''
    return true
}

onMounted(async () => {
    const saved = await getStoreConfig<{ path?: string }>('interface')
    if (saved?.path) {
        interfacePath.value = saved.path
    }
})

async function persistInterfacePath() {
    const path = interfacePath.value.trim()
    await saveStoreConfig('interface', { path })
}

async function onPathBlur() {
    const valid = await validatePath()
    if (valid) {
        await persistInterfacePath()
    }
}

function pickController(candidates: InterfaceControllerCandidate[]): InterfaceControllerCandidate | null {
    for (const candidate of candidates) {
        const normalizedType = candidate.type.trim().toLowerCase()
        if (normalizedType === 'win32' || normalizedType === 'adb' || normalizedType === 'playcover') {
            return candidate
        }
    }
    return candidates[0] ?? null
}

function buildInterfaceResourceProfile(parsed: InterfaceParseResult, controller: InterfaceControllerCandidate | null) {
    const primaryResource = parsed.resource_candidates[0]
    const profileName = primaryResource?.name?.trim() || parsed.name || 'Interface Resource'
    const resourcePaths = parsed.resource_candidates.flatMap((resource) =>
        resource.resolved_paths.map((item) => item.path)
    )
    const attachPaths = controller?.attach_resource_paths ?? []
    return {
        profileName,
        paths: [...resourcePaths, ...attachPaths],
    }
}

async function onLoad() {
    if (!canLoad.value) return

    const valid = await validatePath()
    if (!valid) {
        toast.add({
            id: 'interface-path-toast',
            title: 'Invalid interface path',
            description: 'Please provide an existing file path',
            icon: 'i-lucide-circle-x',
            color: 'error',
        })
        return
    }

    loading.value = true
    try {
        const result = await parseInterface(interfacePath.value.trim())
        if (!result.succeed || !result.data) {
            toast.add({
                id: 'interface-load-failed-toast',
                title: 'Interface load failed',
                description: result.msg || 'Failed to parse interface file',
                icon: 'i-lucide-circle-x',
                color: 'error',
            })
            return
        }

        const parsed = result.data
        loadedInterface.value = parsed

        const controller = pickController(parsed.controller_candidates)
        const controllerApplied = controller ? controllerStore.applyInterfaceController(controller) : false
        const resourceProfile = buildInterfaceResourceProfile(parsed, controller)
        console.log('[Interface] parsed result:', parsed)
        console.log('[Interface] selected controller:', controller)
        console.log('[Interface] merged resource paths:', resourceProfile.paths)
        if (resourceProfile.paths.length > 0) {
            resourceStore.applyInterfaceResourceProfile(resourceProfile.profileName, resourceProfile.paths)
            console.log('[Interface] resource store after patch:', resourceStore.activePaths)
        }

        taskStore.applyInterfaceTasks(parsed.task_candidates)

        await Promise.all([
            saveStoreConfig('controller', JSON.parse(JSON.stringify(controllerStore.$state))),
            saveStoreConfig('resource', JSON.parse(JSON.stringify(resourceStore.$state))),
            saveStoreConfig('task', JSON.parse(JSON.stringify(taskStore.$state))),
            persistInterfacePath(),
        ])

        const resourceCount = parsed.resource_candidates.reduce(
            (count, resource) => count + resource.resolved_paths.length,
            0,
        )
        toast.add({
            id: 'interface-load-success-toast',
            title: 'Interface loaded',
            description: [
                controllerApplied && controller
                    ? `Controller: ${controller.name || controller.type}`
                    : 'Controller: skipped',
                `Resource profile: ${resourceProfile.profileName}`,
                `Resource paths patched: ${resourceProfile.paths.length}`,
                `Import files: ${parsed.imports?.length ?? 0}`,
                parsed.task_candidates[0]?.entry
                    ? `First task entry: ${parsed.task_candidates[0].entry}`
                    : 'First task entry: n/a',
                resourceCount > 0 ? `Resolved resource paths: ${resourceCount}` : 'Resolved resource paths: 0',
            ].filter(Boolean).join('\n'),
            icon: 'i-lucide-check-circle',
            color: 'success',
        })
    } finally {
        loading.value = false
    }
}
</script>
