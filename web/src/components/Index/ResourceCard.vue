<template>
    <UCard class="w-full transition-opacity duration-200" :class="{ 'opacity-50 pointer-events-none': isCardDisabled }"
           size="xl" :ui="{ body: 'p-0 sm:p-0', footer: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Resource</span>
                        <UBadge :color="statusColor" variant="subtle" size="sm" class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="statusStore.resourceStatus === 'loading'"
                                      class="absolute inline-flex size-full animate-ping rounded-full bg-warning opacity-75"></span>
                                <span class="relative inline-flex size-2 rounded-full" :class="dotClass"></span>
                            </span>
                            {{ statusLabel }}
                        </UBadge>
                    </div>
                    <div class="flex flex-row items-center gap-2">
                        <USelect v-model="resourceStore.activeProfileId" :items="resourceStore.profileSelectItems"
                                 class="w-32" size="xl" arrow :disabled="isCardDisabled" />
                        <UDropdownMenu :items="profileMenuItems">
                            <UButton color="neutral" variant="ghost" icon="i-lucide-ellipsis-vertical" size="xs"
                                     :disabled="isCardDisabled" />
                        </UDropdownMenu>
                        <UButton variant="outline" color="neutral" trailing-icon="i-lucide-chevron-down"
                                 :data-state="showFullCard ? 'open' : 'closed'" @click="showFullCard = !showFullCard" />
                    </div>
                </div>
                <div class="grid transition-all duration-200 ease-out"
                     :class="showFullCard ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100'">
                    <div class="overflow-hidden">
                        <div class="text-sm text-dimmed truncate">
                            {{ resourceStore.activeProfile.name }} · {{ resourceStore.activePaths.length }} paths
                        </div>
                    </div>
                </div>
            </div>
        </template>

        <template #default>
            <UCollapsible v-model:open="showFullCard" :unmount-on-hide="false">
                <template #content>
                    <div class="p-4 sm:p-6">
                        <div class="resource-list flex flex-col gap-2 min-h-12"
                             :class="resourceStore.activePaths.length > 3 ? 'max-h-48 overflow-y-auto pr-2' : ''">
                            <div v-if="resourceStore.activePaths.length === 0"
                                 class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-2 text-dimmed gap-2">
                                <UIcon name="i-lucide-folder-open" class="size-5" />
                                <span class="text-sm">No resource paths added</span>
                            </div>

                            <div v-for="(item, index) in resourceStore.activePaths" :key="item.id"
                                 class="group flex flex-col gap-1 rounded-lg border border-default p-2 transition-colors hover:bg-elevated"
                                 :class="{ 'opacity-50': !item.enabled }"
                                 :draggable="resourceStore.activePaths.length > 1" @dragstart="onDragStart(index)"
                                 @dragover.prevent="onDragOver(index)" @dragend="onDragEnd">

                                <div class="flex flex-row items-center gap-2">
                                    <!-- Enable/Disable Checkbox -->
                                    <UCheckbox v-model="item.enabled" />

                                    <!-- Drag Handle -->
                                    <div v-if="resourceStore.activePaths.length > 1"
                                         class="cursor-grab text-dimmed hover:text-default active:cursor-grabbing">
                                        <UIcon name="i-lucide-grip-vertical" class="size-5" />
                                    </div>

                                    <!-- Path Input -->
                                    <UInput v-if="editingIndex === index" v-model="item.path"
                                            placeholder="/path/to/resource" class="flex-1" size="md" autofocus
                                            :color="pathErrors[item.id] ? 'error' : 'neutral'"
                                            @keydown.enter="onFinishEdit(index)" @blur="onFinishEdit(index)" />

                                    <!-- Path Display -->
                                    <UTooltip v-else :text="item.path" :disabled="!item.path">
                                        <div class="flex-1 flex items-center gap-2 min-w-0 cursor-pointer"
                                             @click="onEdit(index)">
                                            <span class="truncate text-md"
                                                  :class="item.path ? '' : 'text-dimmed italic'">
                                                {{ item.path || 'Click to edit path...' }}
                                            </span>
                                        </div>
                                    </UTooltip>

                                    <!-- Action Buttons -->
                                    <div
                                        class="flex flex-row gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <UTooltip text="Edit">
                                            <UButton color="neutral" variant="ghost" icon="i-lucide-square-pen"
                                                     size="xs" @click="onEdit(index)" />
                                        </UTooltip>
                                        <UTooltip text="Remove">
                                            <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                                                     @click="onRemovePath(index, item.id)" />
                                        </UTooltip>
                                    </div>
                                </div>

                                <div v-if="pathErrors[item.id]" class="pl-6 text-xs text-error truncate">
                                    {{ pathErrors[item.id] }}
                                </div>
                            </div>
                        </div>

                        <div class="p-2 sm:p-4">
                            <UButton color="neutral" variant="ghost" icon="i-lucide-plus" label="Add path" block
                                     @click="onAddPath" />
                        </div>

                        <!-- Load Button -->
                        <div class="px-2 sm:px-4 pb-2 sm:pb-4">
                            <UButton color="primary" icon="i-lucide-download" label="Load Resource" block size="xl"
                                     :loading="isLoading" :disabled="enabledPaths.length === 0 || isLoading"
                                     @click="onLoadResource" />
                        </div>
                    </div>
                </template>
            </UCollapsible>
        </template>
    </UCard>

    <!-- Rename Profile Modal -->
    <UModal v-model:open="renameModalOpen" title="Rename Profile" description="Enter a new name for this profile.">
        <template #body>
            <UInput v-model="renameInput" placeholder="Profile name..." size="xl" autofocus
                    @keydown.enter="onConfirmRename" />
        </template>
        <template #footer>
            <div class="flex w-full justify-end gap-2">
                <UButton color="neutral" variant="ghost" label="Cancel" @click="renameModalOpen = false" />
                <UButton color="primary" label="Rename" :disabled="!renameInput.trim()" @click="onConfirmRename" />
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { checkPathExists } from '@/api/http'
import { useResourceStore } from '@/stores/resource'
import { useStatusStore } from '@/stores/status'
import useResourceControl from './useResourceControl'

