export type NodeStatus =
  | "success"
  | "failed"
  | "running"
  | "pending"
  | "skipped";

export interface RecoDetail {
  recoId: number;
  name: string;
  status: NodeStatus;
}

export interface NodeDetail {
  nodeId: number;
  name: string;
  recoList: RecoDetail[];
  actionStatus: NodeStatus;
}
