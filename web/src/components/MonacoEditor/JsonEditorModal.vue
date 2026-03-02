<template>
    <UModal v-model:open="open" :title="title" :description="description">
        <template #body>
            <MonacoEditor ref="editorRef" v-model="draft" :language="language" :read-only="false" :min-height="300"
                          :max-height="500" />
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
import { ref, watch, onBeforeUnmount, nextTick } from 'vue'
import { MonacoEditor, monaco } from '@/components/MonacoEditor'
import { MarkerSeverity } from 'monaco-editor'

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

const instanceId = Math.random().toString(36).slice(2)
const modelUriStr = `json-editor://${instanceId}/editor.json`
const schemaBaseUri = `schema://json-editor-modal/${instanceId}/`
const rootSchemaUri = `${schemaBaseUri}pipeline.schema.json`

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

function registerSchemas() {
    if (!props.schema) return

    const currentOptions = monaco.json.jsonDefaults.diagnosticsOptions
    const existingSchemas = (currentOptions.schemas ?? []) as Array<{
        uri: string
        fileMatch?: string[]
        schema: Record<string, unknown>
    }>

    const ownedUris = new Set<string>([rootSchemaUri])
    if (props.externalSchemas) {
        for (const refPath of Object.keys(props.externalSchemas)) {
            ownedUris.add(`${schemaBaseUri}${normalizeRefPath(refPath)}`)
        }
    }

    const filtered = existingSchemas.filter(s => !ownedUris.has(s.uri))

    const nextSchemas: Array<{ uri: string; fileMatch?: string[]; schema: Record<string, unknown> }> = [
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

    monaco.json.jsonDefaults.setDiagnosticsOptions({
        ...currentOptions,
        validate: true,
        schemas: [...filtered, ...nextSchemas],
    })
}

function removeSchemas() {
    const currentOptions = monaco.json.jsonDefaults.diagnosticsOptions
    const existingSchemas = (currentOptions.schemas ?? []) as Array<{
        uri: string
        fileMatch?: string[]
        schema: Record<string, unknown>
    }>

    const filtered = existingSchemas.filter(s => !s.uri.startsWith(schemaBaseUri))
    if (filtered.length !== existingSchemas.length) {
        monaco.json.jsonDefaults.setDiagnosticsOptions({
            ...currentOptions,
            schemas: filtered,
        })
    }
}

const draft = ref(props.modelValue)

if (props.schema) registerSchemas()

function applySchemaModel() {
    const editor = editorRef.value?.getEditor()
    if (!editor || !props.schema) return

    const uri = monaco.Uri.parse(modelUriStr)
    let model = monaco.editor.getModel(uri)
    if (!model) {
        model = monaco.editor.createModel(draft.value, props.language, uri)
    } else if (model.getValue() !== draft.value) {
        model.setValue(draft.value)
    }

    const oldModel = editor.getModel()
    if (oldModel !== model) {
        editor.setModel(model)
        if (oldModel && oldModel.uri.toString() !== modelUriStr) {
            oldModel.dispose()
        }
    }

    model.onDidChangeContent(() => {
        draft.value = model!.getValue()
    })
}

watch(open, async (isOpen) => {
    if (!isOpen) return

    draft.value = tryFormatJson(props.modelValue)
    registerSchemas()

    await nextTick()
    setTimeout(() => {
        applySchemaModel()
    }, 50)
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
            const errors = markers.filter(m => m.severity === MarkerSeverity.Error)
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
    removeSchemas()
    const model = monaco.editor.getModel(monaco.Uri.parse(modelUriStr))
    model?.dispose()
})
</script>
