<template>
    <UCard class="w-full transition-opacity duration-200" :class="{ 'opacity-50 pointer-events-none': isTaskRunning }"
        size="xl"
        :ui="{ root: 'bg-default/70 ring-default/70', header: 'p-4 sm:px-5', body: 'p-0 sm:p-0', footer: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">{{ t('agent.title') }}</span>
                        <UBadge v-if="store.agents.length > 0" :color="headerBadgeColor" variant="subtle" size="sm"
                            class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="store.hasConnecting"
                                    class="absolute inline-flex size-full animate-ping rounded-full bg-info opacity-75"></span>
                                <span class="relative inline-flex size-2 rounded-full" :class="headerDotClass"></span>
                            </span>
                            {{ headerBadgeLabel }}
                        </UBadge>
                    </div>
                </div>
            </div>
        </template>
        <div class="p-4 sm:p-6">
            <div class="agent-list flex flex-col gap-2 min-h-12"
                :class="store.agents.length > 3 ? 'max-h-48 overflow-y-auto pr-2' : ''">
                <div v-if="store.agents.length === 0"
                    class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-2 text-dimmed gap-2">
                    <UIcon name="i-lucide-terminal" class="size-5" />
                    <span class="text-sm">{{ t('agent.noAgents') }}</span>
                </div>

                <div v-for="(agent, index) in store.agents" :key="index"
                    class="group flex flex-col gap-1 rounded-lg border border-default p-2 transition-colors hover:bg-elevated"
                    :class="{ 'opacity-50': !agent.enabled }">

                    <div class="flex flex-row items-start gap-2">
                        <UTooltip :text="agent.enabled ? t('agent.disable') : t('agent.enable')">
                            <UCheckbox v-model="agent.enabled" class="pt-0.5" />
                        </UTooltip>

                        <div class="flex min-w-0 flex-1 flex-col gap-1">
                            <div class="flex items-center gap-2 min-w-0">
                                <UIcon name="i-lucide-tag" class="size-4 shrink-0 text-dimmed" />

                                <UInput v-if="editingNameIndex === index" v-model="agent.name"
                                    :placeholder="t('agent.namePlaceholder')" class="flex-1" size="md" autofocus
                                    @keydown.enter="editingNameIndex = -1" @blur="editingNameIndex = -1" />

                                <div v-else class="flex min-w-0 flex-1 items-center cursor-pointer"
                                    :class="{ 'pointer-events-none': isAgentBusy(agent) }"
                                    @click="editingNameIndex = index">
                                    <span class="truncate text-md" :class="agent.name ? '' : 'text-dimmed italic'">
                                        {{ agent.name || agent.identifier || t('agent.unnamed') }}
                                    </span>
                                </div>
                            </div>

                            <div class="flex items-center gap-2 min-w-0">
                                <UIcon name="i-lucide-key" class="size-4 shrink-0 text-dimmed" />

                                <UInput v-if="editingIdIndex === index" v-model="agent.identifier"
                                    :placeholder="t('agent.idPlaceholder')" class="flex-1" size="md" autofocus
                                    @keydown.enter="onFinishEditId(agent, index)"
                                    @blur="onFinishEditId(agent, index)" />

                                <UTooltip v-else :text="agent.identifier" :disabled="!agent.identifier">
                                    <div class="flex min-w-0 flex-1 items-center cursor-pointer"
                                        :class="{ 'pointer-events-none': isAgentBusy(agent) }"
                                        @click="editingIdIndex = index">
                                        <span class="truncate text-xs font-mono text-dimmed">
                                            {{ agent.identifier || t('agent.clickToSetId') }}
                                        </span>
                                    </div>
                                </UTooltip>
                            </div>
                        </div>

                        <div class="flex shrink-0 items-start gap-1">
                            <StatusBadge :status="agent.status" />

                            <div class="flex flex-row gap-1 shrink-0">
                                <UTooltip :text="agent.status === 'connected' ? t('common.disconnect') : t('common.connect')">
                                    <UButton :color="agent.status === 'connected' ? 'error' : 'success'" variant="soft"
                                        :icon="getAgentButtonIcon(agent)" :loading="agent.status === 'connecting'"
                                        :disabled="!agent.enabled || isAgentBusy(agent) || (!agent.identifier.trim() && agent.status !== 'connected')"
                                        size="xs" @click="onToggleConnection(agent)" />
                                </UTooltip>

                                <UTooltip :text="t('agent.remove')">
                                    <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                                        :disabled="isAgentBusy(agent)" @click="onRemove(agent, index)" />
                                </UTooltip>
                            </div>
                        </div>
                    </div>

                    <div v-if="agent.errorMsg" class="pl-6 text-xs text-error truncate">
                        {{ agent.errorMsg }}
                    </div>
                </div>
            </div>

            <div class="p-2 sm:p-4">
                <UButton color="neutral" variant="ghost" icon="i-lucide-plus" :label="t('agent.add')" block
                    @click="onAddAgent" />
            </div>
        </div>
    </UCard>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from './agent/StatusBadge.vue'
