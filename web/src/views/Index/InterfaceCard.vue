<template>
    <UCard class="w-full max-w-xl transition-opacity duration-200"
        :class="{ 'opacity-50 pointer-events-none': isTaskRunning }" size="xl"
        :ui="{ root: 'bg-default/70 ring-default/70', header: 'p-4 sm:px-5', body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">{{ t('interface.title') }}</span>
                        <UBadge :color="statusColor" variant="subtle" size="sm" class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="loading"
                                    class="absolute inline-flex size-full animate-ping rounded-full bg-warning opacity-75"></span>
                                <span class="relative inline-flex size-2 rounded-full" :class="dotClass"></span>
                            </span>
                            {{ statusLabel }}
                        </UBadge>

                        <span v-if="loadedInterface" class="font-medium text-muted">
                            {{ t('interface.project', { name: loadedInterface.name || t('interface.unnamed') }) }}
                        </span>
                    </div>
                </div>
            </div>
        </template>

        <div class="p-4 sm:p-6 min-h-36 flex flex-col gap-3">
            <UTooltip :text="interfacePath">
                <UFormField name="interfacePath" :label="t('interface.filePath')" :error="pathError || undefined">
                    <template #content>
                        <p>{{ t('interface.filePathHint') }}</p>
                    </template>
                    <UInput v-model="interfacePath" class="w-full" icon="i-lucide-file-json" size="xl"
                        :color="pathError ? 'error' : 'neutral'" @blur="onPathBlur" />
                </UFormField>
            </UTooltip>

            <UFormField v-if="taskStore.hasInterfaceLanguages" name="interfaceLanguage" :label="t('interface.language')">
                <USelect v-model="selectedInterfaceLanguage" :items="interfaceLanguageItems" value-key="value"
                    class="w-full" size="xl" arrow />
            </UFormField>

            <div class="flex justify-end">
                <UButton color="primary" variant="soft" icon="i-lucide-folder-open" size="xl" :loading="loading"
                    :disabled="!canLoad" :label="t('interface.load')" @click="onLoad" />
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { checkPathExists, getStoreConfig, parseInterface, saveStoreConfig } from '@/api/http'
import { useStatusStore } from '@/stores/status'
import { useControllerStore } from '@/stores/controller'
import { useResourceStore } from '@/stores/resource'
import { useTaskStore } from '@/stores/task'
import type { InterfaceControllerCandidate, InterfaceParseResult } from '@/types/interface'

const toast = useToast()
const { t } = useI18n()
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
const selectedInterfaceLanguage = computed({
    get: () => taskStore.selectedInterfaceLanguage,
    set: (value: string) => {
        taskStore.setInterfaceLanguage(value)
    },
})
const interfaceLanguageItems = computed(() =>
    taskStore.availableInterfaceLanguages.map((item) => ({
        label: item.label,
        value: item.value,
    })),
)
const statusLabel = computed(() => {
    if (loading.value) return t('interface.loading')
    if (loadedInterface.value) return t('interface.loaded')
    return t('interface.idle')
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
        pathError.value = result.msg || t('interface.pathValidationFailed')
        return false
    }

    pathError.value = ''
    return true
}

onMounted(async () => {
    const saved = await getStoreConfig<{ path?: string }>('interface')
    if (saved?.path) {
        interfacePath.value = saved.path
        // 重新解析已保存的 interface 文件，恢复 Project 名称等展示信息。
        // 只刷新展示，不重复应用 controller/resource/task（这些已由持久化 store 恢复）。
        try {
            const result = await parseInterface(saved.path)
            if (result.succeed && result.data) {
                loadedInterface.value = result.data
            }
        } catch {
            // 解析失败（文件被移动/删除等）时保持 Idle 展示
        }
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
            title: t('interface.invalidPath'),
            description: t('interface.invalidPathDescription'),
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
                title: t('interface.loadFailed'),
                description: result.msg || t('interface.parseFailed'),
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

        taskStore.applyInterfaceTasks(parsed.task_candidates, {
            languages: parsed.languages,
            localeValues: parsed.locale_values,
        })

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
            title: t('interface.loadSuccess'),
            description: [
                controllerApplied && controller
                    ? t('interface.controllerApplied', { name: controller.name || controller.type })
                    : t('interface.controllerSkipped'),
                t('interface.resourceProfile', { name: resourceProfile.profileName }),
                t('interface.resourcePathsPatched', { count: resourceProfile.paths.length }),
                t('interface.importFiles', { count: parsed.imports?.length ?? 0 }),
                parsed.task_candidates[0]?.entry
                    ? t('interface.firstTaskEntry', { entry: parsed.task_candidates[0].entry })
                    : t('interface.firstTaskEntryNa'),
                t('interface.resolvedResourcePaths', { count: resourceCount }),
            ].filter(Boolean).join('\n'),
            icon: 'i-lucide-check-circle',
            color: 'success',
        })
    } finally {
        loading.value = false
    }
}
</script>
