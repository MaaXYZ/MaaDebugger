<template>
  <UCard
    class="w-full transition-opacity duration-200"
    :class="{ 'opacity-50 pointer-events-none': isCardDisabled }"
    size="xl"
    :ui="{
      root: 'bg-default/70 ring-default/70',
      header: 'p-4 sm:px-5',
      body: 'p-0 sm:p-0',
      footer: 'p-0 sm:p-0',
    }"
  >
    <template #header>
      <div class="flex flex-col gap-2">
        <div class="flex flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-2">
            <span class="font-bold">{{ t('resource.title') }}</span>
            <UBadge
              :color="statusColor"
              variant="subtle"
              size="sm"
              class="gap-1.5"
            >
              <span class="relative flex size-2">
                <span
                  v-if="statusStore.resourceStatus === 'loading'"
                  class="absolute inline-flex size-full animate-ping rounded-full bg-warning opacity-75"
                ></span>
                <span
                  class="relative inline-flex size-2 rounded-full"
                  :class="dotClass"
                ></span>
              </span>
              {{ statusLabel }}
            </UBadge>
          </div>
          <div class="flex flex-row items-center gap-2">
            <USelect
              v-model="resourceStore.activeProfileId"
              :items="resourceStore.profileSelectItems"
              class="w-32"
              size="xl"
              arrow
              :disabled="isCardDisabled"
            />
            <UDropdownMenu :items="profileMenuItems">
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-ellipsis-vertical"
                size="xs"
                :disabled="isCardDisabled"
              />
            </UDropdownMenu>
          </div>
        </div>
      </div>
    </template>

    <div class="p-4 sm:p-6">
      <div
        class="resource-list flex flex-col gap-2 min-h-12"
        :class="
          resourceStore.activePaths.length > 3
            ? 'max-h-48 overflow-y-auto pr-2'
            : ''
        "
      >
        <div
          v-if="resourceStore.activePaths.length === 0"
          class="flex flex-row items-center justify-center rounded-lg border border-dashed border-default p-2 text-dimmed gap-2"
        >
          <UIcon name="i-lucide-folder-open" class="size-5" />
          <span class="text-sm">{{ t('resource.noPaths') }}</span>
        </div>

        <div
          v-for="(item, index) in resourceStore.activePaths"
          :key="item.id"
          class="group flex flex-col gap-1 rounded-lg border border-default p-2 transition-colors hover:bg-elevated"
          :class="{ 'opacity-50': !item.enabled }"
          :draggable="resourceStore.activePaths.length > 1"
          @dragstart="onDragStart(index)"
          @dragover.prevent="onDragOver(index)"
          @dragend="onDragEnd"
        >
          <div class="flex flex-row items-center gap-2">
            <!-- Enable/Disable Checkbox -->
            <UCheckbox v-model="item.enabled" />

            <!-- Drag Handle -->
            <div
              v-if="resourceStore.activePaths.length > 1"
              class="cursor-grab text-dimmed hover:text-default active:cursor-grabbing"
            >
              <UIcon name="i-lucide-grip-vertical" class="size-5" />
            </div>

            <!-- Path Input -->
            <UInput
              v-if="editingIndex === index"
              v-model="item.path"
              :placeholder="t('resource.pathPlaceholder')"
              class="flex-1"
              size="md"
              autofocus
              :color="pathErrors[item.id] ? 'error' : 'neutral'"
              @keydown.enter="onFinishEdit(index)"
              @blur="onFinishEdit(index)"
            />

            <!-- Path Display -->
            <UTooltip v-else :text="item.path" :disabled="!item.path">
              <div
                class="flex-1 flex items-center gap-2 min-w-0 cursor-pointer"
                @click="onEdit(index)"
              >
                <span
                  class="truncate text-md"
                  :class="item.path ? '' : 'text-dimmed italic'"
                >
                  {{ item.path || t('resource.clickToEditPath') }}
                </span>
              </div>
            </UTooltip>

            <!-- Action Buttons -->
            <div class="flex flex-row gap-1 shrink-0">
              <UTooltip :text="t('common.edit')">
                <UButton
                  color="neutral"
                  variant="ghost"
                  icon="i-lucide-square-pen"
                  size="xs"
                  @click="onEdit(index)"
                />
              </UTooltip>
              <UTooltip :text="t('common.remove')">
                <UButton
                  color="error"
                  variant="ghost"
                  icon="i-lucide-trash-2"
                  size="xs"
                  @click="onRemovePath(index, item.id)"
                />
              </UTooltip>
            </div>
          </div>

          <div
            v-if="pathErrors[item.id]"
            class="pl-6 text-xs text-error truncate"
          >
            {{ pathErrors[item.id] }}
          </div>
        </div>
      </div>

      <div class="p-2 sm:p-4">
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-lucide-plus"
          :label="t('resource.addPath')"
          block
          @click="onAddPath"
        />
      </div>

      <!-- Load Button -->
      <div class="px-2 sm:px-4 pb-2 sm:pb-4">
        <UButton
          color="primary"
          icon="i-lucide-download"
          :label="t('resource.load')"
          block
          size="xl"
          :loading="isLoading"
          :disabled="enabledPaths.length === 0 || isLoading"
          @click="onLoadResource"
        />
      </div>
    </div>
  </UCard>

  <!-- Rename Profile Modal -->
  <UModal
    v-model:open="renameModalOpen"
    :title="t('resource.renameProfile')"
    :description="t('resource.renameProfileDescription')"
  >
    <template #body>
      <UInput
        v-model="renameInput"
        :placeholder="t('resource.profileNamePlaceholder')"
        size="xl"
        autofocus
        @keydown.enter="onConfirmRename"
      />
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton
          color="neutral"
          variant="ghost"
          :label="t('common.cancel')"
          @click="() => { renameModalOpen = false }"
        />
        <UButton
          color="primary"
          :label="t('resource.rename')"
          :disabled="!renameInput.trim()"
          @click="onConfirmRename"
        />
      </div>
    </template>
  </UModal>

  <PipelineIssueModal
    v-model:open="pipelineIssueModalOpen"
    :errors="pipelineErrors"
    :warnings="pipelineWarns"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { checkPathExists } from "@/api/http";
