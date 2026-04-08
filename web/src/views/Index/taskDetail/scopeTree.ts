import type {
  ActionScope,
  AnyNodeScope,
  GeneralStatus,
  NextListScope,
  RecoScope,
  TaskScope,
} from "@/types/taskDetail";

function hasRunningInRecoScope(scope: RecoScope): boolean {
  if (scope.status === "running") return true;
  return hasRunningInAnyNodes(scope.childs);
}

function hasRunningInActionScope(scope: ActionScope): boolean {
  if (scope.status === "running") return true;
  return hasRunningInAnyNodes(scope.childs);
}

export function hasRunningInAnyNodes(nodes: AnyNodeScope[]): boolean {
  return nodes.some((node) => {
    if (node.status === "running") return true;

    switch (node.type) {
      case "pipeline_node":
        return (
          node.reco.some((nextList) =>
            nextList.childs.some((reco) => hasRunningInRecoScope(reco)),
          ) ||
          (node.action != null && hasRunningInActionScope(node.action))
        );
      case "reco_node":
        return node.reco != null && hasRunningInRecoScope(node.reco);
      case "act_node":
        return node.action != null && hasRunningInActionScope(node.action);
    }
  });
}

function hasFailedInRecoScope(scope: RecoScope): boolean {
  if (scope.status === "failed") return true;
  return hasFailedInAnyNodes(scope.childs);
}

function hasFailedInActionScope(scope: ActionScope): boolean {
  if (scope.status === "failed") return true;
  return hasFailedInAnyNodes(scope.childs);
}

export function hasFailedInAnyNodes(nodes: AnyNodeScope[]): boolean {
  return nodes.some((node) => {
    if (node.status === "failed") return true;

    switch (node.type) {
      case "pipeline_node":
        return (
          node.reco.some((nextList) =>
            nextList.childs.some((reco) => hasFailedInRecoScope(reco)),
          ) ||
          (node.action != null && hasFailedInActionScope(node.action))
        );
      case "reco_node":
        return node.reco != null && hasFailedInRecoScope(node.reco);
      case "act_node":
        return node.action != null && hasFailedInActionScope(node.action);
    }
  });
}

export function summarizeAnyNodesStatus(nodes: AnyNodeScope[]): GeneralStatus {
  if (nodes.length === 0) return "success";
  if (hasRunningInAnyNodes(nodes)) return "running";
  if (hasFailedInAnyNodes(nodes)) return "failed";
  return "success";
}

function findRecoNameInRecoScope(
  scope: RecoScope,
  recoId: number,
): string | null {
  if (scope.msg.reco_id === recoId) return scope.msg.name;
  return findRecoNameInAnyNodes(scope.childs, recoId);
}

function findRecoNameInActionScope(
  scope: ActionScope,
  recoId: number,
): string | null {
  return findRecoNameInAnyNodes(scope.childs, recoId);
}

export function findRecoNameInAnyNodes(
  nodes: AnyNodeScope[],
  recoId: number,
): string | null {
  for (const node of nodes) {
    switch (node.type) {
      case "pipeline_node": {
        for (const nextList of node.reco) {
          for (const reco of nextList.childs) {
            const found = findRecoNameInRecoScope(reco, recoId);
            if (found) return found;
          }
        }
        if (node.action) {
          const found = findRecoNameInActionScope(node.action, recoId);
          if (found) return found;
        }
        break;
      }
      case "reco_node": {
        if (!node.reco) break;
        const found = findRecoNameInRecoScope(node.reco, recoId);
        if (found) return found;
        break;
      }
      case "act_node": {
        if (!node.action) break;
        const found = findRecoNameInActionScope(node.action, recoId);
        if (found) return found;
        break;
      }
    }
  }
  return null;
}

export function findRecoNameInTasks(
  tasks: TaskScope[],
  recoId: number,
): string | null {
  for (const task of tasks) {
    const found = findRecoNameInAnyNodes(task.childs, recoId);
    if (found) return found;
  }
  return null;
}

export function countRecoSubflowNodes(scope: RecoScope): number {
  return scope.childs.length;
}

export function countActionSubflowNodes(scope: ActionScope): number {
  return scope.childs.length;
}

export function countNextListReco(nextList: NextListScope): number {
  return nextList.childs.length;
}

export function nodeStableKey(node: AnyNodeScope, index: number): string {
  switch (node.type) {
    case "pipeline_node":
      return `pipeline-${node.msg.node_id}-${index}`;
    case "reco_node":
      return `reco-node-${node.msg.node_id}-${index}`;
    case "act_node":
      return `act-node-${node.msg.node_id}-${index}`;
  }
}

export function nextListStableKey(
  nextList: NextListScope,
  index: number,
): string {
  return `next-${nextList.msg.name}-${index}`;
}

export function recoStableKey(scope: RecoScope, index: number): string {
  return `reco-${scope.msg.reco_id}-${index}`;
}
