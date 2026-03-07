<template>
    <div class="monaco-editor-wrapper relative">
        <div ref="containerRef" class="monaco-editor-container" :style="containerStyle"></div>
        <UButton :icon="copied ? 'i-lucide-check' : 'i-lucide-copy'" size="xs" variant="ghost"
                 :color="copied ? 'success' : 'neutral'" class="absolute top-1 right-1 z-10 opacity-60 hover:opacity-100"
                 @click="copyContent" />
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, computed, type CSSProperties } from 'vue'
import { ensureMonacoReady, monaco } from './setup'
import { useEditorSettingsStore } from '@/stores/editorSettings'
import type { editor as MonacoEditor, IDisposable } from 'monaco-editor'

export interface MonacoEditorProps {
    /** The text content to display/edit */
    modelValue?: string
    /** Editor language (default: 'jsonc') */
    language?: string
    /** Read-only mode (default: true) */
    readOnly?: boolean
    /**
     * Editor theme. Built-in: 'vs' (light), 'vs-dark', 'hc-black', 'hc-light'.
     * Pass a custom theme name after registering it via `monaco.editor.defineTheme()`.
     * Default: 'auto' — automatically switches between 'vs' and 'vs-dark' based on dark mode.
     */
    theme?: string
    /** Minimum height in px (default: 100) */
    minHeight?: number
    /** Maximum height in px (default: 400) */
    maxHeight?: number
    /** Additional monaco editor options */
    options?: MonacoEditor.IStandaloneEditorConstructionOptions
}

const props = withDefaults(defineProps<MonacoEditorProps>(), {
    modelValue: '',
    language: 'json',
    readOnly: true,
    theme: 'auto',
    minHeight: 100,
    maxHeight: 400,
    options: () => ({}),
})

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const editorSettingsStore = useEditorSettingsStore()
const containerRef = ref<HTMLDivElement>()
const ready = ref(false)
let editorInstance: MonacoEditor.IStandaloneCodeEditor | null = null
let darkModeObserver: MutationObserver | null = null
let suggestTriggerListener: IDisposable | null = null
let mounted = true

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | null = null
let readyResolve: ((editor: MonacoEditor.IStandaloneCodeEditor) => void) | null = null
const readyPromise = new Promise<MonacoEditor.IStandaloneCodeEditor>((resolve) => {
    readyResolve = resolve
})

function shouldTriggerSuggestForEmptyJsonString(editor: MonacoEditor.IStandaloneCodeEditor): boolean {
    const model = editor.getModel()
    const position = editor.getPosition()
    if (!model || !position || model.getLanguageId() !== 'json') {
        return false
    }

    const lineContent = model.getLineContent(position.lineNumber)
    const zeroBasedColumn = position.column - 1
    const previousChar = lineContent[zeroBasedColumn - 1] ?? ''
    const nextChar = lineContent[zeroBasedColumn] ?? ''

    return previousChar === '"' && nextChar === '"'
}

function triggerSuggestForEmptyJsonString(editor: MonacoEditor.IStandaloneCodeEditor) {
    suggestTriggerListener?.dispose()
    suggestTriggerListener = editor.onDidChangeCursorPosition(() => {
        if (!shouldTriggerSuggestForEmptyJsonString(editor)) {
            return
        }

        queueMicrotask(() => {
            editor.trigger('json-empty-string', 'editor.action.triggerSuggest', {})
        })
    })
}

async function copyContent() {
    const value = editorInstance?.getValue() ?? props.modelValue
    try {
        await navigator.clipboard.writeText(value)
        copied.value = true
        if (copyTimer) clearTimeout(copyTimer)
        copyTimer = setTimeout(() => {
            copied.value = false
        }, 2000)
    } catch {
        // fallback: do nothing
    }
}

const effectiveFontSize = computed(() => editorSettingsStore.normalizedFontSize)
const effectiveMinHeight = computed(() => Math.max(props.minHeight, editorSettingsStore.normalizedMinHeight))
const effectiveMaxHeight = computed(() => Math.max(props.maxHeight, editorSettingsStore.normalizedMaxHeight, effectiveMinHeight.value))

const containerStyle = computed<CSSProperties>(() => ({
    minHeight: `${effectiveMinHeight.value}px`,
    maxHeight: `${effectiveMaxHeight.value}px`,
}))

/**
 * Compute the content height and resize the editor to fit,
 * clamped between minHeight and maxHeight.
 */
