import { defineStore } from "pinia";
import { ref, computed } from "vue";

/**
 * Represents a keyboard shortcut binding.
 * - `null` means no binding (disabled).
 * - A string like `"P"` means a single key.
 * - Modifiers are prefixed: `"Ctrl+P"`, `"Ctrl+Shift+P"`, `"Alt+S"`, etc.
 */
export type ShortcutBinding = string | null;

/** Known shortcut action identifiers */
export type ShortcutAction = "task.startStop";

/**
 * Default bindings per action.
 * Kept in code only — the persisted state is a plain `{ action: binding }` map,
 * so defaults are never written into the user's config.
 */
const DEFAULT_BINDINGS: Record<ShortcutAction, ShortcutBinding> = {
  "task.startStop": "P",
};

/**
 * Parse a shortcut string into its parts.
 * Example: "Ctrl+Shift+P" → { ctrl: true, shift: true, alt: false, meta: false, key: "p" }
 */
export function parseShortcut(shortcut: string): {
  ctrl: boolean;
  shift: boolean;
  alt: boolean;
  meta: boolean;
  key: string;
} {
  const parts = shortcut.split("+").map((p) => p.trim());
  const key = parts.pop()!.toLowerCase();
  const mods = new Set(parts.map((m) => m.toLowerCase()));

  return {
    ctrl: mods.has("ctrl") || mods.has("control"),
    shift: mods.has("shift"),
    alt: mods.has("alt"),
    meta: mods.has("meta") || mods.has("cmd") || mods.has("command"),
    key,
  };
}

/**
 * Check if a KeyboardEvent matches a shortcut binding string.
 */
export function matchesShortcut(
  event: KeyboardEvent,
  binding: ShortcutBinding,
): boolean {
  if (!binding) return false;

  const parsed = parseShortcut(binding);

  return (
    event.key.toLowerCase() === parsed.key &&
    event.ctrlKey === parsed.ctrl &&
    event.shiftKey === parsed.shift &&
    event.altKey === parsed.alt &&
    event.metaKey === parsed.meta
  );
}

/**
 * Format a shortcut binding for display.
 * Returns an array of key labels suitable for rendering with UKbd.
 * Example: "Ctrl+Shift+P" → ["Ctrl", "Shift", "P"]
 */
export function formatShortcut(binding: ShortcutBinding): string[] {
  if (!binding) return [];
  return binding.split("+").map((p) => p.trim());
}

/**
 * Convert a KeyboardEvent to a shortcut string.
 * Used when recording a new shortcut in settings.
 */
export function eventToShortcut(event: KeyboardEvent): string | null {
  const key = event.key;

  // Ignore pure modifier keys
  if (["Control", "Shift", "Alt", "Meta"].includes(key)) return null;

  const parts: string[] = [];
  if (event.ctrlKey) parts.push("Ctrl");
  if (event.shiftKey) parts.push("Shift");
  if (event.altKey) parts.push("Alt");
  if (event.metaKey) parts.push("Meta");

  // Normalize key display
  const displayKey = key.length === 1 ? key.toUpperCase() : key;
  parts.push(displayKey);

  return parts.join("+");
}

export const useShortcutsStore = defineStore(
  "shortcuts",
  () => {
    // Only the current bindings are persisted (plain map). Defaults live in code.
    const shortcuts = ref<Record<ShortcutAction, ShortcutBinding>>({
      ...DEFAULT_BINDINGS,
    });

    /** Get the current binding for an action */
    function getBinding(action: ShortcutAction): ShortcutBinding {
      return shortcuts.value[action] ?? null;
    }

    /** Set a new binding for an action. Pass `null` to unbind. */
    function setBinding(action: ShortcutAction, binding: ShortcutBinding) {
      shortcuts.value[action] = binding;
    }

    /** Reset a single action to its default binding */
    function resetBinding(action: ShortcutAction) {
      shortcuts.value[action] = DEFAULT_BINDINGS[action] ?? null;
    }

    /** Reset all shortcuts to defaults */
    function resetAll() {
      for (const action of Object.keys(DEFAULT_BINDINGS) as ShortcutAction[]) {
        shortcuts.value[action] = DEFAULT_BINDINGS[action] ?? null;
      }
    }

    /** All shortcut actions as a list (for the settings UI) */
    const allShortcuts = computed(() =>
      (Object.keys(shortcuts.value) as ShortcutAction[]).map((action) => ({
        action,
        binding: shortcuts.value[action],
      })),
    );

    /** Check if an event matches a given action's binding */
    function matches(event: KeyboardEvent, action: ShortcutAction): boolean {
      return matchesShortcut(event, getBinding(action));
    }

    /**
     * 旧配置迁移：早期版本持久化的是 { binding, defaultBinding, label }
     * 对象（label 从未被使用过），归一化为纯 binding 字符串/null。
     */
    function onRestore() {
      for (const action of Object.keys(shortcuts.value) as ShortcutAction[]) {
        const v: unknown = shortcuts.value[action];
        if (typeof v !== "string" && v !== null) {
          const old = v as { binding?: unknown };
          shortcuts.value[action] =
            typeof old.binding === "string" ? old.binding : null;
        }
      }
    }

    return {
      shortcuts,
      getBinding,
      setBinding,
      resetBinding,
      resetAll,
      allShortcuts,
      matches,
      onRestore,
    };
  },
  { persist: true },
);