const toast = useToast()
const resourceStore = useResourceStore()
const statusStore = useStatusStore()
const { enabledPaths, onLoadResource: triggerLoadResource } = useResourceControl()

// --- UI State (not persisted) ---
const showFullCard = ref(true)
const editingIndex = ref<number | null>(null)
const pathErrors = ref<Record<number, string>>({})

// 任务开始运行时自动收起卡片
watch(() => statusStore.taskStatus, (newStatus, oldStatus) => {
    if (oldStatus !== 'running' && newStatus === 'running') {
        showFullCard.value = false
    }
})

// --- Resource Status ---
const isLoading = computed(() => statusStore.resourceStatus === 'loading')
const isConnecting = computed(() => statusStore.controllerStatus === 'connecting')

const isTaskRunning = computed(() => statusStore.taskStatus === 'running')

/**
 * 当 controller 正在连接、资源正在加载或任务正在运行时，禁用整个 ResourceCard 的交互
 */
const isCardDisabled = computed(() => isConnecting.value || isLoading.value || isTaskRunning.value)

const statusLabel = computed(() => {
    switch (statusStore.resourceStatus) {
    case 'loaded':
        return 'Loaded'
    case 'loading':
        return 'Loading'
    case 'failed':
        return 'Failed'
    case 'unloaded':
    default:
        return 'Idle'
    }
})

const statusColor = computed(() => {
    switch (statusStore.resourceStatus) {
    case 'loaded':
        return 'success' as const
    case 'loading':
        return 'warning' as const
    case 'failed':
        return 'error' as const
    case 'unloaded':
    default:
        return 'neutral' as const
    }
})

const dotClass = computed(() => {
    switch (statusStore.resourceStatus) {
    case 'loaded':
        return 'bg-success'
    case 'loading':
        return 'bg-warning'
    case 'failed':
        return 'bg-error'
    case 'unloaded':
    default:
        return 'bg-gray-400 dark:bg-gray-500'
    }
})