function updateEditorHeight() {
    if (!editorInstance || !containerRef.value) return
    const contentHeight = editorInstance.getContentHeight()
    const clampedHeight = Math.min(Math.max(contentHeight, effectiveMinHeight.value), effectiveMaxHeight.value)
    containerRef.value.style.height = `${clampedHeight}px`
    editorInstance.layout()
}

function isDarkMode(): boolean {
    return document.documentElement.classList.contains('dark')
}

/** Resolve the effective theme name based on the `theme` prop */
function resolveTheme(): string {
    if (props.theme === 'auto') {
        return isDarkMode() ? 'vs-dark' : 'vs'
    }
    return props.theme
}

onMounted(async () => {
    if (!containerRef.value) return

    await ensureMonacoReady()
    if (!mounted || !containerRef.value) return

    editorInstance = monaco.editor.create(containerRef.value, {
        value: props.modelValue,
        language: props.language,
        readOnly: props.readOnly,
        theme: resolveTheme(),
        automaticLayout: true,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        lineNumbers: 'on',
        fontSize: effectiveFontSize.value,
        tabSize: 2,
        wordWrap: 'on',
        folding: true,
        quickSuggestions: {
            other: true,
            comments: false,
            strings: true,
        },
        suggestOnTriggerCharacters: true,
        acceptSuggestionOnCommitCharacter: true,
        renderLineHighlight: props.readOnly ? 'none' : 'line',
        overviewRulerLanes: 0,
        hideCursorInOverviewRuler: true,
        overviewRulerBorder: false,
        scrollbar: {
            vertical: 'auto',
            horizontal: 'auto',
            verticalScrollbarSize: 8,
            horizontalScrollbarSize: 8,
        },
        padding: { top: 8, bottom: 8 },
        domReadOnly: props.readOnly,
        contextmenu: !props.readOnly,
        ...props.options,
    })

    ready.value = true
    readyResolve?.(editorInstance)
    readyResolve = null

    triggerSuggestForEmptyJsonString(editorInstance)

    // Auto-resize based on content
    editorInstance.onDidContentSizeChange(() => {
        updateEditorHeight()
    })
    updateEditorHeight()

    // Emit changes when not read-only
    if (!props.readOnly) {
        editorInstance.onDidChangeModelContent(() => {
            const value = editorInstance?.getValue() ?? ''
            emit('update:modelValue', value)
        })
    }

    // Watch for dark mode changes (only in 'auto' theme mode)
    if (props.theme === 'auto') {
        darkModeObserver = new MutationObserver(() => {
            monaco.editor.setTheme(resolveTheme())
        })
        darkModeObserver.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class'],
        })
    }
})

// Watch modelValue changes from parent
watch(
    () => props.modelValue,
    (newValue) => {
        if (!editorInstance) return
        const currentValue = editorInstance.getValue()
        if (newValue !== currentValue) {
            editorInstance.setValue(newValue)
        }
    },
)

// Watch readOnly changes
watch(
    () => props.readOnly,
    (newReadOnly) => {
        if (!editorInstance) return
        editorInstance.updateOptions({
            readOnly: newReadOnly,
            domReadOnly: newReadOnly,
            renderLineHighlight: newReadOnly ? 'none' : 'line',
            contextmenu: !newReadOnly,
        })
    },
)

// Watch language changes
watch(
    () => props.language,
    (newLanguage) => {
        if (!editorInstance) return
        const model = editorInstance.getModel()
        if (model) {
            monaco.editor.setModelLanguage(model, newLanguage)
        }
    },
)

// Watch theme changes
watch(
    () => props.theme,
    () => {
        monaco.editor.setTheme(resolveTheme())
    },
)

watch(effectiveFontSize, (fontSize) => {
    editorInstance?.updateOptions({ fontSize })
    updateEditorHeight()
})

watch([effectiveMinHeight, effectiveMaxHeight], () => {
    updateEditorHeight()
})

onBeforeUnmount(() => {
    mounted = false
    if (copyTimer) clearTimeout(copyTimer)
    darkModeObserver?.disconnect()
    darkModeObserver = null
    suggestTriggerListener?.dispose()
    suggestTriggerListener = null
    editorInstance?.dispose()
    editorInstance = null
    ready.value = false
})

defineExpose({
    /** Get the underlying monaco editor instance */
    getEditor: () => editorInstance,
    ready,
    whenReady: () => readyPromise,
})
</script>

<style scoped>
.monaco-editor-container {
    width: 100%;
    border-radius: 0.375rem;
    overflow: hidden;
}
</style>
