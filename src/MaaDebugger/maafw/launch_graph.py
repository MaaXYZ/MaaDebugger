"""
MaaFramework 任务执行状态机
基于 TypeScript 版本的状态机实现，用于追踪和管理任务执行的完整生命周期
"""

from dataclasses import dataclass, field
from typing import Any, Optional, List, Dict
from enum import Enum


class GeneralStatus(str, Enum):
    """通用状态枚举"""

    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"


class ScopeType(str, Enum):
    """作用域类型枚举"""

    TASK = "task"
    PIPELINE_NODE = "pipeline_node"
    RECO_NODE = "reco_node"
    ACTION_NODE = "act_node"
    NEXT_LIST = "next"
    RECO = "reco"
    ACTION = "act"


@dataclass
class Scope:
    """通用作用域基类"""

    type: str
    msg: Dict[str, Any] = field(default_factory=dict)
    status: GeneralStatus = GeneralStatus.RUNNING
    childs: List["Scope"] = field(default_factory=list)
    reco: Optional[List["Scope"]] = None  # for pipeline_node
    action: Optional["Scope"] = None  # for pipeline_node, act_node
    reco_detail: Optional["Scope"] = None  # for reco_node
    parent: Optional["Scope"] = field(default=None, repr=False)  # 父节点引用


@dataclass
class LaunchGraph:
    """执行图的根结构"""

    depth: int = 0
    childs: List[Scope] = field(default_factory=list)

    def to_dict(self) -> Dict[str, Any]:
        """转换为字典，便于序列化和调试"""
        return {
            "depth": self.depth,
            "childs": [self._scope_to_dict(child) for child in self.childs],
        }

    @staticmethod
    def _scope_to_dict(scope: Scope) -> Dict[str, Any]:
        """递归转换 Scope 为字典"""
        result = {
            "type": scope.type,
            "status": scope.status.value,
            "msg": scope.msg,
        }

        if scope.childs:
            result["childs"] = [LaunchGraph._scope_to_dict(c) for c in scope.childs]
        if scope.reco:
            result["reco"] = [LaunchGraph._scope_to_dict(r) for r in scope.reco]
        if scope.action:
            result["action"] = LaunchGraph._scope_to_dict(scope.action)
        if scope.reco_detail:
            result["reco_detail"] = LaunchGraph._scope_to_dict(scope.reco_detail)

        return result


def _last_of(arr: List[Any]) -> Optional[Any]:
    """获取列表的最后一个元素"""
    return arr[-1] if arr else None


def _iterate_tracker(tracker: Scope) -> Optional[Scope]:
    """
    迭代追踪器，找到当前深度的下一个节点

    Args:
        tracker: 当前作用域

    Returns:
        下一个要追踪的作用域，如果没有则返回 None
    """
    if tracker.type == ScopeType.PIPELINE_NODE:
        if tracker.action:
            return tracker.action
        elif tracker.reco:
            return _last_of(tracker.reco)
    elif tracker.type == ScopeType.RECO_NODE:
        return tracker.reco_detail
    elif tracker.type == ScopeType.ACTION_NODE:
        return tracker.action
    elif tracker.type == ScopeType.NEXT_LIST:
        return _last_of(tracker.childs)
    elif tracker.type in (ScopeType.RECO, ScopeType.ACTION):
        return _last_of(tracker.childs)
    return None


