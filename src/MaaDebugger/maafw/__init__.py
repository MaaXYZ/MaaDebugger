import re
import io
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Tuple, Union

from asyncify import asyncify
from PIL import Image
from maa.controller import (
    AdbController,
    Win32Controller,
    GamepadController,
    CustomController,
    PlayCoverController,
)
from maa.define import MaaGamepadTypeEnum
from maa.context import Context, ContextEventSink
from maa.tasker import Tasker, TaskerEventSink, RecognitionDetail
from maa.resource import Resource, ResourceEventSink
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from maa.agent_client import AgentClient
from maa.library import Library
from maa.event_sink import NotificationType
import numpy as np

from ..utils.img_tools import cvmat_to_image
from .launch_graph import LaunchGraph, reduce_launch_graph, Scope, ScopeType
from ..utils.img_tools import cvmat_to_image
from ..utils.arg_parser import ArgParser
from .launch_graph import LaunchGraph, reduce_launch_graph, Scope, ScopeType

debug_mode = ArgParser.get_debug()


class MyCustomController(CustomController):
    def __init__(self, img_path: Path):
        super().__init__()

        img = Image.open(img_path).convert("RGB")
        # 将 RGB 转换为 BGR 供 OpenCV 使用
        self.img = np.array(img)[:, :, ::-1]

    def connect(self) -> bool:
        return True

    def request_uuid(self) -> str:
        return "0"

    def screencap(self) -> np.ndarray:
        return self.img


