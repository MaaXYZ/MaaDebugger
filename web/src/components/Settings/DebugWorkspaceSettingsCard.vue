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
                    <span class="text-sm font-medium">Auto hide left sidebar while running</span>
                    <span class="text-sm text-dimmed">Temporarily hide setup tabs so Task and Task Detail get more space during runs.</span>
                </div>
                <USwitch :model-value="debugWorkspaceSettingsStore.autoHideLeftTabsWhenRunning"
                    @update:model-value="debugWorkspaceSettingsStore.setAutoHideLeftTabsWhenRunning(Boolean($event))" />
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
                <div class="flex flex-col gap-1">
                    <span class="text-sm font-medium">Collapse left sidebar by default</span>
                    <span class="text-sm text-dimmed">Start the Debug workspace in focused mode until you expand the setup sidebar.</span>
                </div>
                <USwitch :model-value="debugWorkspaceSettingsStore.leftTabsCollapsed"
                    @update:model-value="debugWorkspaceSettingsStore.setLeftTabsCollapsed(Boolean($event))" />
            </div>
        </div>
    </UCard>
</template>
