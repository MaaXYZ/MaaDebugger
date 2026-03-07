<template>
    <UModal v-model:open="open" :title="modalTitle" :ui="{ content: 'sm:max-w-[90vw] sm:w-[90vw] max-h-[90vh]' }">
        <template #body>
            <div class="flex flex-col gap-3 min-h-105">
                <div class="flex items-center justify-between gap-2 flex-wrap">
                    <div class="flex items-center gap-2 min-w-0">
                        <UBadge color="neutral" variant="subtle">Runtime</UBadge>
                        <span class="text-sm text-dimmed truncate">{{ nodeName || 'Unknown node' }}</span>
                    </div>
                </div>

                <div v-if="loading"
                    class="flex flex-1 items-center justify-center rounded-lg border border-default bg-muted/30">
                    <UIcon name="i-lucide-loader" class="size-6 animate-spin text-dimmed" />
                </div>

                <UAlert v-else-if="errorMessage" color="error" variant="soft" icon="i-lucide-circle-alert"
                    :title="errorMessage" />

                <MonacoEditor v-else :model-value="editorValue" language="json" :read-only="true" :min-height="420"
                    :max-height="720" />
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getNodeData } from '@/api/http'
import { MonacoEditor } from '@/components/MonacoEditor'

const props = defineProps<{
    nodeName: string | null
}>()

const open = defineModel<boolean>('open', { default: false })

const loading = ref(false)
const errorMessage = ref('')
const editorValue = ref('{}')

const modalTitle = "Node Data"

function formatNodeJson(value: string): string {
    if (!value.trim()) return '{}'

    try {
        return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
        return value
    }
}

async function loadNodeData() {
    if (!open.value || !props.nodeName) {
        editorValue.value = '{}'
        errorMessage.value = ''
        return
    }

    loading.value = true
    errorMessage.value = ''

    try {
        const detail = await getNodeData(props.nodeName)
        if (!detail?.node_json) {
            editorValue.value = '{}'
            errorMessage.value = '未获取到节点原始定义'
            return
        }

        editorValue.value = formatNodeJson(detail.node_json)
    } catch (error) {
        editorValue.value = '{}'
        errorMessage.value = error instanceof Error ? error.message : '获取节点原始定义失败'
    } finally {
        loading.value = false
    }
}

watch(
    [open, () => props.nodeName],
    async ([isOpen, nodeName]) => {
        if (!isOpen || !nodeName) {
            loading.value = false
            errorMessage.value = ''
            editorValue.value = '{}'
            return
        }

        await loadNodeData()
    },
    { immediate: true },
)
</script>