import { useResourceStore } from "@/stores/resource";
import { useStatusStore } from "@/stores/status";
import { useSignalStore } from "@/stores/signal";
import PipelineIssueModal from "@/components/Modals/PipelineIssue/PipelineIssueModal.vue";
import useResourceControl from "./useResourceControl";

const toast = useToast();
const { t } = useI18n();
const resourceStore = useResourceStore();
const statusStore = useStatusStore();
const signalStore = useSignalStore();
const {
  enabledPaths,
  onLoadResource: triggerLoadResource,
  pipelineIssueModalOpen,
  pipelineErrors,
  pipelineWarns,
} = useResourceControl();

// --- UI State (not persisted) ---
const editingIndex = ref<number | null>(null);
const pathErrors = ref<Record<number, string>>({});

// --- Resource Status ---
const isLoading = computed(() => statusStore.resourceStatus === "loading");
const isConnecting = computed(
  () => statusStore.controllerStatus === "connecting",
);

const isTaskRunning = computed(() => statusStore.taskStatus === "running");

/**
 * 当 controller 正在连接、资源正在加载或任务正在运行时，禁用整个 ResourceCard 的交互
 */
const isCardDisabled = computed(
  () => isConnecting.value || isLoading.value || isTaskRunning.value,
);

const statusLabel = computed(() => {
  switch (statusStore.resourceStatus) {
    case "loaded":
      return t('resource.loaded');
    case "loading":
      return t('resource.loading');
    case "failed":
      return t('resource.failed');
    case "unloaded":
    default:
      return t('resource.idle');
  }
});

const statusColor = computed(() => {
  switch (statusStore.resourceStatus) {
    case "loaded":
      return "success" as const;
    case "loading":
      return "warning" as const;
    case "failed":
      return "error" as const;
    case "unloaded":
    default:
      return "neutral" as const;
  }
});

const dotClass = computed(() => {
  switch (statusStore.resourceStatus) {
    case "loaded":
      return "bg-success";
    case "loading":
      return "bg-warning";
    case "failed":
      return "bg-error";
    case "unloaded":
    default:
      return "bg-gray-400 dark:bg-gray-500";
  }
});