// --- Profile Menu ---
const profileMenuItems = [
    [
        { label: 'New profile', icon: 'i-lucide-plus', onSelect: () => resourceStore.addProfile() },
        { label: 'Rename profile', icon: 'i-lucide-pencil', onSelect: onRenameProfile },
    ],
    [
        {
            label: 'Delete profile',
            icon: 'i-lucide-trash-2',
            color: 'error' as const,
            disabled: resourceStore.profiles.length <= 1,
            onSelect: () => resourceStore.deleteProfile(),
        },
    ],
]

// --- Drag & Drop ---
const dragIndex = ref<number | null>(null)

function onDragStart(index: number) {
    dragIndex.value = index
}

function onDragOver(index: number) {
    if (dragIndex.value === null || dragIndex.value === index) return
    resourceStore.reorderPaths(dragIndex.value, index)
    dragIndex.value = index
}

function onDragEnd() {
    dragIndex.value = null
}

// --- Path Actions ---
function onAddPath() {
    resourceStore.addPath()
    editingIndex.value = resourceStore.activePaths.length - 1
}

function onEdit(index: number) {
    editingIndex.value = index
}

function clearPathError(pathId: number) {
    if (!pathErrors.value[pathId]) return
    const { [pathId]: _removed, ...rest } = pathErrors.value
    pathErrors.value = rest
}

async function validatePath(path: string, pathId: number): Promise<boolean> {
    const trimmed = path.trim()
    if (!trimmed) {
        clearPathError(pathId)
        return false
    }

    const result = await checkPathExists(trimmed, 'dir')
    const exists = Boolean(result.succeed && result.data?.exists)

    if (!exists) {
        pathErrors.value = {
            ...pathErrors.value,
            [pathId]: result.succeed ? 'File does not exist!' : (result.msg || 'Path validation failed'),
        }
        return false
    }

    clearPathError(pathId)
    return true
}

async function onFinishEdit(index: number) {
    editingIndex.value = null

    // Remove empty paths automatically
    const item = resourceStore.activePaths[index]
    if (item && !item.path.trim()) {
        clearPathError(item.id)
        resourceStore.removePath(index)
        return
    }

    if (item) {
        await validatePath(item.path, item.id)
    }
}

function onRemovePath(index: number, pathId: number) {
    clearPathError(pathId)
    resourceStore.removePath(index)
}

async function onLoadResource() {
    const validationList = await Promise.all(
        resourceStore.activePaths
            .filter((item) => item.enabled)
            .map(async (item) => ({
                id: item.id,
                ok: await validatePath(item.path, item.id),
            })),
    )

    const hasInvalidPath = validationList.some((item) => !item.ok)
    if (hasInvalidPath) {
        toast.add({
            id: 'res-path-toast',
            title: 'Invalid Resource Path',
            description: 'Please fix invalid paths before loading resource!',
            icon: 'i-lucide-circle-x',
            color: 'error',
        })
        return
    }

    await triggerLoadResource()
}

// --- Rename Modal ---
const renameModalOpen = ref(false)
const renameInput = ref('')

function onRenameProfile() {
    renameInput.value = resourceStore.activeProfile.name
    renameModalOpen.value = true
}

function onConfirmRename() {
    resourceStore.renameProfile(renameInput.value)
    renameModalOpen.value = false
}
</script>

<style scoped>
.resource-list::-webkit-scrollbar {
    width: 4px;
}

.resource-list::-webkit-scrollbar-track {
    background: transparent;
}

.resource-list::-webkit-scrollbar-thumb {
    background-color: rgba(128, 128, 128, 0.3);
    border-radius: 2px;
}

.resource-list::-webkit-scrollbar-thumb:hover {
    background-color: rgba(128, 128, 128, 0.5);
}

.resource-list {
    scrollbar-width: thin;
    scrollbar-color: rgba(128, 128, 128, 0.3) transparent;
}
</style>
