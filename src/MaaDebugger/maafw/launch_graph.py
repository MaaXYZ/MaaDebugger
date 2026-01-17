"""
MaaFramework 任务执行状态机
基于 TypeScript 版本的状态机实现，用于追踪和管理任务执行的完整生命周期
"""

from dataclasses import dataclass, field
from typing import Optional, List, Dict, TypeVar, Union
from strenum import StrEnum

from ..utils.arg_parser import ArgParser

debug_mode: bool = ArgParser.get_debug()

# 消息中可能包含的值类型
MsgValue = Union[str, int, List[str], None]

# 消息字典类型：键为字符串，值可以是字符串、整数、字符串列表或 None
MsgDict = Dict[str, MsgValue]

# Scope 字典类型（用于序列化）
ScopeDictValue = Union[str, int, List[str], "ScopeDict", List["ScopeDict"], None]
ScopeDict = Dict[str, ScopeDictValue]

# 泛型类型变量，用于 _last_of 函数
T = TypeVar("T")


class GeneralStatus(StrEnum):
    """通用状态枚举"""

    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"


class ScopeType(StrEnum):
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
    msg: MsgDict = field(default_factory=dict)
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

    def to_dict(self) -> Dict[str, Union[int, List[ScopeDict]]]:
        """转换为字典，便于序列化和调试"""
        return {
            "depth": self.depth,
            "childs": [self._scope_to_dict(child) for child in self.childs],
        }

    @staticmethod
    def _scope_to_dict(scope: Scope) -> ScopeDict:
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


def _last_of(arr: List[Scope]) -> Optional[Scope]:
    """获取列表的最后一个元素"""
    return arr[-1] if arr else None


# ============ 辅助函数 ============


def is_nested_recognition(scope: Scope) -> bool:
    """
    判断 Recognition 是否来自嵌套调用（RecognitionNode）

    Args:
        scope: 要检查的作用域

    Returns:
        True 如果该 Recognition 是在 RecognitionNode 内部触发的（嵌套调用）
        False 如果是来自 NextList 的正常识别流程
    """
    if scope.type != ScopeType.RECO:
        return False
    # 如果父节点是 RECO_NODE，则是嵌套调用
    return scope.parent is not None and scope.parent.type == ScopeType.RECO_NODE


def get_parent_chain(scope: Scope) -> List[Scope]:
    """
    获取从当前节点到根节点的完整路径

    Args:
        scope: 起始作用域

    Returns:
        从当前节点到根节点的路径列表（包含当前节点）
    """
    chain: List[Scope] = []
    current: Optional[Scope] = scope
    while current is not None:
        chain.append(current)
        current = current.parent
    return chain


def find_root_pipeline_node(scope: Scope) -> Optional[Scope]:
    """
    找到最顶层的 PipelineNode

    从当前节点向上遍历，找到最接近根的 PipelineNode。
    这在追踪嵌套调用时很有用，可以确定某个识别属于哪个顶层任务。

    Args:
        scope: 起始作用域

    Returns:
        最顶层的 PipelineNode，如果没有找到则返回 None
    """
    root_pipeline: Optional[Scope] = None
    current: Optional[Scope] = scope
    while current is not None:
        if current.type == ScopeType.PIPELINE_NODE:
            root_pipeline = current
        current = current.parent
    return root_pipeline


def find_immediate_pipeline_node(scope: Scope) -> Optional[Scope]:
    """
    找到最近的 PipelineNode 父节点

    Args:
        scope: 起始作用域

    Returns:
        最近的 PipelineNode 父节点，如果没有找到则返回 None
    """
    current: Optional[Scope] = scope.parent
    while current is not None:
        if current.type == ScopeType.PIPELINE_NODE:
            return current
        current = current.parent
    return None


def get_nesting_depth(scope: Scope) -> int:
    """
    获取当前作用域的嵌套深度

    计算从当前节点到根的 PipelineNode 数量。
    深度为 1 表示顶层执行，大于 1 表示嵌套调用。

    Args:
        scope: 要检查的作用域

    Returns:
        嵌套深度（PipelineNode 的数量）
    """
    depth = 0
    current: Optional[Scope] = scope
    while current is not None:
        if current.type == ScopeType.PIPELINE_NODE:
            depth += 1
        current = current.parent
    return depth


