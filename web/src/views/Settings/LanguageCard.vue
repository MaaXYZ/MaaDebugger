<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useLocaleStore } from "@/stores/locale";
import type { LocaleChoice } from "@/i18n";

const { t } = useI18n();
const localeStore = useLocaleStore();

const languageItems = computed(() => [
  { label: t('settings.language.auto'), value: "auto" as LocaleChoice },
  { label: t('settings.language.zhCN'), value: "zh-CN" as LocaleChoice },
  { label: t('settings.language.en'), value: "en" as LocaleChoice },
]);

const languageChoice = computed({
  get: () => localeStore.choice,
  set: (value: LocaleChoice) => {
    void localeStore.setChoice(value);
  },
});
</script>

<template>
  <UCard size="xl">
    <template #header>
      <div id="language" class="flex flex-row items-center justify-between gap-3">
        <div class="flex flex-col">
          <span class="font-bold">{{ t('settings.language.title') }}</span>
          <span class="text-sm text-dimmed">{{ t('settings.language.description') }}</span>
        </div>
      </div>
    </template>

    <div class="flex items-center justify-between gap-4 rounded-lg border border-default p-3">
      <div class="flex flex-col gap-1">
        <span class="text-sm font-medium">{{ t('locale.language') }}</span>
        <span class="text-sm text-dimmed text-bold">{{ t('locale.description') }}</span>
      </div>
      <USelect v-model="languageChoice" :items="languageItems" value-key="value" class="min-w-44" arrow />
    </div>
  </UCard>
</template>