class MaaFW:

    resource: Optional[Resource]
    controller: Union[
        AdbController,
        Win32Controller,
        GamepadController,
        CustomController,
        PlayCoverController,
        None,
    ]
    tasker: Optional[Tasker]
    agent: Optional[AgentClient]
    agent_identifier: Optional[str]
    context_event_sink: Optional[ContextEventSink]
    resource_event_sink: Optional[ResourceEventSink]
    tasker_event_sink: Optional[TaskerEventSink]

    def __init__(self):
        Toolkit.init_option("./")
        Tasker.set_debug_mode(True)

        self.resource = None
        self.controller = None
        self.tasker = None
        self.agent = None
        self.agent_identifier = None
        self.context_event_sink = None
        self.resource_event_sink = None
        self.tasker_event_sink = None

        self.screenshotter = Screenshotter(self.screencap)

    @property
    def version(self) -> str:
        return Library.version()

    @staticmethod
    @asyncify
    def detect_adb() -> List[AdbDevice]:
        return Toolkit.find_adb_devices()

    @staticmethod
    @asyncify
    def detect_win32hwnd(window_regex: str) -> List[DesktopWindow]:
        windows = Toolkit.find_desktop_windows()
        result = []
        for win in windows:
            if re.search(window_regex, win.window_name):
                result.append(win)

        return result

    @asyncify
    def connect_adb(
        self, path: Path, address: str, config: dict
    ) -> Tuple[bool, Optional[str]]:
        self.controller = AdbController(path, address, config=config)

        if self.controller is None:
            return False, "Controller is None!"
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {path} {address}"

        return True, None

    @asyncify
    def connect_win32(
        self,
        hwnd: str,
        screencap_method: int,
        mouse_method: int,
        keyboard_method: int,
    ) -> Tuple[bool, Optional[str]]:
        _hwnd = int(hwnd, 16)

        self.controller = Win32Controller(
            _hwnd,
            screencap_method=screencap_method,
            mouse_method=mouse_method,
            keyboard_method=keyboard_method,
        )

        if self.controller is None:
            return False, "Controller is None!"
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {hwnd}"

        return True, None

    @asyncify
    def connect_playcover(self, address: str, uuid: str) -> Tuple[bool, Optional[str]]:
        self.controller = PlayCoverController(address, uuid)

        if self.controller is None:
            return False, "Controller is None!"
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {address}"

        return True, None

    @asyncify
    def connect_gamepad(
        self, hwnd: str, gamepad_type: MaaGamepadTypeEnum, screencap_method: int
    ):
        _hwnd = int(hwnd, 16)

        self.controller = GamepadController(_hwnd, gamepad_type, screencap_method)

        if self.controller is None:
            return False, "Controller is None!"
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {hwnd}"

        return True, None

    def connect_custom_controller(self, img_path: Path) -> Tuple[bool, Optional[str]]:
        self.controller = MyCustomController(img_path)

        if self.controller is None:
            return False, "Controller is None!"
        self.controller.set_screenshot_use_raw_size(True)
        self.controller.post_connection().wait()

        return True, None

    @asyncify
    def load_resource(self, dir: List[Path]) -> Tuple[bool, Optional[str]]:
        if not self.resource:
            self.resource = Resource()
            if self.resource_event_sink:
                self.resource.add_sink(self.resource_event_sink)

        if not self.resource:
            return False, "Resource is None!"

        if not self.resource.clear():
            return False, "Fail to clear Resource!"

        for d in dir:
            if not d.exists():
                return False, f"{d} does not exist."

            status = self.resource.post_bundle(d).wait().succeeded
            if not status:
                return (
                    False,
                    "Fail to load resource, please check the outputs of CLI.",
                )
        return True, None

    @asyncify
    def create_agent(self, identifier: Optional[str]) -> Tuple[bool, Optional[str]]:
        self.agent = None  # del existing agent

        if not self.resource:
            self.resource = Resource()
        agent = AgentClient(identifier or None)
        agent.set_timeout(1000 * 5)
        if not agent.bind(self.resource):
            return False, "Failed to bind Resource to AgentClient."

        self.agent_identifier = agent.identifier
        self.agent = agent
        return True, None

    @asyncify
    def connect_agent(self) -> Tuple[bool, Optional[str]]:
        if not self.agent:
            return False, "AgentClient is not created."

        if self.agent.connect():
            return True, None
        else:
            return False, "Failed to connect AgentClient."

    @asyncify
    def run_task(
        self, entry: str, pipeline_override: dict = {}
    ) -> Tuple[bool, Optional[str]]:
        if not self.tasker:
            self.tasker = Tasker()
            # 添加 ContextEventSink
            if self.context_event_sink:
                self.tasker.add_context_sink(self.context_event_sink)
            # 添加 TaskerEventSink 以支持 Task 级别事件
            if self.tasker_event_sink:
                self.tasker.add_sink(self.tasker_event_sink)

        if not self.resource:
            return False, "Resource is not initialized."

        if not self.controller:
            return False, "Controller is not initialized."

        if not self.tasker:
            return False, "Tasker is None!"

        self.tasker.bind(self.resource, self.controller)
        if not self.tasker.inited:
            return False, "Failed to initialize Tasker."

        if self.agent:
            if not AgentClient(self.agent.identifier).register_sink(
                self.resource, self.controller, self.tasker
            ):
                return False, "Failed to register AgentClientSink."

        if isinstance(self.controller, CustomController):
            # disable action
            pipeline_override.update(
                {entry: {"action": {"type": "DoNothing"}, "next": []}}
            )
            return (
                self.tasker.post_task(entry, pipeline_override).wait().succeeded,
                None,
            )
        else:
            return (
                self.tasker.post_task(entry, pipeline_override).wait().succeeded,
                None,
            )

    @asyncify
    def stop_task(self) -> None:
        if not self.tasker:
            return

        self.tasker.post_stop().wait()

    @asyncify
    def screencap(self, capture: bool = True) -> Optional[Image.Image]:
        if not self.controller:
            return None

        if capture:
            self.controller.post_screencap().wait()
        im = self.controller.cached_image
        if im is None:
            return None

        return cvmat_to_image(im)

    @asyncify
    def click(self, x, y) -> bool:
        if not self.controller:
            return False

        if isinstance(self.controller, CustomController):
            return False

        return self.controller.post_click(x, y).wait().succeeded

    @asyncify
    def get_reco_detail(self, reco_id: int) -> Optional[RecognitionDetail]:
        if not self.tasker:
            return None

        return self.tasker.get_recognition_detail(reco_id)

    @asyncify
    def clear_cache(self) -> bool:
        if not self.tasker:
            return False

        return self.tasker.clear_cache()

    @asyncify
    def get_node_list(self) -> list[str]:
        if self.resource:
            return self.resource.node_list
        else:
            return []

    # @asyncify
    def get_node_data(self, name: str) -> dict:
        if self.resource:
            return self.resource.get_node_data(name) or {}
        else:
            return {}


class LaunchGraphTaskerEventSink(TaskerEventSink):
    """Tasker 级别事件处理，用于 Task.Starting/Succeeded/Failed"""

    def __init__(self, graph_manager: "LaunchGraphManager") -> None:
        self.graph_manager = graph_manager

    def on_tasker_task(
        self,
        tasker: Tasker,
        noti_type: NotificationType,
        detail: Any,  # TaskerEventSink.TaskerTaskDetail
    ):
        """处理 Task 级别事件"""
        msg_suffix = {
            NotificationType.Starting: "Starting",
            NotificationType.Succeeded: "Succeeded",
            NotificationType.Failed: "Failed",
        }.get(noti_type)

        if msg_suffix:
            self.graph_manager.dispatch(
                {
                    "msg": f"Task.{msg_suffix}",
                    "entry": detail.entry,
                    "task_id": detail.task_id,
                    "uuid": detail.uuid,
                }
            )


