import { computed, ref, watch } from "vue";
import { DoPipelineCheck, getTaskNodes, runTask, stopTask } from "@/api/http";
import { useShortcutsStore, formatShortcut } from "@/stores/shortcuts";
import { useStatusStore } from "@/stores/status";
import { useTaskStore } from "@/stores/task";
import { useAgentStore } from "@/stores/agent";
import { useSignalStore } from "@/stores/signal";
import { useDebugSettingsStore } from "@/stores/debugSettings";
import type { TaskStatus } from "./types";
import useResourceControl from "../useResourceControl";

interface TaskEntry {
  label: string;
  value: string;
}

interface ToastApi {
  add: (options: {
    id?: string;
    title: string;
    description?: string;
    icon?: string;
    color?: "error" | "success" | "warning" | "primary" | "neutral";
    actions?: any[];
  }) => void;
}

function fuzzyMatch(text: string, query: string): boolean {
  let ti = 0;
  for (let qi = 0; qi < query.length; qi++) {
    const idx = text.indexOf(query[qi]!, ti);
    if (idx < 0) return false;
    ti = idx + 1;
  }
  return true;
}

export default function useTaskControls(toast: ToastApi) {
  const shortcutsStore = useShortcutsStore();
  const statusStore = useStatusStore();
  const taskStore = useTaskStore();
  const agentStore = useAgentStore();
  const signalStore = useSignalStore();
  const debugSettingsStore = useDebugSettingsStore();
  const { tryLoadResource, openPipelineCheckDetails } = useResourceControl();

  const TASK_TOAST_ID = "task-toast";

  const entries = ref<TaskEntry[]>([]);
  const entrySearchTerm = ref("");
  const isStopping = ref(false);
  const overrideEditorOpen = ref(false);

  const selectedEntry = computed({
    get: () => taskStore.selectedEntry,
    set: (v: string) => {
      taskStore.selectedEntry = v;
    },
  });

  const taskStatus = computed<TaskStatus>(() => statusStore.taskStatus);
  const isRunning = computed(() => taskStatus.value === "running");
  const canStart = computed(
    () =>
      !isRunning.value &&
      statusStore.controllerStatus === "connected" &&
      statusStore.resourceStatus === "loaded" &&
      !agentStore.hasConnecting &&
      !!taskStore.effectiveEntry,
  );
  const startStopKeys = computed(() =>
    formatShortcut(shortcutsStore.getBinding("task.startStop")),
  );

  const interfaceTaskItems = computed(() =>
    taskStore.interfaceTasks.map((task) => ({
      label: taskStore.getDisplayName(task.name, task.label),
      value: task.name,
      name: task.name,
    })),
  );
  const hasInterfaceTasks = computed(() => interfaceTaskItems.value.length > 0);
  const selectedInterfaceTask = computed(() => taskStore.selectedInterfaceTask);
  const selectedTaskOptionDefs = computed(
    () => taskStore.selectedTaskOptionDefs,
  );
  const selectedTaskOptionSelections = computed(
    () => taskStore.selectedTaskOptionSelections,
  );
  const selectedTaskInputSelections = computed(
    () => taskStore.selectedTaskInputSelections,
  );
  const taskLaunchMode = computed({
    get: () => taskStore.taskLaunchMode,
    set: (value: "manual" | "interface") => {
      taskStore.setTaskLaunchMode(value);
    },
  });
  const usingInterfaceTask = computed(() => taskStore.usingInterfaceTask);
  const effectiveEntry = computed(() => taskStore.effectiveEntry);

  const entrySelectItems = computed(() => {
    const all = entries.value.map((e) => ({ label: e.label, value: e.value }));
    const q = entrySearchTerm.value.toLowerCase();
    if (!q) return all;

    const startsWith: typeof all = [];
    const endsWith: typeof all = [];
    const contains: typeof all = [];
    const fuzzy: typeof all = [];

    for (const item of all) {
      const label = item.label.toLowerCase();
      if (label.startsWith(q)) startsWith.push(item);
      else if (label.endsWith(q)) endsWith.push(item);
      else if (label.includes(q)) contains.push(item);
      else if (fuzzyMatch(label, q)) fuzzy.push(item);
    }

    return [...startsWith, ...endsWith, ...contains, ...fuzzy];
  });

  const cjkRegex =
    /[\u2E80-\u2FFF\u3000-\u303F\u3040-\u309F\u30A0-\u30FF\u3100-\u312F\u3200-\u32FF\u3400-\u4DBF\u4E00-\u9FFF\uF900-\uFAFF\uFE30-\uFE4F\uFF00-\uFFEF]/;

  const entryContentMinWidth = computed(() => {
    const items = entrySelectItems.value;
    if (items.length === 0) return "0px";

    let longestWidth = 0;
    let bestCjk = 0;
    let bestAscii = 0;
    for (const item of items) {
      let w = 0;
      let cjk = 0;
      let ascii = 0;
      for (const ch of item.label) {
        if (cjkRegex.test(ch)) {
          w += 2;
          cjk++;
        } else {
          w += 1;
          ascii++;
        }
      }
      if (w > longestWidth) {
        longestWidth = w;
        bestCjk = cjk;
        bestAscii = ascii;
      }
    }

    const padding = 3;
    const parts: string[] = [];
    if (bestCjk > 0) parts.push(`${bestCjk}em`);
    parts.push(`${bestAscii + padding}ch`);

    return `calc(${parts.join(" + ")})`;
  });

  watch(
    () => signalStore.refreshNode,
    (status) => {
      if (status > 0) void refreshNodes();
    },
  );

  function selectInterfaceTask(taskName: string) {
    taskStore.selectInterfaceTask(taskName);
  }

  function setInterfaceOptionCase(optionName: string, caseName: string) {
    taskStore.setSelectedOptionCase(optionName, caseName);
  }

  function setInterfaceInputValue(optionName: string, value: string) {
    taskStore.setSelectedInputValue(optionName, value);
  }

  function setOverrideJson(value: string) {
    taskStore.setOverrideJson(value);
  }

  function setManualOverrideJson(value: string) {
    taskStore.setManualOverrideJson(value);
  }

  async function onStart() {
    if (!canStart.value) return;

    if (agentStore.hasConnecting) {
      toast.add({
        id: TASK_TOAST_ID,
        title: "Agent Connecting",
        description: "Please wait for agent connection to finish",
        icon: "i-lucide-loader",
        color: "warning",
      });
      return;
    }

    const loadResult = await tryLoadResource();
    if (!loadResult.success) {
      toast.add({
        id: TASK_TOAST_ID,
        title: "Resource Load Failed",
        description: loadResult.msg,
        icon: "i-lucide-circle-x",
        color: "error",
      });
      return;
    }

    if (debugSettingsStore.checkPipeline) {
      const pipelineCheck = await DoPipelineCheck();
      if (!pipelineCheck.succeed) {
        toast.add({
          id: TASK_TOAST_ID,
          title: "Pipeline Check Failed",
          description: pipelineCheck.msg,
          icon: "i-lucide-circle-x",
          color: "error",
        });
      }

      const issues = Array.isArray(pipelineCheck.data)
        ? pipelineCheck.data
        : [];
      const errorCount = issues.filter((item) => item.level === "error").length;
      const warningCount = issues.filter(
        (item) => item.level === "warning",
      ).length;

      if (debugSettingsStore.preventRunning && errorCount > 0) {
        toast.add({
          id: TASK_TOAST_ID,
          title: "Task Blocked",
          description: `Pipeline has ${errorCount} error(s). Fix errors or disable Prevent Running to continue.`,
          icon: "i-lucide-octagon-x",
          color: "error",
          actions: [
            {
              icon: "i-lucide-info",
              label: "View Details",
              color: "neutral",
              variant: "outline",
              onClick: () => {
                openPipelineCheckDetails();
              },
            },
          ],
        });
        return;
      }

      const notifyLevel = debugSettingsStore.checkPipelineNotifyLevel;
      const shouldNotify =
        notifyLevel === "WARNING"
          ? errorCount + warningCount > 0
          : notifyLevel === "ERROR"
            ? errorCount > 0
            : false;

      if (shouldNotify) {
        toast.add({
          id: TASK_TOAST_ID,
          title: "Pipeline Issues Found",
          description: `Found ${errorCount} error(s) and ${warningCount} warning(s).`,
          icon: "i-lucide-alert-triangle",
          color: errorCount > 0 ? "error" : "warning",
        });
      }
    }

    const result = await runTask(
      effectiveEntry.value,
      taskStore.effectiveOverrideObject,
    );
    if (!result.succeed) {
      toast.add({
        id: TASK_TOAST_ID,
        title: "Task Run Failed",
        description: result.msg,
        icon: "i-lucide-circle-x",
        color: "error",
      });
    }
  }

  async function onStop() {
    if (isStopping.value) return;
    isStopping.value = true;
    try {
      const result = await stopTask();
      if (!result.succeed) {
        toast.add({
          id: TASK_TOAST_ID,
          title: "Task Stop Failed",
          description: result.msg,
          icon: "i-lucide-circle-x",
          color: "error",
        });
      } else {
        toast.add({
          id: TASK_TOAST_ID,
          title: "Task Stop Requested",
          icon: "i-lucide-circle-stop",
          color: "warning",
        });
      }
    } finally {
      isStopping.value = false;
    }
  }

  async function refreshNodes() {
    const nodes = await getTaskNodes();
    entries.value = nodes.map((n) => ({ label: n, value: n }));
    const stillExists = entries.value.some(
      (e) => e.value === selectedEntry.value,
    );
    if (!stillExists && entries.value.length > 0) {
      selectedEntry.value = entries.value[0]!.value;
    }
  }

  function onKeydown(e: KeyboardEvent) {
    const target = e.target as HTMLElement | null;
    const tag = target?.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
    if (target?.isContentEditable) return;

    if (shortcutsStore.matches(e, "task.startStop")) {
      e.preventDefault();
      if (isRunning.value) {
        void onStop();
      } else {
        void onStart();
      }
    }
  }

  function mount() {
    void refreshNodes();
    window.addEventListener("keydown", onKeydown);
  }

  function unmount() {
    window.removeEventListener("keydown", onKeydown);
  }

  return {
    entries,
    selectedEntry,
    taskStatus,
    entrySearchTerm,
    entrySelectItems,
    entryContentMinWidth,
    isRunning,
    canStart,
    isStopping,
    startStopKeys,
    overrideEditorOpen,
    taskStore,
    hasInterfaceTasks,
    interfaceTaskItems,
    selectedInterfaceTask,
    selectedTaskOptionDefs,
    selectedTaskOptionSelections,
    selectedTaskInputSelections,
    taskLaunchMode,
    usingInterfaceTask,
    effectiveEntry,
    selectInterfaceTask,
    setInterfaceOptionCase,
    setInterfaceInputValue,
    setOverrideJson,
    setManualOverrideJson,
    onStart,
    onStop,
    refreshNodes,
    mount,
    unmount,
  };
}
