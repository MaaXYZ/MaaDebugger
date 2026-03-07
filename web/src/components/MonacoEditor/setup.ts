import * as monaco from "monaco-editor/esm/vs/editor/editor.api.js";
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import * as jsonContribution from "monaco-editor/esm/vs/language/json/monaco.contribution.js";

export interface MonacoJsonSchemaEntry {
  uri: string;
  fileMatch?: string[];
  schema: Record<string, unknown>;
}

interface MonacoJsonDiagnosticsOptions {
  validate?: boolean;
  allowComments?: boolean;
  trailingCommas?: string;
  schemas?: MonacoJsonSchemaEntry[];
}

const jsonDefaults = (
  jsonContribution as typeof jsonContribution & {
    jsonDefaults: {
      diagnosticsOptions: MonacoJsonDiagnosticsOptions;
      setDiagnosticsOptions(options: MonacoJsonDiagnosticsOptions): void;
    };
  }
).jsonDefaults;

const defaultJsonDiagnosticsOptions: MonacoJsonDiagnosticsOptions = {
  validate: true,
  allowComments: true,
  trailingCommas: "ignore",
};

let monacoInitialized = false;
let monacoReadyPromise: Promise<typeof monaco> | null = null;
let monacoWarmupPromise: Promise<void> | null = null;
let schemaSessionQueue = Promise.resolve();

function configureMonacoEnvironment() {
  if (monacoInitialized) {
    return;
  }

  self.MonacoEnvironment = {
    getWorker(_: unknown, label: string) {
      if (label === "json") {
        return new jsonWorker();
      }
      return new editorWorker();
    },
  };

  jsonDefaults.setDiagnosticsOptions({ ...defaultJsonDiagnosticsOptions });
  monacoInitialized = true;
}

export function ensureMonacoReady(): Promise<typeof monaco> {
  monacoReadyPromise ??= Promise.resolve().then(() => {
    configureMonacoEnvironment();
    return monaco;
  });

  return monacoReadyPromise;
}

export function getJsonDiagnosticsOptions(): MonacoJsonDiagnosticsOptions {
  return {
    ...defaultJsonDiagnosticsOptions,
    ...jsonDefaults.diagnosticsOptions,
    schemas: [...(jsonDefaults.diagnosticsOptions.schemas ?? [])],
  };
}

export function setJsonDiagnosticsOptions(
  options: MonacoJsonDiagnosticsOptions,
) {
  jsonDefaults.setDiagnosticsOptions({
    ...defaultJsonDiagnosticsOptions,
    ...options,
    schemas: options.schemas ? [...options.schemas] : undefined,
  });
}

export async function runMonacoJsonSession<T>(
  task: () => Promise<T> | T,
): Promise<T> {
  const previousSession = schemaSessionQueue;
  let releaseSession!: () => void;

  schemaSessionQueue = new Promise<void>((resolve) => {
    releaseSession = resolve;
  });

  await previousSession;
  await ensureMonacoReady();

  try {
    return await task();
  } finally {
    releaseSession();
  }
}

export function warmupMonacoJsonWorker(): Promise<void> {
  monacoWarmupPromise ??= runMonacoJsonSession(async () => {
    const uri = monaco.Uri.parse("json-warmup://singleton/warmup.json");
    let model = monaco.editor.getModel(uri);
    const created = !model;

    if (!model) {
      model = monaco.editor.createModel("{}", "json", uri);
    }

    await new Promise<void>((resolve) => {
      let settled = false;

      const finish = () => {
        if (settled) {
          return;
        }
        settled = true;
        markerListener.dispose();
        clearTimeout(timeoutId);
        resolve();
      };

      const markerListener = monaco.editor.onDidChangeMarkers((changed) => {
        if (
          changed.some((resource) => resource.toString() === uri.toString())
        ) {
          finish();
        }
      });

      const timeoutId = window.setTimeout(() => {
        finish();
      }, 500);

      queueMicrotask(() => {
        const markers = monaco.editor.getModelMarkers({ resource: uri });
        if (markers.length > 0) {
          finish();
        }
      });
    });

    if (created) {
      model.dispose();
    }
  });

  return monacoWarmupPromise;
}

void ensureMonacoReady();
void warmupMonacoJsonWorker();

export { monaco };
