<template>
    <UModal v-model:open="open" :title="title" :description="description" :ui="{ content: 'sm:max-w-5xl' }">
        <template #body>
            <MonacoEditor ref="editorRef" v-model="draft" :language="language" :read-only="false" :min-height="420"
                :max-height="720" />
        </template>
        <template #footer>
            <div class="flex justify-end gap-2 w-full">
                <UButton variant="ghost" color="neutral" label="Cancel" @click="onCancel" />
                <UButton color="primary" label="Save" icon="i-lucide-save" @click="onSave" />
            </div>
        </template>
    </UModal>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue'
import { MonacoEditor, monaco, ensureMonacoReady, getJsonDiagnosticsOptions, setJsonDiagnosticsOptions, runMonacoJsonSession } from '@/components/MonacoEditor'
import type { editor as MonacoEditorNamespace, IDisposable } from 'monaco-editor'
import type { MonacoJsonSchemaEntry } from './setup'

const props = withDefaults(
    defineProps<{
        title?: string
        description?: string
        modelValue?: string
        language?: string
        /** 主 schema */
        schema?: Record<string, unknown>
        /** 外部 schema 映射：key 必须与 $ref 相对路径一致 */
        externalSchemas?: Record<string, Record<string, unknown>>
    }>(),
    {
        title: 'Edit JSON',
        description: '',
        modelValue: '',
        language: 'json',
        schema: undefined,
        externalSchemas: undefined,
    },
)

const emit = defineEmits<{
    'update:modelValue': [value: string]
}>()

const open = defineModel<boolean>('open', { default: false })

const toast = useToast()
const editorRef = ref<InstanceType<typeof MonacoEditor> | null>(null)
const draft = ref(props.modelValue)

const instanceId = Math.random().toString(36).slice(2)
const modelUriStr = `json-editor://${instanceId}/editor.json`
const schemaBaseUri = `schema://json-editor-modal/${instanceId}/`
const rootSchemaUri = `${schemaBaseUri}pipeline.schema.json`
const workerReadyTimeoutMs = 500

let activeSessionToken = 0
let contentListener: IDisposable | null = null

