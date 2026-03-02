<template>
    <UCard class="w-full max-w-xl" size="xl" :ui="{ body: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Interface</span>
                    </div>
                    <UButton variant="outline" color="neutral" trailing-icon="i-lucide-chevron-down"
                        :data-state="showFullCard ? 'open' : 'closed'" @click="showFullCard = !showFullCard" />
                </div>
                <div class="grid transition-all duration-200 ease-out"
                    :class="showFullCard ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100'">
                    <div class="overflow-hidden">
                        <div class="text-sm text-dimmed truncate">
                            {{ summaryText }}
                        </div>
                    </div>
                </div>
            </div>
        </template>

        <template #default>
            <UCollapsible v-model:open="showFullCard" :unmount-on-hide="false">
                <template #content>
                    <div class="p-4 sm:p-6 min-h-36 flex flex-col gap-3">
                        <UFormField name="interfacePath" label="File Path" :error="pathError || undefined">
                            <UInput v-model="interfacePath" class="w-full"
                                placeholder="Enter interface.json file path..." icon="i-lucide-file-json" size="xl"
                                :color="pathError ? 'error' : 'neutral'" @blur="onPathBlur" />
                        </UFormField>

                        <div class="flex flex-row items-center justify-end">
                            <UButton color="primary" variant="soft" icon="i-lucide-folder-open" size="xl"
                                :loading="loading" :disabled="!canLoad" @click="onLoad">
                                Load
                            </UButton>
                        </div>
                    </div>
                </template>
            </UCollapsible>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { checkPathExists } from '@/api/http'
import { useStatusStore } from '@/stores/status'

const toast = useToast()
const statusStore = useStatusStore()

const showFullCard = ref(true)
const interfacePath = ref('')
const loading = ref(false)
const pathError = ref('')

const canLoad = computed(() => interfacePath.value.trim().length > 0 && !loading.value && !pathError.value)
const summaryText = computed(() => interfacePath.value.trim() || 'No interface loaded')

// 任务开始运行时自动收起卡片
watch(() => statusStore.taskStatus, (newStatus, oldStatus) => {
    if (oldStatus !== 'running' && newStatus === 'running') {
        showFullCard.value = false
    }
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

async function onPathBlur() {
    await validatePath()
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
        toast.add({
            id: 'interface-todo-toast',
            title: 'Interface load is not available yet',
            description: 'Backend support is not implemented. UI is ready and waiting for API integration.',
            icon: 'i-lucide-info',
            color: 'info',
        })
    } finally {
        loading.value = false
    }
}
</script>
