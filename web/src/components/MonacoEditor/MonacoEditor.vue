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

const MONACO_VSCODE_DARK_THEME = 'maa-vscode-dark'
const MONACO_VSCODE_LIGHT_THEME = 'maa-vscode-light'
let monacoThemeRegistered = false

function ensureMonacoVsCodeTheme() {
    if (monacoThemeRegistered) {
        return
    }

    monaco.editor.defineTheme(MONACO_VSCODE_DARK_THEME, {
        base: 'vs-dark',
        inherit: true,
        rules: [
            { token: 'string.key.json', foreground: '9CDCFE' },
            { token: 'string.value.json', foreground: 'CE9178' },
            { token: 'string', foreground: 'CE9178' },
            { token: 'number', foreground: 'B5CEA8' },
            { token: 'number.json', foreground: 'B5CEA8' },
            { token: 'keyword.json', foreground: '569CD6' },
            { token: 'boolean', foreground: '569CD6' },
            { token: 'keyword', foreground: '569CD6' },
            { token: 'comment', foreground: '6A9955' },
        ],
        colors: {
            'editor.background': '#1E1E1E',
            'editor.foreground': '#D4D4D4',
            'editorLineNumber.foreground': '#858585',
            'editorLineNumber.activeForeground': '#C6C6C6',
            'editor.selectionBackground': '#264F78',
            'editor.inactiveSelectionBackground': '#3A3D41',
            'editor.lineHighlightBackground': '#2A2D2E',
            'editorCursor.foreground': '#AEAFAD',
            'editorWhitespace.foreground': '#3B3B3B',
        },
    })

    monaco.editor.defineTheme(MONACO_VSCODE_LIGHT_THEME, {
        base: 'vs',
        inherit: true,
        rules: [
            { token: 'string.key.json', foreground: '0451A5' },
            { token: 'string.value.json', foreground: 'A31515' },
            { token: 'string', foreground: 'A31515' },
            { token: 'number', foreground: '098658' },
            { token: 'number.json', foreground: '098658' },
            { token: 'keyword.json', foreground: '0000FF' },
            { token: 'boolean', foreground: '0000FF' },
            { token: 'keyword', foreground: '0000FF' },
            { token: 'comment', foreground: '008000' },
        ],
        colors: {
            'editor.background': '#FFFFFF',
            'editor.foreground': '#000000',
            'editorLineNumber.foreground': '#237893',
            'editorLineNumber.activeForeground': '#0B216F',
            'editor.selectionBackground': '#ADD6FF',
            'editor.inactiveSelectionBackground': '#E5EBF1',
            'editor.lineHighlightBackground': '#F7F7F7',
            'editorCursor.foreground': '#000000',
            'editorWhitespace.foreground': '#D3D3D3',
        },
    })

    monacoThemeRegistered = true
}
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
     * Reserved for backward compatibility. The editor always uses the VS Code style theme.
     */
    theme?: string
    /** Minimum height in px (default: 100) */
    minHeight?: number
    /** Maximum height in px (default: 400) */
    maxHeight?: number
    /** Whether to apply global editor height settings from the store */
    useGlobalHeightSettings?: boolean
    /** Additional monaco editor options */
    options?: MonacoEditor.IStandaloneEditorConstructionOptions
}

const props = withDefaults(defineProps<MonacoEditorProps>(), {
    modelValue: '',
    language: 'json',
    readOnly: true,
    theme: MONACO_VSCODE_DARK_THEME,
    minHeight: 100,
    maxHeight: 400,
    useGlobalHeightSettings: true,
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
const effectiveMinHeight = computed(() => {
    if (!props.useGlobalHeightSettings) {
        return props.minHeight
    }
    return Math.max(props.minHeight, editorSettingsStore.normalizedMinHeight)
})
const effectiveMaxHeight = computed(() => {
    if (!props.useGlobalHeightSettings) {
        return Math.max(props.maxHeight, effectiveMinHeight.value)
    }
    return Math.max(props.maxHeight, editorSettingsStore.normalizedMaxHeight, effectiveMinHeight.value)
})

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

/** Monaco always uses the VS Code style theme, and follows light/dark mode automatically. */
function resolveTheme(): string {
    return isDarkMode() ? MONACO_VSCODE_DARK_THEME : MONACO_VSCODE_LIGHT_THEME
}

onMounted(async () => {
    if (!containerRef.value) return

    await ensureMonacoReady()
    if (!mounted || !containerRef.value) return

    ensureMonacoVsCodeTheme()

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

    monaco.editor.setTheme(resolveTheme())

    darkModeObserver = new MutationObserver(() => {
        monaco.editor.setTheme(resolveTheme())
    })
    darkModeObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
    })
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

// Keep backward compatibility if older callers still pass the theme prop.
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
