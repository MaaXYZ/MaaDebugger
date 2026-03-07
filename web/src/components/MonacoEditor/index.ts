export { default as MonacoEditor } from "./MonacoEditor.vue";
export type { MonacoEditorProps } from "./MonacoEditor.vue";
export { default as JsonEditorModal } from "./JsonEditorModal.vue";
export {
  monaco,
  ensureMonacoReady,
  getJsonDiagnosticsOptions,
  setJsonDiagnosticsOptions,
  runMonacoJsonSession,
  warmupMonacoJsonWorker,
} from "./setup";
export type { MonacoJsonSchemaEntry } from "./setup";