import { disconnectAgent as apiDisconnect } from '@/api/http'
import useAgentControl from './useAgentControl'

import { useAgentStore, type AgentItem } from '@/stores/agent'
import { useStatusStore } from '@/stores/status'

const store = useAgentStore()
const statusStore = useStatusStore()
const { t } = useI18n()
const isTaskRunning = computed(() => statusStore.taskStatus === 'running')
const { doConnect, doDisconnect } = useAgentControl()

const editingNameIndex = ref(-1)
const editingIdIndex = ref(-1)

store.resetRuntimeState()

const headerBadgeColor = computed(() => {
    if (store.hasConnecting) return 'info' as const
    if (store.connectedCount > 0 && !store.hasError) return 'success' as const
    if (store.hasError) return 'error' as const
    return 'neutral' as const
})

const headerDotClass = computed(() => {
    if (store.hasConnecting) return 'bg-info'
    if (store.connectedCount > 0 && !store.hasError) return 'bg-success'
    if (store.hasError) return 'bg-error'
    return 'bg-gray-400 dark:bg-gray-500'
})

const headerBadgeLabel = computed(() => {
    if (store.hasConnecting) return t('agent.connecting')
    const total = store.agents.length
    if (store.connectedCount === total && total > 0) return t('agent.connectedCount', { count: total })
    if (store.connectedCount > 0) return t('agent.connectedPartial', { connected: store.connectedCount, total })
    if (store.hasError) return t('agent.error')
    return t('agent.idle')
})

function isAgentBusy(agent: AgentItem): boolean {
    return agent.status === 'connecting'
}

function getAgentButtonIcon(agent: AgentItem): string {
    if (agent.status === 'connected') return 'i-lucide-link-2-off'
    if (agent.status === 'connecting') return 'i-lucide-loader'
    return 'i-lucide-link'
}

function onAddAgent() {
    store.addAgent()
    editingIdIndex.value = store.agents.length - 1
}

function onRemove(agent: AgentItem, index: number) {
    if (isAgentBusy(agent)) return
    if (agent.status === 'connected') {
        apiDisconnect(agent.identifier)
    }
    store.removeByIndex(index)
}

function onFinishEditId(agent: AgentItem, index: number) {
    editingIdIndex.value = -1
    if (!agent.identifier.trim()) {
        store.removeByIndex(index)
    }
}

function onToggleConnection(agent: AgentItem) {
    if (isAgentBusy(agent)) return
    if (agent.status === 'connected') {
        doDisconnect(agent)
    } else {
        doConnect(agent)
    }
}
</script>

<style scoped>
.agent-list::-webkit-scrollbar {
    width: 4px;
}

.agent-list::-webkit-scrollbar-track {
    background: transparent;
}

.agent-list::-webkit-scrollbar-thumb {
    background-color: rgba(128, 128, 128, 0.3);
    border-radius: 2px;
}

.agent-list::-webkit-scrollbar-thumb:hover {
    background-color: rgba(128, 128, 128, 0.5);
}

.agent-list {
    scrollbar-width: thin;
    scrollbar-color: rgba(128, 128, 128, 0.3) transparent;
}
</style>
