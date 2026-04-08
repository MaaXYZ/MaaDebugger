export interface CheckResponse {
  level: "error" | "warning";
  msg: string;
  path: string;
  line: string;
}
