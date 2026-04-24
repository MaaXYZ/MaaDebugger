<script setup lang="ts">
import { useDebugSettingsStore } from "@/stores/debugSettings";
import type { PipelineNotifyLevel } from "@/stores/debugSettings";

const debugSettingsStore = useDebugSettingsStore();

const pipelineNotifyLevel: { label: string; value: PipelineNotifyLevel }[] = [
  { label: "Error", value: "ERROR" },
  { label: "Warning", value: "WARNING" },
  { label: "Never", value: "NULL" },
];
</script>

<template>
  <UCard size="xl">
    <template #header>
      <div class="flex flex-row items-center justify-between gap-3" id="debug">
        <div class="flex flex-col">
          <span class="font-bold">Debug Workspace</span>
          <span class="text-sm text-dimmed">Configure the Debug page layout and sidebar behavior.</span>
        </div>
        <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" label="Reset" size="xs"
          @click="debugSettingsStore.reset()" />
      </div>
    </template>

    <div class="flex flex-col gap-3">
      <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3"
        id="debug-auto-collapse">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Auto collapse left sidebar on run start</span>
          <span class="text-sm text-dimmed">Collapse setup tabs automatically when a run starts, then keep that
            collapsed state until you expand them again.</span>
        </div>
        <USwitch :model-value="debugSettingsStore.autoCollapseLeftTabsOnRunStart" @update:model-value="
          debugSettingsStore.setAutoCollapseLeftTabsOnRunStart(
            Boolean($event),
          )
          " />
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3" id="debug-collapse">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Collapse left sidebar by default</span>
          <span class="text-sm text-dimmed">Start the Debug workspace in focused mode until you expand the
            setup sidebar.</span>
        </div>
        <USwitch :model-value="debugSettingsStore.leftTabsCollapsed" @update:model-value="
          debugSettingsStore.setLeftTabsCollapsed(Boolean($event))
          " />
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3"
        id="debug-watchResource">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Watch resource changes</span>
          <span class="text-sm text-dimmed">
            Whether to watch the resource changes to
            <b>automatically reload the resource</b>
          </span>
        </div>
        <USwitch :model-value="debugSettingsStore.watchResourceChange" @update:model-value="
          debugSettingsStore.setWatchResourceChange(Boolean($event))
          " />
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3"
        id="debug-checkResourceInterval">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Resource change check interval</span>
          <span class="text-sm text-dimmed">The interval (in milliseconds) at which to check for resource
            changes.</span>
        </div>
        <UInputNumber v-model="debugSettingsStore.watchResourceChangeInterval" :min="100" :max="10000" />
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3"
        id="debug-checkPipeline">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Check Pipeline Issues</span>
          <span class="text-sm text-dimmed">When resource loaded, check if the pipeline has any error or
            warning.</span>
        </div>
        <USwitch :model-value="debugSettingsStore.checkPipeline" @update:model-value="
          debugSettingsStore.setCheckPipeline(Boolean($event))
          " />
      </div>

      <div v-if="debugSettingsStore.checkPipeline"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3" id="debug-pipelineNotify">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Pipeline Notify Level</span>
          <span class="text-sm text-dimmed">If the pipeline has issues at or above the selected level, send
            notification.</span>
        </div>
        <USelect :model-value="debugSettingsStore.checkPipelineNotifyLevel" :items="pipelineNotifyLevel"
          class="min-w-48" @update:model-value="
            debugSettingsStore.setCheckPipelineNotifyLevel(
              $event as PipelineNotifyLevel,
            )
            " />
      </div>

      <div v-if="debugSettingsStore.checkPipeline"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3"
        id="debug-preventResourceLoaded">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">Prevent Resource Loaded</span>
          <span class="text-sm text-dimmed">If the pipeline has errors, prevent the resource from being loaded.</span>
        </div>
        <USwitch :model-value="debugSettingsStore.preventResourceLoaded" @update:model-value="
          debugSettingsStore.setPreventResourceLoaded(Boolean($event))
          " />
      </div>
    </div>
  </UCard>
</template>
