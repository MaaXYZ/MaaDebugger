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

function cloneLocaleValues(
  values: Record<string, Record<string, string>>,
): Record<string, Record<string, string>> {
  return JSON.parse(JSON.stringify(values)) as Record<
    string,
    Record<string, string>
  >;
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

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

function buildManualOverridePatch(
  base: Record<string, unknown>,
  target: Record<string, unknown>,
): Record<string, unknown> {
  const patch: Record<string, unknown> = {};

  for (const [key, rawTargetValue] of Object.entries(target)) {
    const targetValue = normalizeOverrideValue(rawTargetValue);
    const baseValue = base[key];

    if (isPlainObject(baseValue) && isPlainObject(targetValue)) {
      const nestedPatch = buildManualOverridePatch(baseValue, targetValue);
      if (Object.keys(nestedPatch).length > 0) {
        patch[key] = nestedPatch;
      }
      continue;
    }

    if (JSON.stringify(baseValue) !== JSON.stringify(targetValue)) {
      patch[key] = cloneObject(targetValue);
    }
  }

  return patch;
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
    const taskLaunchMode = ref<"manual" | "interface">("manual");
    /** Pipeline override JSON string (derived result + manual patch), persisted */
    const overrideJson = ref("{}");
    /** User-editable extra patch merged on top of interface-derived override */
    const manualOverrideJson = ref("{}");

    const interfaceTasks = ref<InterfaceTaskCandidate[]>([]);
    const interfaceLanguages = ref<Record<string, string>>({});
    const interfaceLocaleValues = ref<Record<string, Record<string, string>>>(
      {},
    );
    const selectedInterfaceLanguage = ref("");
    const selectedInterfaceTaskName = ref("");
    const selectedOptionCases = ref<InterfaceTaskOptionSelection[]>([]);

    const availableInterfaceLanguages = computed(() =>
      Object.entries(interfaceLanguages.value).map(([value, path]) => ({
        value,
        label: value,
        path,
      })),
    );

    const hasInterfaceLanguages = computed(
      () => availableInterfaceLanguages.value.length > 0,
    );

    const activeLocaleMap = computed<Record<string, string>>(
      () => interfaceLocaleValues.value[selectedInterfaceLanguage.value] ?? {},
    );

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

    const usingInterfaceTask = computed(
      () =>
        taskLaunchMode.value === "interface" &&
        interfaceTasks.value.length > 0 &&
        !!selectedInterfaceTask.value,
    );

    const effectiveEntry = computed(() => {
      if (usingInterfaceTask.value && selectedInterfaceTask.value?.entry) {
        return selectedInterfaceTask.value.entry;
      }
      return selectedEntry.value;
    });

    const effectiveBaseOverride = computed<Record<string, unknown>>(() =>
      usingInterfaceTask.value ? derivedInterfaceOverride.value : {},
    );

    const effectiveOverrideObject = computed<Record<string, unknown>>(() => {
      const manualPatch = parseStoredJsonObject(manualOverrideJson.value);
      return deepMergeOverride(effectiveBaseOverride.value, manualPatch);
    });

    function syncOverrideJson() {
      overrideJson.value = JSON.stringify(
        effectiveOverrideObject.value,
        null,
        2,
      );
    }

    function setOverrideJson(value: string) {
      overrideJson.value = value;
      const parsedOverride = parseStoredJsonObject(value);
      const manualPatch = buildManualOverridePatch(
        effectiveBaseOverride.value,
        parsedOverride,
      );
      manualOverrideJson.value = JSON.stringify(manualPatch, null, 2);
      syncOverrideJson();
    }

    function setManualOverrideJson(value: string) {
      manualOverrideJson.value = value;
      syncOverrideJson();
    }

    function setTaskLaunchMode(value: "manual" | "interface") {
      taskLaunchMode.value = value;
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

    function setInterfaceLanguage(language: string) {
      if (!language || !interfaceLanguages.value[language]) {
        const fallback = availableInterfaceLanguages.value[0]?.value ?? "";
        selectedInterfaceLanguage.value = fallback;
        return;
      }
      selectedInterfaceLanguage.value = language;
    }

    function resolveInterfaceText(value?: string | null): string {
      const trimmed = (value ?? "").trim();
      if (!trimmed) return "";
      if (!trimmed.startsWith("$")) return trimmed;
      const localeKey = trimmed.slice(1);
      return activeLocaleMap.value[localeKey] ?? trimmed;
    }

    function getDisplayName(
      name?: string | null,
      label?: string | null,
    ): string {
      const fallbackName = (name ?? "").trim();
      const resolvedLabel = resolveInterfaceText(label);
      if (resolvedLabel && resolvedLabel !== (label ?? "").trim()) {
        return resolvedLabel;
      }
      return fallbackName;
    }

    function applyInterfaceTasks(
      tasks: InterfaceTaskCandidate[],
      options?: {
        languages?: Record<string, string>;
        localeValues?: Record<string, Record<string, string>>;
      },
    ) {
      interfaceTasks.value = cloneTasks(tasks);
      interfaceLanguages.value = { ...(options?.languages ?? {}) };
      interfaceLocaleValues.value = cloneLocaleValues(
        options?.localeValues ?? {},
      );
      setInterfaceLanguage(selectedInterfaceLanguage.value);

      const matchedTask = interfaceTasks.value.find(
        (task) => task.name === selectedInterfaceTaskName.value,
      );
      const activeTask = matchedTask ?? interfaceTasks.value[0] ?? null;

      selectedInterfaceTaskName.value = activeTask?.name ?? "";
      rebuildSelections(activeTask);
      syncOverrideJson();
    }

    function selectInterfaceTask(taskName: string) {
      selectedInterfaceTaskName.value = taskName;
      const task = selectedInterfaceTask.value;
      rebuildSelections(task);
      syncOverrideJson();
    }

    function clearInterfaceTasks() {
      interfaceTasks.value = [];
      interfaceLanguages.value = {};
      interfaceLocaleValues.value = {};
      selectedInterfaceLanguage.value = "";
      selectedInterfaceTaskName.value = "";
      selectedOptionCases.value = [];
      if (taskLaunchMode.value === "interface") {
        taskLaunchMode.value = "manual";
      }
      syncOverrideJson();
    }

    function onRestore() {
      interfaceTasks.value = cloneTasks(interfaceTasks.value);
      interfaceLocaleValues.value = cloneLocaleValues(
        interfaceLocaleValues.value,
      );
      syncOverrideJson();
    }

    return {
      selectedEntry,
      taskLaunchMode,
      overrideJson,
      manualOverrideJson,
      interfaceTasks,
      interfaceLanguages,
      interfaceLocaleValues,
      selectedInterfaceLanguage,
      availableInterfaceLanguages,
      hasInterfaceLanguages,
      selectedInterfaceTaskName,
      selectedOptionCases,
      selectedInterfaceTask,
      selectedTaskOptionDefs,
      selectedTaskOptionSelections,
      activeLocaleMap,
      derivedInterfaceOverride,
      usingInterfaceTask,
      effectiveEntry,
      effectiveOverrideObject,
      setOverrideJson,
      setManualOverrideJson,
      setTaskLaunchMode,
      setSelectedOptionCase,
      setInterfaceLanguage,
      resolveInterfaceText,
      getDisplayName,
      applyInterfaceTasks,
      selectInterfaceTask,
      clearInterfaceTasks,
      syncOverrideJson,
      onRestore,
    };
  },
  { persist: true },
);