class LaunchGraphContextEventSink(ContextEventSink):
    """状态机驱动的 EventSink，将所有事件转换为消息发送到状态机"""

    def __init__(self, graph_manager: "LaunchGraphManager") -> None:
        self.graph_manager = graph_manager

    def _get_msg_suffix(self, noti_type: NotificationType) -> Optional[str]:
        return {
            NotificationType.Starting: "Starting",
            NotificationType.Succeeded: "Succeeded",
            NotificationType.Failed: "Failed",
        }.get(noti_type)

    def on_node_pipeline_node(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodePipelineNodeDetail
    ):
        """处理 PipelineNode 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"PipelineNode.{msg_suffix}",
                "name": detail.name,
                "node_id": detail.node_id,
            }
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)

    def on_node_recognition_node(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodeRecognitionNodeDetail
    ):
        """处理 RecognitionNode 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"RecognitionNode.{msg_suffix}",
                "name": detail.name,
                "node_id": detail.node_id,
            }
            print(f"[DEBUG EventSink] {msg}")
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)

    def on_node_action_node(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodeActionNodeDetail
    ):
        """处理 ActionNode 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"ActionNode.{msg_suffix}",
                "name": detail.name,
                "node_id": detail.node_id,
            }
            print(f"[DEBUG EventSink] {msg}")
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)

    def on_node_next_list(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodeNextListDetail
    ):
        """处理 NextList 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"NextList.{msg_suffix}",
                "name": detail.name,
                "next_list": [attr.name for attr in detail.next_list],
            }
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)

    def on_node_recognition(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodeRecognitionDetail
    ):
        """处理 Recognition 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"Recognition.{msg_suffix}",
                "name": detail.name,
                "reco_id": detail.reco_id,
            }
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)

    def on_node_action(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: Any,  # ContextEventSink.NodeActionDetail
    ):
        """处理 Action 事件"""
        msg_suffix = self._get_msg_suffix(noti_type)
        if msg_suffix:
            msg = {
                "msg": f"Action.{msg_suffix}",
                "name": detail.name,
            }
            if debug_mode:
                print(f"[DEBUG EventSink] {msg}")
            self.graph_manager.dispatch(msg)


class LaunchGraphManager:
    """
    状态机管理器
    负责管理 LaunchGraph 的状态更新和订阅
    """

    def __init__(self) -> None:
        self._graph = LaunchGraph()
        self._subscribers: List[Callable[[LaunchGraph, Dict[str, Any]], None]] = []

    @property
    def graph(self) -> LaunchGraph:
        """获取当前状态图（只读）"""
        return self._graph

    def reset(self) -> None:
        """重置状态机"""
        self._graph = LaunchGraph()
        self._notify_subscribers({"msg": "Reset"})

    def dispatch(self, msg: Dict[str, Any]) -> None:
        """
        分发消息到状态机

        Args:
            msg: 消息字典，必须包含 "msg" 字段
        """
        self._graph = reduce_launch_graph(self._graph, msg)
        self._notify_subscribers(msg)

    def subscribe(
        self, callback: Callable[[LaunchGraph, Dict[str, Any]], None]
    ) -> Callable[[], None]:
        """
        订阅状态变化

        Args:
            callback: 状态变化时调用的回调函数，接收 (graph, msg) 两个参数

        Returns:
            取消订阅的函数
        """
        self._subscribers.append(callback)

        def unsubscribe():
            if callback in self._subscribers:
                self._subscribers.remove(callback)

        return unsubscribe

    def _notify_subscribers(self, msg: Dict[str, Any]) -> None:
        """通知所有订阅者"""
        for callback in self._subscribers:
            try:
                callback(self._graph, msg)
            except Exception as e:
                print(f"[LaunchGraphManager] Subscriber error: {e}")

    def get_current_task(self) -> Optional[Scope]:
        """获取当前正在执行的任务"""
        if self._graph.childs:
            return self._graph.childs[-1]
        return None

    def get_all_recognitions(self) -> List[Dict[str, Any]]:
        """
        获取所有识别记录
        遍历整个执行图，收集所有 Recognition 节点
        """
        results: List[Dict[str, Any]] = []

        def traverse(scope: Scope, depth: int = 0):
            if scope.type == ScopeType.RECO:
                results.append(
                    {
                        "msg": scope.msg,
                        "status": scope.status,
                        "depth": depth,
                    }
                )

            # 递归遍历所有子节点
            for child in scope.childs:
                traverse(child, depth + 1)
            if scope.reco:
                for reco in scope.reco:
                    traverse(reco, depth + 1)
            if scope.action:
                traverse(scope.action, depth + 1)
            if scope.reco_detail:
                traverse(scope.reco_detail, depth + 1)

        for task in self._graph.childs:
            traverse(task)

        return results


class MyResourceEventSink(ResourceEventSink):
    def __init__(self, on_resource_loading: Callable) -> None:
        self.on_resource_loading = on_resource_loading


class Screenshotter:
    source: Optional[Image.Image] = None
    screencap_func: Callable

    def __init__(self, screencap_func: Callable):
        self.screencap_func = screencap_func

    def __del__(self):
        self.source = None

    async def refresh(self, capture: bool = True):
        im: Image.Image = await self.screencap_func(capture)
        if im is not None:
            self.source = im


maafw = MaaFW()
