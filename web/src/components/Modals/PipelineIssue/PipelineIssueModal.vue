<template>
  <UModal v-model:open="open" :ui="{ content: 'sm:max-w-[85vw] sm:w-[85vw]' }">
    <template #header>
      <div class="flex flex-row items-center gap-2 flex-wrap">
        <span class="text-sm text-highlighted font-semibold"
          >Pipeline Issues</span
        >
        <UBadge
          color="error"
          variant="subtle"
          :label="`Error ${errors.length}`"
        />
        <UBadge
          color="warning"
          variant="subtle"
          :label="`Warning ${warnings.length}`"
        />
      </div>
    </template>

    <template #body>
      <div class="flex flex-col gap-4">
        <div
          v-if="issues.length === 0"
          class="text-sm text-dimmed p-4 text-center"
        >
          No issues found
        </div>

        <div
          v-else
          class="flex flex-col gap-3 max-h-[60vh] overflow-y-auto pr-1"
        >
          <div
            v-for="(issue, index) in issues"
            :key="`${issue.code}-${issue.path}-${issue.line}-${issue.task ?? ''}-${issue.msg}-${index}`"
            class="rounded-lg border border-default p-3 flex flex-col gap-2"
          >
            <div class="flex flex-row items-center gap-2 flex-wrap">
              <UBadge
                v-if="issue.task"
                color="info"
                variant="subtle"
                :label="`${issue.task}`"
              />
              <UBadge
                variant="subtle"
                :label="issue.code"
                :color="issue.level === 'error' ? 'error' : 'warning'"
              />
              <span class="text-sm text-dimmed">#{{ index + 1 }}</span>
            </div>

            <div class="text-sm text-dimmed break-all font-mono">
              {{ issue.path }}:{{ issue.line }}
            </div>

            <div class="text-md text-highlighted wrap-break-word">
              {{ issue.msg }}
            </div>
          </div>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from "vue";

import type { CheckResponse } from "@/types/pipeline";

const open = defineModel<boolean>("open", { default: false });
const props = defineProps<{
  errors: CheckResponse[];
  warnings: CheckResponse[];
}>();

const issues = computed<CheckResponse[]>(() => [
  ...props.errors,
  ...props.warnings,
]);
</script>
