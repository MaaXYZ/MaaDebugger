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

interface ShortcutDef {
  /** Display label for the action */
  label: string;
  /** Default binding */
  defaultBinding: ShortcutBinding;
  /** Current binding */
  binding: ShortcutBinding;
}

const DEFAULT_SHORTCUTS: Record<
  ShortcutAction,
  Omit<ShortcutDef, "binding">
> = {
  "task.startStop": {
    label: "Start / Stop Task",
    defaultBinding: "P",
  },
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
    // Initialize shortcuts from defaults
    const shortcuts = ref<Record<ShortcutAction, ShortcutDef>>(
      Object.fromEntries(
        Object.entries(DEFAULT_SHORTCUTS).map(([action, def]) => [
          action,
          { ...def, binding: def.defaultBinding },
        ]),
      ) as Record<ShortcutAction, ShortcutDef>,
    );

    /** Get the current binding for an action */
    function getBinding(action: ShortcutAction): ShortcutBinding {
      return shortcuts.value[action]?.binding ?? null;
    }

    /** Get shortcut definition for an action */
    function getShortcut(action: ShortcutAction): ShortcutDef | undefined {
      return shortcuts.value[action];
    }

    /** Set a new binding for an action. Pass `null` to unbind. */
    function setBinding(action: ShortcutAction, binding: ShortcutBinding) {
      if (shortcuts.value[action]) {
        shortcuts.value[action].binding = binding;
      }
    }

    /** Reset a single action to its default binding */
    function resetBinding(action: ShortcutAction) {
      const def = DEFAULT_SHORTCUTS[action];
      if (def && shortcuts.value[action]) {
        shortcuts.value[action].binding = def.defaultBinding;
      }
    }

    /** Reset all shortcuts to defaults */
    function resetAll() {
      for (const [action, def] of Object.entries(DEFAULT_SHORTCUTS)) {
        if (shortcuts.value[action as ShortcutAction]) {
          shortcuts.value[action as ShortcutAction].binding =
            def.defaultBinding;
        }
      }
    }

    /** All shortcut actions as a list */
    const allShortcuts = computed(() => {
      return Object.entries(shortcuts.value).map(([action, def]) => ({
        action: action as ShortcutAction,
        ...def,
      }));
    });

    /** Check if an event matches a given action's binding */
    function matches(event: KeyboardEvent, action: ShortcutAction): boolean {
      return matchesShortcut(event, getBinding(action));
    }

    return {
      shortcuts,
      getBinding,
      getShortcut,
      setBinding,
      resetBinding,
      resetAll,
      allShortcuts,
      matches,
    };
  },
  { persist: true },
);
