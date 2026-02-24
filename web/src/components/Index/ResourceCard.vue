<template>
    <UCard class="w-full" size="xl" :ui="{ body: 'p-0 sm:p-0', footer: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <span class="font-bold">Resource</span>
                    <div class="flex flex-row items-center gap-2">
                        <USelect v-model="resourceStore.activeProfileId" :items="resourceStore.profileSelectItems"
                            class="w-32" size="xl" arrow />
                        <UDropdownMenu :items="profileMenuItems">
                            <UButton color="neutral" variant="ghost" icon="i-lucide-ellipsis-vertical" size="xs" />
                        </UDropdownMenu>
                        <UButton variant="outline" color="neutral" trailing-icon="i-lucide-chevron-down"
                            :data-state="showFullCard ? 'open' : 'closed'" @click="showFullCard = !showFullCard" />
                    </div>
                </div>
                <div v-show="!showFullCard" class="text-sm text-dimmed">
                    {{ resourceStore.activeProfile.name }} · {{ resourceStore.activePaths.length }} paths
                </div>
            </div>
        </template>

        <template #default>
            <div v-show="showFullCard" class="p-4 sm:p-6">
                <div class="resource-list flex flex-col gap-2 min-h-12"
                    :class="resourceStore.activePaths.length > 3 ? 'max-h-48 overflow-y-auto pr-2' : ''">
                    <div v-if="resourceStore.activePaths.length === 0"
                        class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-2 text-dimmed gap-2">
                        <UIcon name="i-lucide-folder-open" class="size-5" />
                        <span class="text-sm">No resource paths added</span>
                    </div>

                    <div v-for="(item, index) in resourceStore.activePaths" :key="item.id"
                        class="group flex flex-row items-center gap-2 rounded-lg border border-default p-2 transition-colors hover:bg-elevated"
                        :class="{ 'opacity-50': !item.enabled }" :draggable="resourceStore.activePaths.length > 1"
                        @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @dragend="onDragEnd">

                        <!-- Enable/Disable Checkbox -->
                        <UCheckbox v-model="item.enabled" />

                        <!-- Drag Handle -->
                        <div v-if="resourceStore.activePaths.length > 1"
                            class="cursor-grab text-dimmed hover:text-default active:cursor-grabbing">
                            <UIcon name="i-lucide-grip-vertical" class="size-5" />
                        </div>

                        <!-- Path Input -->
                        <UInput v-if="editingIndex === index" v-model="item.path" placeholder="/path/to/resource"
                            class="flex-1" size="xl" autofocus @keydown.enter="onFinishEdit(index)"
                            @blur="onFinishEdit(index)" />

                        <!-- Path Display -->
                        <UTooltip v-else :text="item.path" :disabled="!item.path">
                            <div class="flex-1 flex items-center gap-2 min-w-0 cursor-pointer" @click="onEdit(index)">
                                <span class="truncate text-xl" :class="item.path ? '' : 'text-dimmed italic'">
                                    {{ item.path || 'Click to edit path...' }}
                                </span>
                            </div>
                        </UTooltip>

                        <!-- Action Buttons -->
                        <div class="flex flex-row gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                            <UTooltip text="Edit">
                                <UButton color="neutral" variant="ghost" icon="i-lucide-square-pen" size="xs"
                                    @click="onEdit(index)" />
                            </UTooltip>
                            <UTooltip text="Remove">
                                <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                                    @click="resourceStore.removePath(index)" />
                            </UTooltip>
                        </div>
                    </div>
                </div>
            </div>
        </template>

        <template #footer>
            <div v-show="showFullCard" class="p-2 sm:p-4">
                <UButton color="neutral" variant="ghost" icon="i-lucide-plus" label="Add path" block
                    @click="onAddPath" />
            </div>
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
import { ref } from 'vue'
import { useResourceStore } from '@/stores/resource'

const resourceStore = useResourceStore()

// --- UI State (not persisted) ---
const showFullCard = ref(true)
const editingIndex = ref<number | null>(null)

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

function onFinishEdit(index: number) {
    editingIndex.value = null
    // Remove empty paths automatically
    const item = resourceStore.activePaths[index]
    if (item && !item.path.trim()) {
        resourceStore.removePath(index)
    }
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
