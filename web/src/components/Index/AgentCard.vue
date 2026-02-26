<template>
    <UCard class="w-full" size="xl" :ui="{ body: 'p-0 sm:p-0', footer: 'p-0 sm:p-0' }">
        <template #header>
            <div class="flex flex-col gap-2">
                <div class="flex flex-row items-center justify-between gap-4">
                    <div class="flex items-center gap-2">
                        <span class="font-bold">Agent</span>
                        <UBadge v-if="store.agents.length > 0" :color="headerBadgeColor" variant="subtle" size="sm"
                            class="gap-1.5">
                            <span class="relative flex size-2">
                                <span v-if="store.hasConnecting"
                                    class="absolute inline-flex size-full animate-ping rounded-full bg-info opacity-75" />
                                <span class="relative inline-flex size-2 rounded-full" :class="headerDotClass" />
                            </span>
                            {{ headerBadgeLabel }}
                        </UBadge>
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
                    <div class="p-4 sm:p-6">
                        <div class="agent-list flex flex-col gap-2 min-h-12"
                            :class="store.agents.length > 1 ? 'max-h-24 overflow-y-auto pr-2' : ''">
                            <div v-if="store.agents.length === 0"
                                class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-3 text-dimmed gap-2">
                                <UIcon name="i-lucide-terminal" class="size-5" />
                                <span class="text-sm">No agents added</span>
                            </div>

                            <div v-for="(agent, index) in store.agents" :key="agent.identifier || index"
                                class="group flex flex-col gap-2 rounded-lg border border-default p-3 transition-colors hover:bg-elevated">

                                <!-- Row 1: Name + Status + Actions -->
                                <div class="flex flex-row items-center gap-2">
                                    <UIcon name="i-lucide-tag" class="size-4 shrink-0 text-dimmed" />

                                    <UInput v-if="editingNameIndex === index" v-model="agent.name"
                                        placeholder="Agent name..." class="flex-1" size="sm" autofocus
                                        @keydown.enter="editingNameIndex = -1" @blur="editingNameIndex = -1" />

                                    <div v-else class="flex-1 flex items-center min-w-0 cursor-pointer"
                                        :class="{ 'pointer-events-none': isAgentBusy(agent) }"
                                        @click="editingNameIndex = index">
                                        <span class="truncate text-sm font-medium"
                                            :class="agent.name ? '' : 'text-dimmed italic'">
                                            {{ agent.name || agent.identifier || 'Unnamed agent' }}
                                        </span>
                                    </div>

                                    <StatusBadge :status="agent.status" />

                                    <div class="flex flex-row gap-1 shrink-0">
                                        <UTooltip :text="agent.status === 'connected' ? 'Disconnect' : 'Connect'">
                                            <UButton :color="agent.status === 'connected' ? 'error' : 'success'"
                                                variant="soft" :icon="getAgentButtonIcon(agent)"
                                                :loading="agent.status === 'connecting'"
                                                :disabled="isAgentBusy(agent) || (!agent.identifier.trim() && agent.status !== 'connected')"
                                                size="xs" @click="onToggleConnection(agent)" />
                                        </UTooltip>

                                        <UTooltip text="Remove">
                                            <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                                                :disabled="isAgentBusy(agent)" @click="onRemove(agent, index)" />
                                        </UTooltip>
                                    </div>
                                </div>

                                <!-- Row 2: Identifier -->
                                <div class="flex flex-row items-center gap-2">
                                    <UIcon name="i-lucide-fingerprint" class="size-4 shrink-0 text-dimmed" />

                                    <UInput v-if="editingIdIndex === index" v-model="agent.identifier"
                                        placeholder="Enter agent identifier..." class="flex-1" size="sm" autofocus
                                        @keydown.enter="onFinishEditId(agent, index)"
                                        @blur="onFinishEditId(agent, index)" />

                                    <UTooltip v-else :text="agent.identifier" :disabled="!agent.identifier">
                                        <div class="flex-1 flex items-center min-w-0 cursor-pointer"
                                            :class="{ 'pointer-events-none': isAgentBusy(agent) }"
                                            @click="editingIdIndex = index">
                                            <span class="truncate text-xs font-mono text-dimmed">
                                                {{ agent.identifier || 'Click to set identifier...' }}
                                            </span>
                                        </div>
                                    </UTooltip>
                                </div>

                                <!-- Row 3: Error message -->
                                <div v-if="agent.errorMsg" class="flex flex-row items-start gap-2 pl-6">
                                    <span class="text-xs text-error truncate">{{ agent.errorMsg }}</span>
                                </div>
                            </div>
                        </div>

                        <div class="p-2 sm:p-4">
                            <UButton color="neutral" variant="ghost" icon="i-lucide-plus" label="Add agent" block
                                @click="onAddAgent" />
                        </div>
                    </div>
                </template>
            </UCollapsible>
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import StatusBadge from './agent/StatusBadge.vue'
import { connectAgent as apiConnect, disconnectAgent as apiDisconnect } from '@/api/http'
import type { AgentInfo } from '@/api/http'
import { latestAgentUpdate } from '@/api/agentEvents'
import { useAgentStore, type AgentItem } from '@/stores/agent'

const store = useAgentStore()
const showFullCard = ref(false)

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
    if (store.hasConnecting) return 'Connecting...'
    const total = store.agents.length
    if (store.connectedCount === total && total > 0) return `${total} connected`
    if (store.connectedCount > 0) return `${store.connectedCount}/${total} connected`
    if (store.hasError) return 'Error'
    return 'Idle'
})

const summaryText = computed(() => {
    if (store.agents.length === 0) return 'No agents configured'
    const names = store.agents
        .map(a => a.name || a.identifier || 'Unnamed')
        .join(', ')
    return `${store.agents.length} agent${store.agents.length > 1 ? 's' : ''}: ${names}`
})

function isAgentBusy(agent: AgentItem): boolean {
    return agent.status === 'connecting'
}

function getAgentButtonIcon(agent: AgentItem): string {
    if (agent.status === 'connected') return 'i-material-symbols:link-off'
    if (agent.status === 'connecting') return 'i-lucide-loader'
    return 'i-material-symbols:add-link'
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

async function doConnect(agent: AgentItem) {
    const identifier = agent.identifier.trim()
    if (!identifier) return

    agent.status = 'connecting'
    agent.errorMsg = ''

    try {
        const result = await apiConnect(identifier)

        if (result.succeed) {
            agent.status = 'connected'
            agent.errorMsg = ''
        } else {
            agent.status = 'failed'
            agent.errorMsg = result.msg || 'Connection failed'
        }
    } catch (error: any) {
        agent.status = 'failed'
        agent.errorMsg = error?.message || 'Unknown error'
    }
}

async function doDisconnect(agent: AgentItem) {
    await apiDisconnect(agent.identifier)
    agent.status = 'idle'
    agent.errorMsg = ''
}

watch(latestAgentUpdate, (serverAgents) => {
    const serverMap = new Map<string, AgentInfo>(serverAgents.map(a => [a.identifier, a]))

    for (const agent of store.agents) {
        const serverEntry = serverMap.get(agent.identifier)
        if (serverEntry) {
            if (serverEntry.status === 'connected') {
                agent.status = 'connected'
                agent.errorMsg = ''
            } else if (serverEntry.status === 'failed' && agent.status === 'connected') {
                agent.status = 'failed'
                agent.errorMsg = serverEntry.error || 'Connection lost'
            }
        }
    }
})
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