// --- Profile Menu ---
const profileMenuItems = computed(() => [
  [
    {
      label: t('resource.newProfile'),
      icon: "i-lucide-plus",
      onSelect: () => resourceStore.addProfile(),
    },
    {
      label: t('resource.renameProfileMenu'),
      icon: "i-lucide-pencil",
      onSelect: onRenameProfile,
    },
  ],
  [
    {
      label: t('resource.deleteProfile'),
      icon: "i-lucide-trash-2",
      color: "error" as const,
      disabled: resourceStore.profiles.length <= 1,
      onSelect: () => resourceStore.deleteProfile(),
    },
  ],
]);

// --- Drag & Drop ---
const dragIndex = ref<number | null>(null);

function onDragStart(index: number) {
  dragIndex.value = index;
}

function onDragOver(index: number) {
  if (dragIndex.value === null || dragIndex.value === index) return;
  resourceStore.reorderPaths(dragIndex.value, index);
  dragIndex.value = index;
}

function onDragEnd() {
  dragIndex.value = null;
}

// --- Path Actions ---
function onAddPath() {
  resourceStore.addPath();
  editingIndex.value = resourceStore.activePaths.length - 1;
}

function onEdit(index: number) {
  editingIndex.value = index;
}

function clearPathError(pathId: number) {
  if (!pathErrors.value[pathId]) return;
  const { [pathId]: _removed, ...rest } = pathErrors.value;
  pathErrors.value = rest;
}

async function validatePath(path: string, pathId: number): Promise<boolean> {
  const trimmed = path.trim();
  if (!trimmed) {
    clearPathError(pathId);
    return false;
  }

  const result = await checkPathExists(trimmed, "dir");
  const exists = Boolean(result.succeed && result.data?.exists);

  if (!exists) {
    pathErrors.value = {
      ...pathErrors.value,
      [pathId]: result.msg || t('resource.pathValidationFailed'),
    };
    return false;
  }

  clearPathError(pathId);
  return true;
}

async function onFinishEdit(index: number) {
  editingIndex.value = null;

  // Remove empty paths automatically
  const item = resourceStore.activePaths[index];
  if (item && !item.path.trim()) {
    clearPathError(item.id);
    resourceStore.removePath(index);
    return;
  }

  if (item) {
    await validatePath(item.path, item.id);
  }
}

function onRemovePath(index: number, pathId: number) {
  clearPathError(pathId);
  resourceStore.removePath(index);
}

async function onLoadResource() {
  const validationList = await Promise.all(
    resourceStore.activePaths
      .filter((item) => item.enabled)
      .map(async (item) => ({
        id: item.id,
        ok: await validatePath(item.path, item.id),
      })),
  );

  const hasInvalidPath = validationList.some((item) => !item.ok);
  if (hasInvalidPath) {
    toast.add({
      id: "res-path-toast",
      title: t('resource.invalidPath'),
      description: t('resource.invalidPathDescription'),
      icon: "i-lucide-circle-x",
      color: "error",
    });
    return;
  }

  await triggerLoadResource();
}

// --- Rename Modal ---
const renameModalOpen = ref(false);
const renameInput = ref("");

function onRenameProfile() {
  renameInput.value = resourceStore.activeProfile.name;
  renameModalOpen.value = true;
}

function onConfirmRename() {
  resourceStore.renameProfile(renameInput.value);
  renameModalOpen.value = false;
}

watch(
  () => signalStore.reloadResource,
  (status) => {
    if (status > 0) {
      void triggerLoadResource({
        manual: false,
        changedPath: signalStore.changedPath,
      });
    }
  },
);
</script>

<style scoped>
.resource-list::-webkit-scrollbar {
  width: 4px;
}

.resource-list::-webkit-scrollbar-track {
  background: transparent;
}

.resource-list::-webkit-scrollbar-thumb {
  background-color: rgba(128, 128, 128, 0.3);
  border-radius: 2px;
}

.resource-list::-webkit-scrollbar-thumb:hover {
  background-color: rgba(128, 128, 128, 0.5);
}

.resource-list {
  scrollbar-width: thin;
  scrollbar-color: rgba(128, 128, 128, 0.3) transparent;
}
</style>
