<script setup lang="ts">
import { useDebugWorkspaceSettingsStore } from '@/stores/debugWorkspaceSettings'

const debugWorkspaceSettingsStore = useDebugWorkspaceSettingsStore()
</script>

<template>
    <UCard size="xl">
        <template #header>
            <div class="flex flex-row items-center justify-between gap-3">
                <div class="flex flex-col">
                    <span class="font-bold">Debug Workspace</span>
                    <span class="text-sm text-dimmed">Configure the Debug page layout and sidebar behavior.</span>
                </div>
                <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" label="Reset" size="xs"
                    @click="debugWorkspaceSettingsStore.reset()" />
            </div>
        </template>

        <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Auto collapse left sidebar on run start</span>
                    <span class="text-sm text-dimmed">Collapse setup tabs automatically when a run starts, then keep
                        that
                        collapsed state until you expand them again.</span>
                </div>
                <USwitch :model-value="debugWorkspaceSettingsStore.autoCollapseLeftTabsOnRunStart"
                    @update:model-value="debugWorkspaceSettingsStore.setAutoCollapseLeftTabsOnRunStart(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Collapse left sidebar by default</span>
                    <span class="text-sm text-dimmed">Start the Debug workspace in focused mode until you expand the
                        setup
                        sidebar.</span>
                </div>
                <USwitch :model-value="debugWorkspaceSettingsStore.leftTabsCollapsed"
                    @update:model-value="debugWorkspaceSettingsStore.setLeftTabsCollapsed(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Watch resource changes</span>
                    <span class="text-sm text-dimmed">
                        Whether to watch the resource changes to <b>automatically reload the resource</b>
                    </span>
                </div>
                <USwitch :model-value="debugWorkspaceSettingsStore.watchResourceChange"
                    @update:model-value="debugWorkspaceSettingsStore.setWatchResourceChange(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Resource change check interval</span>
                    <span class="text-sm text-dimmed">The interval (in milliseconds) at which to check for resource
                        changes.</span>
                </div>
                <UInputNumber v-model="debugWorkspaceSettingsStore.watchResourceChangeInterval" :min="100"
                    :max="10000" />
            </div>
        </div>
    </UCard>
</template>