def _iterate_tracker(tracker: Optional[Scope]) -> Optional[Scope]:
    """
    迭代追踪器，找到当前深度的下一个节点

    Args:
        tracker: 当前作用域

    Returns:
        下一个要追踪的作用域，如果没有则返回 None
    """
    if tracker is None:
        return None

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


def reduce_launch_graph(current: LaunchGraph, msg: Dict[str, MsgValue]) -> LaunchGraph:
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
        if debug_mode:
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
        elif debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, reason: no root")
        return current

    # 深度 > 0，需要追踪到当前节点
    top_scope = _last_of(task.childs)
    if not top_scope and debug_mode:
        print(f"[LaunchGraph] Drop msg: {msg_type}, reason: no root")
        return current

    tracker = top_scope
    for i in range(1, current.depth):
        new_tracker = _iterate_tracker(tracker)
        if not new_tracker:
            if tracker and debug_mode:
                print(
                    f"[LaunchGraph] Drop msg: {msg_type}, reason: trace failed at depth {i}"
                )
            return current
        tracker = new_tracker

    # 根据消息类型更新状态机
    if msg_type == "PipelineNode.Starting":
        if tracker and tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.PIPELINE_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                reco=[],
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("PipelineNode.Succeeded", "PipelineNode.Failed"):
        if tracker and tracker.type == ScopeType.PIPELINE_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "PipelineNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "RecognitionNode.Starting":
        if tracker and tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.RECO_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("RecognitionNode.Succeeded", "RecognitionNode.Failed"):
        if tracker and tracker.type == ScopeType.RECO_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "RecognitionNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif (
            tracker
            and tracker.type == ScopeType.RECO
            and hasattr(tracker, "reco_detail") is False
        ):
            # 如果在 RECO 中但不是标准结构，检查父节点
            # 这种情况可能是从 reco_node.reco_detail 追踪过来的
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "ActionNode.Starting":
        if tracker and tracker.type in (ScopeType.RECO, ScopeType.ACTION):
            new_scope = Scope(
                type=ScopeType.ACTION_NODE,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("ActionNode.Succeeded", "ActionNode.Failed"):
        if tracker and tracker.type == ScopeType.ACTION_NODE:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "ActionNode.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "NextList.Starting":
        if tracker and tracker.type == ScopeType.PIPELINE_NODE:
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
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("NextList.Succeeded", "NextList.Failed"):
        if tracker and tracker.type == ScopeType.NEXT_LIST:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "NextList.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "Recognition.Starting":
        if tracker and tracker.type == ScopeType.RECO_NODE:
            new_scope = Scope(
                type=ScopeType.RECO,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.reco_detail = new_scope
            current.depth += 1
        elif tracker and tracker.type == ScopeType.NEXT_LIST:
            new_scope = Scope(
                type=ScopeType.RECO,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.childs.append(new_scope)
            current.depth += 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("Recognition.Succeeded", "Recognition.Failed"):
        if tracker and tracker.type == ScopeType.RECO:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "Recognition.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type == "Action.Starting":
        if tracker and tracker.type in (ScopeType.PIPELINE_NODE, ScopeType.ACTION_NODE):
            new_scope = Scope(
                type=ScopeType.ACTION,
                msg=msg,
                status=GeneralStatus.RUNNING,
                parent=tracker,
            )
            tracker.action = new_scope
            current.depth += 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif msg_type in ("Action.Succeeded", "Action.Failed"):
        if tracker and tracker.type == ScopeType.ACTION:
            tracker.msg = msg
            tracker.status = (
                GeneralStatus.SUCCESS
                if msg_type == "Action.Succeeded"
                else GeneralStatus.FAILED
            )
            current.depth -= 1
        elif tracker and debug_mode:
            print(f"[LaunchGraph] Drop msg: {msg_type}, tracker type: {tracker.type}")

    elif debug_mode:
        print(f"[LaunchGraph] Drop msg: unknown type {msg_type}")

    return current


# 全局状态机实例
launch_graph = LaunchGraph()
