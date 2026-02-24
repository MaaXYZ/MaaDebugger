<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Resource</span>
                <div class="flex-1" />
                <USelect v-model="activeProfileId" :items="profileSelectItems" class="w-32" size="xl" />
                <UDropdownMenu :items="profileMenuItems">
                    <UButton color="neutral" variant="ghost" icon="i-lucide-ellipsis-vertical" size="xs" />
                </UDropdownMenu>
            </div>
        </template>

        <template #default>
            <div class="resource-list flex flex-col gap-2"
                :class="activePaths.length >= 4 ? 'max-h-49 overflow-y-auto pr-2' : ''">
                <div v-if="activePaths.length === 0"
                    class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-2 text-dimmed gap-2">
                    <UIcon name="i-lucide-folder-open" class="size-5" />
                    <span class="text-sm">No resource paths added</span>
                </div>

                <div v-for="(item, index) in activePaths" :key="item.id"
                    class="group flex flex-row items-center gap-2 rounded-lg border border-default p-2 transition-colors hover:bg-elevated"
                    :class="{ 'opacity-50': !item.enabled }" :draggable="activePaths.length > 1"
                    @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @dragend="onDragEnd">

                    <!-- Enable/Disable Checkbox -->
                    <UCheckbox v-model="item.enabled" />

                    <!-- Drag Handle -->
                    <div v-if="activePaths.length > 1"
                        class="cursor-grab text-dimmed hover:text-default active:cursor-grabbing">
                        <UIcon name="i-lucide-grip-vertical" class="size-5" />
                    </div>

                    <!-- Path Input -->
                    <UInput v-if="item.editing" v-model="item.path" placeholder="/path/to/resource" class="flex-1"
                        size="xl" autofocus @keydown.enter="onFinishEdit(index)" @blur="onFinishEdit(index)" />

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
                                @click="onRemove(index)" />
                        </UTooltip>
                    </div>
                </div>
            </div>
        </template>

        <template #footer>
            <UButton color="neutral" variant="ghost" icon="i-lucide-plus" label="Add path" block @click="onAddPath" />
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
import { ref, computed } from 'vue'

// --- Types ---
interface PathItem {
    id: number
    path: string
    enabled: boolean
    editing: boolean
}

interface Profile {
    id: number
    name: string
    paths: PathItem[]
}

// --- State ---
let nextId = 0
let nextProfileId = 0

const profiles = ref<Profile[]>([
    { id: nextProfileId++, name: 'Default', paths: [] }
])

const activeProfileId = ref<number>(profiles.value[0]!.id)

// --- Computed ---
const activeProfile = computed(() => {
    return profiles.value.find(p => p.id === activeProfileId.value) ?? profiles.value[0]!
})

const activePaths = computed({
    get: () => activeProfile.value.paths,
    set: (val) => { activeProfile.value.paths = val }
})

const profileSelectItems = computed(() => {
    return profiles.value.map(p => ({
        label: p.name,
        value: p.id
    }))
})

const profileMenuItems = computed(() => [
    [
        { label: 'New profile', icon: 'i-lucide-plus', onSelect: onAddProfile },
        { label: 'Rename profile', icon: 'i-lucide-pencil', onSelect: onRenameProfile },
    ],
    [
        {
            label: 'Delete profile',
            icon: 'i-lucide-trash-2',
            color: 'error' as const,
            disabled: profiles.value.length <= 1,
            onSelect: onDeleteProfile,
        },
    ],
])

// --- Drag & Drop ---
const dragIndex = ref<number | null>(null)

function onDragStart(index: number) {
    dragIndex.value = index
}

function onDragOver(index: number) {
    if (dragIndex.value === null || dragIndex.value === index) return
    const items = [...activePaths.value]
    const dragged = items.splice(dragIndex.value, 1)[0]
    if (!dragged) return
    items.splice(index, 0, dragged)
    activePaths.value = items
    dragIndex.value = index
}

function onDragEnd() {
    dragIndex.value = null
}

// --- Path Actions ---
function onAddPath() {
    activePaths.value.push({
        id: nextId++,
        path: '',
        enabled: true,
        editing: true,
    })
}

function onEdit(index: number) {
    const item = activePaths.value[index]
    if (item) {
        item.editing = true
    }
}

function onFinishEdit(index: number) {
    const item = activePaths.value[index]
    if (!item) return
    item.editing = false
    // Remove empty paths automatically to keep list clean
    if (!item.path.trim()) {
        activePaths.value.splice(index, 1)
    }
}

function onRemove(index: number) {
    activePaths.value.splice(index, 1)
}

// --- Profile Actions ---
function onAddProfile() {
    const newProfile: Profile = {
        id: nextProfileId++,
        name: `Profile ${profiles.value.length + 1}`,
        paths: []
    }
    profiles.value.push(newProfile)
    activeProfileId.value = newProfile.id
}

function onDeleteProfile() {
    if (profiles.value.length <= 1) return
    const idx = profiles.value.findIndex(p => p.id === activeProfileId.value)
    if (idx === -1) return
    profiles.value.splice(idx, 1)
    activeProfileId.value = profiles.value[0]!.id
}

function onRenameProfile() {
    renameInput.value = activeProfile.value.name
    renameModalOpen.value = true
}

// --- Rename Modal ---
const renameModalOpen = ref(false)
const renameInput = ref('')

function onConfirmRename() {
    if (renameInput.value.trim()) {
        activeProfile.value.name = renameInput.value.trim()
    }
    renameModalOpen.value = false
}

// --- Public API ---

/**
 * Get the list of enabled resource paths for the active profile.
 */
function getPaths(): string[] {
    return activePaths.value
        .filter(item => item.enabled && item.path)
        .map(item => item.path)
}

/**
 * Set the list of resource paths for the active profile from backend data.
 */
function setPaths(newPaths: string[]) {
    activePaths.value = newPaths.map((p) => ({
        id: nextId++,
        path: p,
        enabled: true,
        editing: false,
    }))
}

// Expose for parent component
defineExpose({
    paths: activePaths,
    profiles,
    activeProfileId,
    getPaths,
    setPaths,
})
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
