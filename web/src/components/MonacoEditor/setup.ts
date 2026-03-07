import * as monaco from "monaco-editor/esm/vs/editor/editor.api.js";
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import * as jsonContribution from "monaco-editor/esm/vs/language/json/monaco.contribution.js";

const jsonDefaults = (
  jsonContribution as typeof jsonContribution & {
    jsonDefaults: {
      diagnosticsOptions: {
        schemas?: Array<{
          uri: string;
          fileMatch?: string[];
          schema: Record<string, unknown>;
        }>;
      };
      setDiagnosticsOptions(options: unknown): void;
    };
  }
).jsonDefaults;

jsonDefaults.setDiagnosticsOptions({
  validate: true,
  allowComments: true,
  trailingCommas: "ignore",
});

self.MonacoEnvironment = {
  getWorker(_: unknown, label: string) {
    if (label === "json") {
      return new jsonWorker();
    }
    return new editorWorker();
  },
};

export { monaco };
