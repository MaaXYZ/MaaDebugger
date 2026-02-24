<template>
    <UCard class="w-full" size="xl">
        <template #header>
            <div class="flex flex-row items-center gap-2 min-h-10">
                <span class="font-bold">Agent</span>
            </div>
        </template>

        <template #default>
            <div class="agent-list flex flex-col gap-2 min-h-24"
                :class="agents.length > 1 ? 'max-h-24 overflow-y-auto pr-2' : ''">
                <div v-if="agents.length === 0"
                    class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-3 text-dimmed gap-2">
                    <UIcon name="i-lucide-terminal" class="size-5" />
                    <span class="text-sm">No agents added</span>
                </div>

                <div v-for="(agent, index) in agents" :key="agent.id"
                    class="group flex flex-col gap-2 rounded-lg border border-default p-3 transition-colors hover:bg-elevated">

                    <!-- Row 1: Name + Status + Actions -->
                    <div class="flex flex-row items-center gap-2">
                        <!-- Agent Name (editable inline) -->
                        <UIcon name="i-lucide-tag" class="size-4 shrink-0 text-dimmed" />

                        <UInput v-if="agent.editingName" v-model="agent.name" placeholder="Agent name..." class="flex-1"
                            size="sm" autofocus @keydown.enter="onFinishEditName(index)"
                            @blur="onFinishEditName(index)" />

                        <div v-else class="flex-1 flex items-center min-w-0 cursor-pointer"
                            :class="{ 'pointer-events-none': isAgentBusy(agent) }" @click="onEditName(index)">
                            <span class="truncate text-sm font-medium" :class="agent.name ? '' : 'text-dimmed italic'">
                                {{ agent.name || agent.agentId || 'Unnamed agent' }}
                            </span>
                        </div>

                        <!-- Status Badge -->
                        <StatusBadge :status="agent.status" />

                        <!-- Action Buttons -->
                        <div class="flex flex-row gap-1 shrink-0">
                            <UTooltip :text="agent.status === 'connected' ? 'Disconnect' : 'Connect'">
                                <UButton :color="agent.status === 'connected' ? 'error' : 'success'"
                                    :variant="agent.status === 'connected' ? 'soft' : 'soft'"
                                    :icon="getAgentButtonIcon(agent)" :loading="agent.status === 'connecting'"
                                    :disabled="isAgentBusy(agent) || (!agent.agentId.trim() && agent.status !== 'connected')"
                                    size="xs" @click="onToggleConnection(index)" />
                            </UTooltip>

                            <UTooltip text="Remove">
                                <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                                    :disabled="isAgentBusy(agent)" @click="onRemove(index)" />
                            </UTooltip>
                        </div>
                    </div>

                    <!-- Row 2: Agent ID (UUID) -->
                    <div class="flex flex-row items-center gap-2">
                        <UIcon name="i-lucide-fingerprint" class="size-4 shrink-0 text-dimmed" />

                        <UInput v-if="agent.editingId" v-model="agent.agentId" placeholder="Enter agent ID (UUID)..."
                            class="flex-1" size="sm" autofocus @keydown.enter="onFinishEditId(index)"
                            @blur="onFinishEditId(index)" />

                        <UTooltip v-else :text="agent.agentId" :disabled="!agent.agentId">
                            <div class="flex-1 flex items-center min-w-0 cursor-pointer"
                                :class="{ 'pointer-events-none': isAgentBusy(agent) }" @click="onEditId(index)">
                                <span class="truncate text-xs font-mono text-dimmed">
                                    {{ agent.agentId || 'Click to set agent ID...' }}
                                </span>
                            </div>
                        </UTooltip>
                    </div>
                </div>
            </div>
        </template>

        <template #footer>
            <UButton color="neutral" variant="ghost" icon="i-lucide-plus" label="Add agent" block @click="onAddAgent" />
        </template>
    </UCard>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import StatusBadge from './agent/StatusBadge.vue'
import type { ConnectionStatus } from './agent/types'

// --- Types ---
interface AgentItem {
    id: number
    name: string
    agentId: string
    status: ConnectionStatus
    editingName: boolean
    editingId: boolean
}

// --- State ---
let nextId = 0
const agents = ref<AgentItem[]>([])

// --- Helpers ---
function isAgentBusy(agent: AgentItem): boolean {
    return agent.status === 'connecting'
}

function getAgentButtonIcon(agent: AgentItem): string {
    if (agent.status === 'connected') return 'i-material-symbols:link-off'
    if (agent.status === 'connecting') return 'i-lucide-loader'
    return 'i-material-symbols:add-link'
}

// --- Add / Remove ---
function onAddAgent() {
    agents.value.push({
        id: nextId++,
        name: '',
        agentId: '',
        status: 'idle',
        editingName: false,
        editingId: true,
    })
}

function onRemove(index: number) {
    const agent = agents.value[index]
    if (agent && !isAgentBusy(agent)) {
        agents.value.splice(index, 1)
    }
}

// --- Name Edit ---
function onEditName(index: number) {
    const agent = agents.value[index]
    if (agent && !isAgentBusy(agent)) {
        agent.editingName = true
    }
}

function onFinishEditName(index: number) {
    const agent = agents.value[index]
    if (!agent) return
    agent.editingName = false
}

// --- ID Edit ---
function onEditId(index: number) {
    const agent = agents.value[index]
    if (agent && !isAgentBusy(agent)) {
        agent.editingId = true
    }
}

function onFinishEditId(index: number) {
    const agent = agents.value[index]
    if (!agent) return
    agent.editingId = false
    // Remove agent if ID is empty (never been set)
    if (!agent.agentId.trim()) {
        agents.value.splice(index, 1)
    }
}

// --- Connection Actions ---
function onToggleConnection(index: number) {
    const agent = agents.value[index]
    if (!agent || isAgentBusy(agent)) return

    if (agent.status === 'connected') {
        disconnectAgent(agent)
    } else {
        connectAgent(agent)
    }
}

async function connectAgent(agent: AgentItem) {
    const id = agent.agentId.trim()
    if (!id) return

    agent.status = 'connecting'

    try {
        // TODO: Replace with actual connection logic
        await simulateConnection(id)
        agent.status = 'connected'
    } catch (error: any) {
        if (error?.message === 'timeout') {
            agent.status = 'timeout'
        } else {
            agent.status = 'failed'
        }
    }
}

function disconnectAgent(agent: AgentItem) {
    // TODO: Replace with actual disconnection logic
    agent.status = 'idle'
}

// --- Simulation (to be replaced) ---
function simulateConnection(_id: string): Promise<void> {
    return new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
            reject(new Error('timeout'))
        }, 10000)

        setTimeout(() => {
            clearTimeout(timeout)
            if (Math.random() > 0.3) {
                resolve()
            } else {
                reject(new Error('connection_failed'))
            }
        }, 2000)
    })
}

// --- Public API ---
function getAgents() {
    return agents.value
        .filter(a => a.agentId.trim())
        .map(a => ({ name: a.name, agentId: a.agentId, status: a.status }))
}

function getConnectedAgents() {
    return agents.value
        .filter(a => a.status === 'connected')
        .map(a => ({ name: a.name, agentId: a.agentId }))
}

defineExpose({
    agents,
    getAgents,
    getConnectedAgents,
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
