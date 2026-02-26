import { computed, ref, watch } from "vue";
import { getTaskNodes, runTask, stopTask } from "@/api/http";
import { useShortcutsStore, formatShortcut } from "@/stores/shortcuts";
import { useStatusStore } from "@/stores/status";
import { useTaskStore } from "@/stores/task";
import type { TaskStatus } from "./types";

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

export function useTaskControls(toast: ToastApi) {
  const shortcutsStore = useShortcutsStore();
  const statusStore = useStatusStore();
  const taskStore = useTaskStore();

  const entries = ref<TaskEntry[]>([]);
  const entrySearchTerm = ref("");
  const isStopping = ref(false);

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
      statusStore.controllerStatus === "connected" &&
      statusStore.resourceStatus === "loaded" &&
      !!selectedEntry.value,
  );
  const startStopKeys = computed(() =>
    formatShortcut(shortcutsStore.getBinding("task.startStop")),
  );

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

  watch(
    () => statusStore.resourceStatus,
    (newStatus, oldStatus) => {
      if (oldStatus === "loading" && newStatus === "loaded") {
        void refreshNodes();
      }
    },
  );

  async function onStart() {
    if (!canStart.value) return;
    const result = await runTask(selectedEntry.value, {});
    if (!result.succeed) {
      toast.add({
        id: "task-toast",
        title: "Task Run Failed",
        description: result.msg,
        icon: "i-lucide-circle-x",
        color: "error",
      });
    } else {
      toast.add({
        id: "task-toast",
        title: "Task Started",
        icon: "i-lucide-play",
        color: "success",
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
          id: "task-toast",
          title: "Task Stop Failed",
          description: result.msg,
          icon: "i-lucide-circle-x",
          color: "error",
        });
      } else {
        toast.add({
          id: "task-toast",
          title: "Task Stop Requested",
          icon: "i-lucide-circle-stop",
          color: "warning",
        });
      }
    } finally {
      isStopping.value = false;
    }
  }

  function onEditOverride() {
    // TODO: Open editor for task override
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
      } else if (canStart.value) {
        void onStart();
      }
    }
  }

  function mount() {
    window.addEventListener("keydown", onKeydown);
    void refreshNodes();
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
    isRunning,
    canStart,
    isStopping,
    startStopKeys,
    onStart,
    onStop,
    onEditOverride,
    refreshNodes,
    mount,
    unmount,
  };
}
