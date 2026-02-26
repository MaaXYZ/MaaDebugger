export type NodeStatus =
  | "success"
  | "failed"
  | "running"
  | "pending"
  | "skipped";

export type GeneralStatus = "running" | "success" | "failed";

// ============================================================
// LaunchGraph Scope Types (参考 maa-js 状态机)
// ============================================================

export interface TaskScope {
  type: "task";
  msg: TaskEventMsg;
  status: GeneralStatus;
  childs: PipelineNodeScope[];
}

export interface PipelineNodeScope {
  type: "pipeline_node";
  msg: PipelineNodeEventMsg;
  status: GeneralStatus;
  reco: NextListScope[];
  action: ActionScope | null;
}

export interface RecoNodeScope {
  type: "reco_node";
  msg: RecognitionNodeEventMsg;
  status: GeneralStatus;
  reco: RecoScope | null;
}

export interface ActionNodeScope {
  type: "act_node";
  msg: ActionNodeEventMsg;
  status: GeneralStatus;
  action: ActionScope | null;
}

export interface NextListScope {
  type: "next";
  msg: NextListEventMsg;
  status: GeneralStatus;
  childs: RecoScope[];
}

export interface RecoScope {
  type: "reco";
  msg: RecognitionEventMsg;
  status: GeneralStatus;
  childs: AnyNodeScope[];
}

export interface ActionScope {
  type: "act";
  msg: ActionEventMsg;
  status: GeneralStatus;
  childs: AnyNodeScope[];
}

export type AnyNodeScope = PipelineNodeScope | RecoNodeScope | ActionNodeScope;
export type AllScope = AnyNodeScope | NextListScope | RecoScope | ActionScope;

export interface LaunchGraph {
  depth: number;
  childs: TaskScope[];
}

// ============================================================
// Event Message Types (from Go backend WS)
// ============================================================

export interface TaskEventMsg {
  msg: `Task.${"Starting" | "Succeeded" | "Failed"}`;
  task_id: number;
  entry: string;
  uuid: string;
}

export interface PipelineNodeEventMsg {
  msg: `PipelineNode.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  node_id: number;
}

export interface RecognitionNodeEventMsg {
  msg: `RecognitionNode.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  node_id: number;
}

export interface ActionNodeEventMsg {
  msg: `ActionNode.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  node_id: number;
}

export interface NextListItem {
  name: string;
  jump_back: boolean;
  anchor: boolean;
}

export interface NextListEventMsg {
  msg: `NextList.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  list: NextListItem[];
}

export interface RecognitionEventMsg {
  msg: `Recognition.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  reco_id: number;
}

export interface ActionEventMsg {
  msg: `Action.${"Starting" | "Succeeded" | "Failed"}`;
  name: string;
  action_id: number;
}

export type TaskEvent =
  | TaskEventMsg
  | PipelineNodeEventMsg
  | RecognitionNodeEventMsg
  | ActionNodeEventMsg
  | NextListEventMsg
  | RecognitionEventMsg
  | ActionEventMsg;

// ============================================================
// Reco Detail Response (from GET /api/task/node/{name})
// ============================================================

export interface RectResponse {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface RecoResultItem {
  box?: RectResponse;
  extra?: Record<string, unknown>;
}

export interface RecoResultsResponse {
  all: RecoResultItem[];
  best: RecoResultItem[];
  filtered: RecoResultItem[];
}

export interface RecoDetailResponse {
  name: string;
  algorithm: string;
  hit: boolean;
  box?: RectResponse;
  detail_json?: unknown;
  combined_result?: RecoDetailResponse[];
  draw_images?: string[];
  raw_image?: string;
  results?: RecoResultsResponse;
}

export interface ActionDetailResponse {
  name: string;
  action: string;
  box?: RectResponse;
  success: boolean;
  detail_json?: unknown;
}

export interface NodeDetailResponse {
  name: string;
  recognition?: RecoDetailResponse;
  action?: ActionDetailResponse;
  run_completed: boolean;
}
