<template>
    <div class="flex h-full flex-col gap-3 rounded-xl border border-default bg-default/40 p-3">
        <div class="flex items-center gap-2">
            <UIcon name="i-lucide-list-tree" class="size-4 text-dimmed" />
            <span class="text-sm font-semibold">{{ t('taskDetail.navigator') }}</span>
        </div>

        <UEmpty v-if="tasks.length === 0" icon="i-lucide-list-checks" :title="t('taskDetail.noTasks')"
            :description="t('taskDetail.noTasksDescription')" />

        <template v-else>
            <div class="flex flex-col gap-2">
                <span class="text-xs font-medium uppercase tracking-wide text-dimmed">{{ t('taskDetail.tasks') }}</span>
                <div class="flex flex-col gap-2">
                    <button v-for="(task, index) in tasks" :key="task.msg.uuid" type="button"
                        class="flex w-full items-center gap-2 rounded-lg border px-3 py-2 text-left transition-colors"
                        :class="index === activeIndex ? 'border-primary bg-primary/10 text-default' : 'border-default bg-default hover:bg-elevated'"
                        @click="$emit('selectTask', index)">
                        <UBadge color="neutral" variant="soft" size="xs" class="shrink-0">#{{ index + 1 }}</UBadge>
                        <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ task.msg.entry }}</span>
                        <UBadge
                            :color="task.status === 'success' ? 'success' : task.status === 'failed' ? 'error' : 'info'"
                            variant="subtle" size="xs" class="capitalize shrink-0">
                            {{ task.status }}
                        </UBadge>
                    </button>
                </div>
            </div>

            <div v-if="activeTask" class="flex min-h-0 flex-1 flex-col gap-2">
                <div class="flex items-center justify-between gap-2">
                    <span class="text-xs font-medium uppercase tracking-wide text-dimmed">{{ t('taskDetail.nodes') }}</span>
                    <UButton v-if="isHistoryMode" size="xs" color="primary" variant="soft"
                        icon="i-lucide-arrow-down-to-line" @click="$emit('goLatest')">
                        {{ t('taskDetail.latest') }}
                    </UButton>
                </div>

                <div class="min-h-0 flex-1 overflow-y-auto pr-1">
                    <div class="flex flex-col gap-2">
                        <button v-for="node in displayedNodes" :key="`${node.msg.node_id}`" type="button"
                            class="flex w-full items-start gap-2 rounded-lg border px-3 py-2 text-left transition-colors"
                            :class="node.msg.node_id === selectedNodeId ? 'border-primary bg-primary/10 text-default' : 'border-default bg-default hover:bg-elevated'"
                            @click="$emit('selectNode', node.msg.node_id)">
                            <UIcon :name="node.msg.node_id === entryNodeId ? 'i-lucide-flag' : 'i-lucide-workflow'"
                                class="mt-0.5 size-4 shrink-0 text-dimmed" />
                            <div class="min-w-0 flex-1">
                                <div class="truncate text-sm font-medium">{{ node.msg.name }}</div>
                                <div class="mt-1 flex items-center gap-2 text-xs text-dimmed">
                                    <span>#{{ node.msg.node_id }}</span>
                                    <span>•</span>
                                    <span class="capitalize">{{ node.status }}</span>
                                </div>
                            </div>
                        </button>
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { TaskScope, PipelineNodeScope } from '@/types/taskDetail'

const { t } = useI18n()

defineProps<{
    tasks: TaskScope[]
    activeTask: TaskScope | null
    activeIndex: number
    displayedNodes: PipelineNodeScope[]
    selectedNodeId: number | null
    entryNodeId: number | null
    isHistoryMode: boolean
}>()

defineEmits<{
    selectTask: [index: number]
    selectNode: [nodeId: number]
    goLatest: []
}>()
</script>
