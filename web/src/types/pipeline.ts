export interface CheckResponse {
  level: "error" | "warning";
  code: string;
  msg: string;
  task?: string;
  path: string;
  line: string;
}
