import { ref } from "vue";
import type {
  LaunchGraph,
  TaskScope,
  PipelineNodeScope,
  NextListScope,
  RecoScope,
  ActionScope,
  AllScope,
  TaskEvent,
  RecoNodeScope,
  ActionNodeScope,
} from "@/components/Index/taskDetail/types";

function lastOf<T>(arr: T[]): T | undefined {
  return arr.length > 0 ? arr[arr.length - 1] : undefined;
}

function iterateTracker(tracker: AllScope): AllScope | undefined {
  switch (tracker.type) {
    case "pipeline_node":
      if (tracker.action) return tracker.action;
      return lastOf(tracker.reco);
    case "reco_node":
      return tracker.reco ?? undefined;
    case "act_node":
      return tracker.action ?? undefined;
    case "next":
      return lastOf(tracker.childs);
    case "reco":
    case "act":
      return lastOf(tracker.childs);
  }
}

/**
 * 全局 LaunchGraph 响应式状态 (使用深层 ref 保证所有属性修改都能触发更新)
 */
export const launchGraph = ref<LaunchGraph>({
  depth: 0,
  childs: [],
});

/**
 * 参考 maa-js reduceLaunchGraph 状态机实现
 * 直接修改 ref 的内部状态，由 Vue 进行深度响应
 */
export function reduceLaunchGraph(graph: LaunchGraph, msg: TaskEvent) {
  switch (msg.msg) {
    case "Task.Starting":
      graph.childs.push({
        type: "task",
        msg,
        status: "running",
        childs: [],
      } as TaskScope);
      graph.depth = 0;
      return;
    case "Task.Succeeded": {
      const task = lastOf(graph.childs);
      if (task) {
        task.msg = msg;
        task.status = "success";
      }
      return;
    }
    case "Task.Failed": {
      const task = lastOf(graph.childs);
      if (task) {
        task.msg = msg;
        task.status = "failed";
      }
      return;
    }
  }

  const task = lastOf(graph.childs);
  if (!task) {
    console.log("[LaunchGraph] drop msg, no task:", msg);
    return;
  }

  if (graph.depth === 0) {
    switch (msg.msg) {
      case "PipelineNode.Starting":
        task.childs.push({
          type: "pipeline_node",
          msg,
          status: "running",
          reco: [],
          action: null,
        } as PipelineNodeScope);
        graph.depth++;
        return;
      default:
        console.log("[LaunchGraph] drop msg, no root:", msg);
        return;
    }
  }

  const topScope = lastOf(task.childs);
  if (!topScope) {
    console.log("[LaunchGraph] drop msg, no root:", msg);
    return;
  }

  let tracker: AllScope = topScope;
  for (let i = 1; i < graph.depth; i++) {
    const next = iterateTracker(tracker);
    if (!next) {
      console.log("[LaunchGraph] drop msg, trace failed:", msg);
      return;
    }
    tracker = next;
  }

  switch (msg.msg) {
    case "PipelineNode.Starting":
      if (tracker.type === "reco" || tracker.type === "act") {
        tracker.childs.push({
          type: "pipeline_node",
          msg,
          status: "running",
          reco: [],
          action: null,
        } as PipelineNodeScope);
        graph.depth++;
        return;
      }
      break;
    case "PipelineNode.Succeeded":
    case "PipelineNode.Failed":
      if (tracker.type === "pipeline_node") {
        tracker.msg = msg;
        tracker.status =
          msg.msg === "PipelineNode.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;

    case "RecognitionNode.Starting":
      if (tracker.type === "reco" || tracker.type === "act") {
        tracker.childs.push({
          type: "reco_node",
          msg,
          status: "running",
          reco: null,
        } as RecoNodeScope);
        graph.depth++;
        return;
      }
      break;
    case "RecognitionNode.Succeeded":
    case "RecognitionNode.Failed":
      if (tracker.type === "reco_node") {
        tracker.msg = msg;
        tracker.status =
          msg.msg === "RecognitionNode.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;

    case "ActionNode.Starting":
      if (tracker.type === "reco" || tracker.type === "act") {
        tracker.childs.push({
          type: "act_node",
          msg,
          status: "running",
          action: null,
        } as ActionNodeScope);
        graph.depth++;
        return;
      }
      break;
    case "ActionNode.Succeeded":
    case "ActionNode.Failed":
      if (tracker.type === "act_node") {
        tracker.msg = msg;
        tracker.status =
          msg.msg === "ActionNode.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;

    case "NextList.Starting":
      if (tracker.type === "pipeline_node") {
        tracker.reco.push({
          type: "next",
          msg,
          status: "running",
          childs: [],
        } as NextListScope);
        graph.depth++;
        return;
      }
      break;
    case "NextList.Succeeded":
    case "NextList.Failed":
      if (tracker.type === "next") {
        tracker.msg = msg;
        tracker.status =
          msg.msg === "NextList.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;

    case "Recognition.Starting":
      if (tracker.type === "reco_node") {
        tracker.reco = {
          type: "reco",
          msg,
          status: "running",
          childs: [],
        } as RecoScope;
        graph.depth++;
        return;
      } else if (tracker.type === "next") {
        tracker.childs.push({
          type: "reco",
          msg,
          status: "running",
          childs: [],
        } as RecoScope);
        graph.depth++;
        return;
      }
      break;
    case "Recognition.Succeeded":
    case "Recognition.Failed":
      if (tracker.type === "reco") {
        tracker.msg = msg;
        tracker.status =
          msg.msg === "Recognition.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;

    case "Action.Starting":
      if (tracker.type === "pipeline_node" || tracker.type === "act_node") {
        tracker.action = {
          type: "act",
          msg,
          status: "running",
          childs: [],
        } as ActionScope;
        graph.depth++;
        return;
      }
      break;
    case "Action.Succeeded":
    case "Action.Failed":
      if (tracker.type === "act") {
        tracker.msg = msg;
        tracker.status = msg.msg === "Action.Succeeded" ? "success" : "failed";
        graph.depth--;
        return;
      }
      break;
  }

  console.log("[LaunchGraph] drop msg:", msg);
}

/**
 * 处理从 WS 收到的 task.event 消息
 */
export function handleTaskEvent(msg: TaskEvent) {
  reduceLaunchGraph(launchGraph.value, msg);
}

/**
 * 重置 LaunchGraph
 */
export function resetLaunchGraph() {
  launchGraph.value = { depth: 0, childs: [] };
}
