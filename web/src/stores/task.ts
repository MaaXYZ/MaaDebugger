import { computed, ref } from "vue";
import { defineStore } from "pinia";
import type {
  InterfaceTaskCandidate,
  InterfaceTaskOptionCase,
  InterfaceTaskOptionDefinition,
} from "@/types/interface";

export interface InterfaceTaskOptionSelection {
  optionName: string;
  caseName: string;
}

function cloneTasks(tasks: InterfaceTaskCandidate[]): InterfaceTaskCandidate[] {
  return JSON.parse(JSON.stringify(tasks)) as InterfaceTaskCandidate[];
}

function cloneObject<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function normalizeOverrideValue(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => normalizeOverrideValue(item));
  }

  if (value && typeof value === "object") {
    const source = value as Record<string, unknown>;
    const result: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(source)) {
      result[key] = normalizeOverrideValue(nested);
    }

    if (typeof result.enable === "boolean" && result.enabled === undefined) {
      result.enabled = result.enable;
      delete result.enable;
    }

    return result;
  }

  return value;
}

function deepMergeOverride(
  base: Record<string, unknown>,
  patch: Record<string, unknown>,
): Record<string, unknown> {
  const next = cloneObject(base);

  for (const [key, value] of Object.entries(patch)) {
    const normalizedValue = normalizeOverrideValue(value);
    const current = next[key];

    if (
      current &&
      typeof current === "object" &&
      !Array.isArray(current) &&
      normalizedValue &&
      typeof normalizedValue === "object" &&
      !Array.isArray(normalizedValue)
    ) {
      next[key] = deepMergeOverride(
        current as Record<string, unknown>,
        normalizedValue as Record<string, unknown>,
      );
      continue;
    }

    next[key] = normalizedValue;
  }

  return next;
}

function parseStoredJsonObject(raw: string): Record<string, unknown> {
  const trimmed = raw.trim();
  if (!trimmed) return {};

  try {
    const parsed = JSON.parse(trimmed);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
  } catch {
    // keep invalid manual JSON untouched in editor; runtime merge falls back to {}
  }

  return {};
}

function getOptionCases(
  optionDef?: InterfaceTaskOptionDefinition,
): InterfaceTaskOptionCase[] {
  return optionDef?.cases ?? [];
}

function getDefaultCaseName(optionDef?: InterfaceTaskOptionDefinition): string {
  if (!optionDef) return "";
  return optionDef.default_case || getOptionCases(optionDef)[0]?.name || "";
}

export const useTaskStore = defineStore(
  "task",
  () => {
    const selectedEntry = ref("");
    /** Pipeline override JSON string (derived result + manual patch), persisted */
    const overrideJson = ref("{}");
    /** User-editable extra patch merged on top of interface-derived override */
    const manualOverrideJson = ref("{}");

    const interfaceTasks = ref<InterfaceTaskCandidate[]>([]);
    const selectedInterfaceTaskName = ref("");
    const selectedOptionCases = ref<InterfaceTaskOptionSelection[]>([]);

    const selectedInterfaceTask = computed(
      () =>
        interfaceTasks.value.find(
          (task) => task.name === selectedInterfaceTaskName.value,
        ) ?? null,
    );

    const selectedTaskOptionDefs = computed(
      () => selectedInterfaceTask.value?.option_defs ?? [],
    );

    const selectedTaskOptionSelections = computed(() => {
      const selectedMap = new Map(
        selectedOptionCases.value.map((item) => [
          item.optionName,
          item.caseName,
        ]),
      );

      return selectedTaskOptionDefs.value.map((optionDef) => ({
        optionName: optionDef.name,
        caseName:
          selectedMap.get(optionDef.name) ?? getDefaultCaseName(optionDef),
      }));
    });

    const derivedInterfaceOverride = computed<Record<string, unknown>>(() => {
      let merged: Record<string, unknown> = {};

      for (const selection of selectedTaskOptionSelections.value) {
        const optionDef = selectedTaskOptionDefs.value.find(
          (item) => item.name === selection.optionName,
        );
        const selectedCase = getOptionCases(optionDef).find(
          (item) => item.name === selection.caseName,
        );

        if (!selectedCase?.pipeline_override) {
          continue;
        }

        merged = deepMergeOverride(
          merged,
          selectedCase.pipeline_override as Record<string, unknown>,
        );
      }

      return merged;
    });

    const effectiveOverrideObject = computed<Record<string, unknown>>(() => {
      const manualPatch = parseStoredJsonObject(manualOverrideJson.value);
      return deepMergeOverride(derivedInterfaceOverride.value, manualPatch);
    });

    function syncOverrideJson() {
      overrideJson.value = JSON.stringify(
        effectiveOverrideObject.value,
        null,
        2,
      );
    }

    function setManualOverrideJson(value: string) {
      manualOverrideJson.value = value;
      syncOverrideJson();
    }

    function setSelectedOptionCase(optionName: string, caseName: string) {
      const next = selectedOptionCases.value.filter(
        (item) => item.optionName !== optionName,
      );
      next.push({ optionName, caseName });
      selectedOptionCases.value = next;
      syncOverrideJson();
    }

    function rebuildSelections(task: InterfaceTaskCandidate | null) {
      const nextSelections: InterfaceTaskOptionSelection[] = [];
      for (const optionDef of task?.option_defs ?? []) {
        const previous = selectedOptionCases.value.find(
          (item) => item.optionName === optionDef.name,
        );
        const allowedCaseNames = getOptionCases(optionDef).map(
          (item) => item.name,
        );
        const fallbackCaseName = getDefaultCaseName(optionDef);
        const caseName =
          previous && allowedCaseNames.includes(previous.caseName)
            ? previous.caseName
            : fallbackCaseName;
        if (caseName) {
          nextSelections.push({ optionName: optionDef.name, caseName });
        }
      }
      selectedOptionCases.value = nextSelections;
    }

    function applyInterfaceTasks(tasks: InterfaceTaskCandidate[]) {
      interfaceTasks.value = cloneTasks(tasks);

      const matchedTask = interfaceTasks.value.find(
        (task) => task.name === selectedInterfaceTaskName.value,
      );
      const activeTask = matchedTask ?? interfaceTasks.value[0] ?? null;

      selectedInterfaceTaskName.value = activeTask?.name ?? "";
      if (activeTask?.entry) {
        selectedEntry.value = activeTask.entry;
      }

      rebuildSelections(activeTask);
      syncOverrideJson();
    }

    function selectInterfaceTask(taskName: string) {
      selectedInterfaceTaskName.value = taskName;
      const task = selectedInterfaceTask.value;
      if (task?.entry) {
        selectedEntry.value = task.entry;
      }
      rebuildSelections(task);
      syncOverrideJson();
    }

    function clearInterfaceTasks() {
      interfaceTasks.value = [];
      selectedInterfaceTaskName.value = "";
      selectedOptionCases.value = [];
      syncOverrideJson();
    }

    function onRestore() {
      interfaceTasks.value = cloneTasks(interfaceTasks.value);
      syncOverrideJson();
    }

    return {
      selectedEntry,
      overrideJson,
      manualOverrideJson,
      interfaceTasks,
      selectedInterfaceTaskName,
      selectedOptionCases,
      selectedInterfaceTask,
      selectedTaskOptionDefs,
      selectedTaskOptionSelections,
      derivedInterfaceOverride,
      effectiveOverrideObject,
      setManualOverrideJson,
      setSelectedOptionCase,
      applyInterfaceTasks,
      selectInterfaceTask,
      clearInterfaceTasks,
      syncOverrideJson,
      onRestore,
    };
  },
  { persist: true },
);