def reduce_launch_graph(current: LaunchGraph, msg: Dict[str, Any]) -> LaunchGraph:
    """
    状态机的 reducer 函数，根据消息更新执行图（原地修改）

    Args:
        current: 当前的执行图状态
        msg: 从 MaaFramework 接收到的消息

    Returns:
        更新后的执行图（同一个实例，已被原地修改）

    Note:
        与 TypeScript 版本使用 immer 实现不可变更新不同，
        Python 版本采用原地修改以避免 deepcopy 带来的内存开销。
    """
    msg_type = msg.get("msg", "")

    # 处理 Task 级别的消息
    if msg_type == "Task.Starting":
        new_scope = Scope(
            type=ScopeType.TASK,
            msg=msg,
            status=GeneralStatus.RUNNING,
        )
        current.childs.append(new_scope)
        current.depth = 0
        return current

    elif msg_type == "Task.Succeeded":
        task = _last_of(current.childs)
        if task:
            task.msg = msg
            task.status = GeneralStatus.SUCCESS
        return current

    elif msg_type == "Task.Failed":
        task = _last_of(current.childs)
        if task:
            task.msg = msg
            task.status = GeneralStatus.FAILED
        return current

    # 获取当前任务
    task = _last_of(current.childs)
    if not task:
        print(f"[LaunchGraph] Drop msg: {msg_type}, reason: no task")
        return current

    # 深度为 0 时，只能处理 PipelineNode.Starting
    if current.depth == 0:
        if msg_type == "PipelineNode.Starting":
            new_scope = Scope(
                type=ScopeType.PIPELINE_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                reco=[],
                parent=task,
            )
            task.childs.append(new_scope)
            current.depth += 1
            return current
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, reason: no root")
            return current

    # 深度 > 0，需要追踪到当前节点
    top_scope = _last_of(task.childs)
    if not top_scope:
        print(f"[LaunchGraph] Drop msg: {msg_type}, reason: no root")
        return current

    tracker = top_scope
    for i in range(1, current.depth):
        new_tracker = _iterate_tracker(tracker)
        if not new_tracker:
            print(
                f"[LaunchGraph] Drop msg: {msg_type}, reason: trace failed at depth {i}"
            )
            return current
        tracker = new_tracker

    # 根据消息类型更新状态机
    if msg_type == "PipelineNode.Starting":
        if tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.PIPELINE_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                reco=[],
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("PipelineNode.Succeeded", "PipelineNode.Failed"):
        if tracker.type == ScopeType.PIPELINE_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "PipelineNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "RecognitionNode.Starting":
        if tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.RECO_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("RecognitionNode.Succeeded", "RecognitionNode.Failed"):
        if tracker.type == ScopeType.RECO_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "RecognitionNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif (
            tracker.type == ScopeType.RECO and hasattr(tracker, "reco_detail") is False
        ):
            # 如果在 RECO 中但不是标准结构，检查父节点
            # 这种情况可能是从 reco_node.reco_detail 追踪过来的
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "ActionNode.Starting":
        if tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.ACTION_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("ActionNode.Succeeded", "ActionNode.Failed"):
        if tracker.type == ScopeType.ACTION_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "ActionNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "NextList.Starting":
        if tracker.type == ScopeType.PIPELINE_NODE:
            if tracker.reco is None:
                tracker.reco = []
            new_scope = Scope(
                type=ScopeType.NEXT_LIST,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.reco.append(new_scope)
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("NextList.Succeeded", "NextList.Failed"):
        if tracker.type == ScopeType.NEXT_LIST:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "NextList.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "Recognition.Starting":
        if tracker.type == ScopeType.RECO_NODE:
            new_scope = Scope(
                type=ScopeType.RECO,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.reco_detail = new_scope
            current.depth += 1
        elif tracker.type == ScopeType.NEXT_LIST:
            new_scope = Scope(
                type=ScopeType.RECO,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("Recognition.Succeeded", "Recognition.Failed"):
        if tracker.type == ScopeType.RECO:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "Recognition.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "Action.Starting":
        if tracker.type in (ScopeType.PIPELINE_NODE, ScopeType.ACTION_NODE):
            new_scope = Scope(
                type=ScopeType.ACTION,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.action = new_scope
            current.depth += 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("Action.Succeeded", "Action.Failed"):
        if tracker.type == ScopeType.ACTION:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "Action.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        else:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    else:
        print(f"[LaunchGraph] Drop msg: unknown type {msg_type}")

    return current


# 全局状态机实例
launch_graph = LaunchGraph()
