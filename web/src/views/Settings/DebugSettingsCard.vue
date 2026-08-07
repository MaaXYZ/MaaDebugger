<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useDebugSettingsStore } from "@/stores/debugSettings";
import type {
  PipelineNotifyLevel,
  SidebarAutoCollapseMode,
} from "@/stores/debugSettings";

const { t } = useI18n();
const debugSettingsStore = useDebugSettingsStore();

const pipelineNotifyLevel: { label: string; value: PipelineNotifyLevel }[] = [
  { label: t('settings.debug.error'), value: "ERROR" },
  { label: t('settings.debug.warning'), value: "WARNING" },
  { label: t('settings.debug.never'), value: "NULL" },
];

const sidebarAutoCollapseItems: {
  label: string;
  value: SidebarAutoCollapseMode;
}[] = [
  { label: t('settings.debug.sidebarNever'), value: "never" },
  { label: t('settings.debug.sidebarTask'), value: "task-running" },
  { label: t('settings.debug.sidebarAlways'), value: "always" },
];

// 通过 store 的 setter 写入（setter 会同步当前位置，避免模式与状态矛盾）
const sidebarAutoCollapse = computed({
  get: () => debugSettingsStore.sidebarAutoCollapse,
  set: (value: SidebarAutoCollapseMode) =>
    debugSettingsStore.setSidebarAutoCollapse(value),
});
</script>

<template>
  <UCard size="xl">
    <template #header>
      <div id="debug" class="flex flex-row items-center justify-between gap-3">
        <div class="flex flex-col">
          <span class="font-bold">{{ t('settings.debug.title') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.description') }}</span>
        </div>
        <UButton color="neutral" variant="ghost" icon="i-lucide-rotate-ccw" :label="t('settings.debug.reset')" size="xs"
          @click="debugSettingsStore.reset()" />
      </div>
    </template>

    <div class="flex flex-col gap-3">
      <div id="debug-auto-collapse"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.sidebarAutoCollapse') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.sidebarAutoCollapseDescription') }}</span>
        </div>
        <USelect v-model="sidebarAutoCollapse" :items="sidebarAutoCollapseItems" class="min-w-48" arrow />
      </div>

      <div id="debug-showTaskFps" class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.showFps') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.showFpsDescription') }}</span>
        </div>
        <USwitch :model-value="debugSettingsStore.showTaskFps" @update:model-value="
          debugSettingsStore.setShowTaskFps(Boolean($event))
          " />
      </div>

      <div id="debug-watchResource"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.watchResource') }}</span>
          <span class="text-sm text-dimmed">
            {{ t('settings.debug.watchResourceDescription') }}
          </span>
        </div>
        <USwitch :model-value="debugSettingsStore.watchResourceChange" @update:model-value="
          debugSettingsStore.setWatchResourceChange(Boolean($event))
          " />
      </div>

      <div id="debug-checkResourceInterval"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.checkInterval') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.checkIntervalDescription') }}</span>
        </div>
        <UInputNumber v-model="debugSettingsStore.watchResourceChangeInterval" :min="100" :max="10000" />
      </div>

      <div id="debug-checkPipeline"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.checkPipeline') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.checkPipelineDescription') }}</span>
        </div>
        <USwitch :model-value="debugSettingsStore.checkPipeline" @update:model-value="
          debugSettingsStore.setCheckPipeline(Boolean($event))
          " />
      </div>

      <div v-if="debugSettingsStore.checkPipeline" id="debug-pipelineNotify"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.notifyLevel') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.notifyLevelDescription') }}</span>
        </div>
        <USelect :model-value="debugSettingsStore.checkPipelineNotifyLevel" :items="pipelineNotifyLevel"
          class="min-w-48" @update:model-value="
            debugSettingsStore.setCheckPipelineNotifyLevel(
              $event as PipelineNotifyLevel,
            )
            " />
      </div>

      <div v-if="debugSettingsStore.checkPipeline" id="debug-preventResourceLoaded"
        class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
        <div class="flex flex-col gap-1">
          <span class="text-sm font-medium">{{ t('settings.debug.preventResourceLoaded') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.debug.preventResourceLoadedDescription') }}</span>
        </div>
        <USwitch :model-value="debugSettingsStore.preventResourceLoaded" @update:model-value="
          debugSettingsStore.setPreventResourceLoaded(Boolean($event))
          " />
      </div>
    </div>
  </UCard>
</template>
