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
  label?: string;
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

export interface TaskImageRef {
  id: string;
  url: string;
  mime: string;
  width?: number;
  height?: number;
  content_size?: number;
}

export interface RecoDetailResponse {
  name: string;
  algorithm: string;
  hit: boolean;
  box?: RectResponse;
  detail_json?: unknown;
  combined_result?: RecoDetailResponse[];
  draw_images?: TaskImageRef[];
  raw_image?: TaskImageRef;
  results?: RecoResultsResponse;
}

export interface PointResponse {
  x: number;
  y: number;
}

export interface ClickActionResult {
  type: "Click";
  point: PointResponse;
  contact: number;
  pressure: number;
}

export interface LongPressActionResult {
  type: "LongPress";
  point: PointResponse;
  duration: number;
  contact: number;
  pressure: number;
}

export interface SwipeActionResult {
  type: "Swipe";
  begin: PointResponse;
  end: PointResponse[];
  end_hold: number[];
  duration: number[];
  only_hover: boolean;
  starting: number;
  contact: number;
  pressure: number;
}

export interface MultiSwipeActionResult {
  type: "MultiSwipe";
  swipes: Omit<SwipeActionResult, "type">[];
}

export interface TouchActionResult {
  type: "TouchDown" | "TouchMove" | "TouchUp";
  point: PointResponse;
  contact: number;
  pressure: number;
}

export interface ScrollActionResult {
  type: "Scroll";
  point: PointResponse;
  dx: number;
  dy: number;
}

export interface ClickKeyActionResult {
  type: "ClickKey" | "KeyDown" | "KeyUp";
  keycode: number[];
}

export interface LongPressKeyActionResult {
  type: "LongPressKey";
  keycode: number[];
  duration: number;
}

export interface InputTextActionResult {
  type: "InputText";
  text: string;
}

export interface AppActionResult {
  type: "StartApp" | "StopApp";
  package: string;
}

export interface ShellActionResult {
  type: "Shell" | "Command";
  cmd: string;
  timeout: number;
  success: boolean;
  output: string;
}

export interface GenericActionResult {
  type: "DoNothing" | "StopTask" | "Custom" | string;
}

export type ActionResult =
  | ClickActionResult
  | LongPressActionResult
  | SwipeActionResult
  | MultiSwipeActionResult
  | TouchActionResult
  | ScrollActionResult
  | ClickKeyActionResult
  | LongPressKeyActionResult
  | InputTextActionResult
  | AppActionResult
  | ShellActionResult
  | GenericActionResult;

export type ActionTypeWithCoords =
  | "Click"
  | "LongPress"
  | "Swipe"
  | "MultiSwipe"
  | "TouchDown"
  | "TouchMove"
  | "TouchUp"
  | "Scroll";

export function actionHasCoords(
  result: ActionResult | undefined,
): result is
  | ClickActionResult
  | LongPressActionResult
  | SwipeActionResult
  | MultiSwipeActionResult
  | TouchActionResult
  | ScrollActionResult {
  if (!result) return false;
  const types: string[] = [
    "Click",
    "LongPress",
    "Swipe",
    "MultiSwipe",
    "TouchDown",
    "TouchMove",
    "TouchUp",
    "Scroll",
  ];
  return types.includes(result.type);
}

export interface ActionDetailResponse {
  name: string;
  action: string;
  box?: RectResponse;
  success: boolean;
  detail_json?: unknown;
  result?: ActionResult;
  raw_image?: TaskImageRef;
  controller_type?: string;
}

export interface NodeDataResponse {
  name: string;
  node_json: string;
}

export interface NodeDetailResponse {
  name: string;
  recognition?: RecoDetailResponse;
  action?: ActionDetailResponse;
  run_completed: boolean;
}