function normalizeRefPath(path: string): string {
    return path.replace(/^\.\//, '')
}

function tryFormatJson(value: string): string {
    if (!value.trim()) return value
    try {
        return JSON.stringify(JSON.parse(value), null, 2)
    } catch {
        return value
    }
}

function buildOwnedSchemaUris() {
    const ownedUris = new Set<string>([rootSchemaUri])
    if (props.externalSchemas) {
        for (const refPath of Object.keys(props.externalSchemas)) {
            ownedUris.add(`${schemaBaseUri}${normalizeRefPath(refPath)}`)
        }
    }
    return ownedUris
}

function buildSessionSchemas(): MonacoJsonSchemaEntry[] {
    if (!props.schema) {
        return []
    }

    const nextSchemas: MonacoJsonSchemaEntry[] = [
        {
            uri: rootSchemaUri,
            fileMatch: [modelUriStr],
            schema: props.schema,
        },
    ]

    if (props.externalSchemas) {
        for (const [refPath, schemaObj] of Object.entries(props.externalSchemas)) {
            nextSchemas.push({
                uri: `${schemaBaseUri}${normalizeRefPath(refPath)}`,
                schema: schemaObj,
            })
        }
    }

    return nextSchemas
}

function unregisterContentListener() {
    contentListener?.dispose()
    contentListener = null
}

function waitForJsonWorkerReady(model: MonacoEditorNamespace.ITextModel, sessionToken: number): Promise<void> {
    const resource = model.uri

    return new Promise((resolve) => {
        let settled = false

        const finish = () => {
            if (settled) {
                return
            }
            settled = true
            markerListener.dispose()
            clearTimeout(timeoutId)
            resolve()
        }

        const markerListener = monaco.editor.onDidChangeMarkers((changedResources) => {
            if (!open.value || sessionToken !== activeSessionToken) {
                finish()
                return
            }

            if (changedResources.some(uri => uri.toString() === resource.toString())) {
                finish()
            }
        })

        const timeoutId = window.setTimeout(() => {
            finish()
        }, workerReadyTimeoutMs)

        queueMicrotask(() => {
            if (!open.value || sessionToken !== activeSessionToken) {
                finish()
                return
            }

            const markers = monaco.editor.getModelMarkers({ resource })
            if (markers.length > 0) {
                finish()
            }
        })
    })
}

async function ensureSessionInitialized() {
    if (!open.value) return

    const sessionToken = ++activeSessionToken
    const formattedDraft = tryFormatJson(props.modelValue)
    draft.value = formattedDraft

    try {
        await runMonacoJsonSession(async () => {
            await ensureMonacoReady()
            const editor = await editorRef.value?.whenReady()
            if (!editor || !open.value || sessionToken !== activeSessionToken) {
                return
            }

            const ownedUris = buildOwnedSchemaUris()
            const currentOptions = getJsonDiagnosticsOptions()
            const filteredSchemas = (currentOptions.schemas ?? []).filter(schema => !ownedUris.has(schema.uri))
            const sessionSchemas = buildSessionSchemas()

            setJsonDiagnosticsOptions({
                ...currentOptions,
                validate: true,
                schemas: [...filteredSchemas, ...sessionSchemas],
            })

            const uri = monaco.Uri.parse(modelUriStr)
            let model = monaco.editor.getModel(uri)
            if (!model) {
                model = monaco.editor.createModel(formattedDraft, props.language, uri)
            } else {
                if (model.getValue() !== formattedDraft) {
                    model.setValue(formattedDraft)
                }
                if (model.getLanguageId() !== props.language) {
                    monaco.editor.setModelLanguage(model, props.language)
                }
            }

            const oldModel = editor.getModel()
            if (oldModel !== model) {
                editor.setModel(model)
            }

            unregisterContentListener()
            contentListener = model.onDidChangeContent(() => {
                draft.value = model!.getValue()
            })

            await waitForJsonWorkerReady(model, sessionToken)
        })
    } finally {
        if (sessionToken !== activeSessionToken || !open.value) {
            return
        }
    }
}

async function cleanupSession() {
    ++activeSessionToken
    unregisterContentListener()

    await runMonacoJsonSession(async () => {
        await ensureMonacoReady()

        const currentOptions = getJsonDiagnosticsOptions()
        const filteredSchemas = (currentOptions.schemas ?? []).filter(schema => !schema.uri.startsWith(schemaBaseUri))
        if (filteredSchemas.length !== (currentOptions.schemas ?? []).length) {
            setJsonDiagnosticsOptions({
                ...currentOptions,
                schemas: filteredSchemas,
            })
        }

        const model = monaco.editor.getModel(monaco.Uri.parse(modelUriStr)) as MonacoEditorNamespace.ITextModel | null
        model?.dispose()
    })
}

watch(open, async (isOpen) => {
    if (isOpen) {
        await ensureSessionInitialized()
        return
    }

    await cleanupSession()
})

watch(
    () => props.modelValue,
    (newVal) => {
        if (!open.value) draft.value = newVal
    },
)

function onSave() {
    const editor = editorRef.value?.getEditor()
    if (editor) {
        const model = editor.getModel()
        if (model) {
            const markers = monaco.editor.getModelMarkers({ resource: model.uri })
            const errors = markers.filter(m => m.severity === monaco.MarkerSeverity.Error)
            if (errors.length > 0) {
                const firstError = errors[0]!
                toast.add({
                    id: 'json-editor-error',
                    title: 'JSON / Schema Error',
                    description: `Line ${firstError.startLineNumber}: ${firstError.message}`,
                    icon: 'i-lucide-circle-x',
                    color: 'error',
                })
                return
            }
        }
    }

    emit('update:modelValue', draft.value)
    open.value = false
}

function onCancel() {
    open.value = false
}

onBeforeUnmount(() => {
    void cleanupSession()
})
</script>
