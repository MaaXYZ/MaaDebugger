import re
from pathlib import Path
from typing import Callable, List, Optional, Tuple, Union

from asyncify import asyncify
from PIL import Image
from maa.controller import AdbController, Win32Controller
from maa.context import Context, ContextEventSink
from maa.tasker import Tasker, RecognitionDetail
from maa.resource import Resource
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from maa.agent_client import AgentClient
from maa.library import Library
from maa.event_sink import NotificationType

from ..utils import cvmat_to_image


class MaaFW:

    resource: Optional[Resource]
    controller: Union[AdbController, Win32Controller, None]
    tasker: Optional[Tasker]
    agent: Optional[AgentClient]
    event_sink: Optional[ContextEventSink]

    def __init__(self):
        Toolkit.init_option("./")
        Tasker.set_debug_mode(True)

        self.resource = None
        self.controller = None
        self.tasker = None
        self.agent = None

        self.screenshotter = Screenshotter(self.screencap)
        self.event_sink = None

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
            if not re.search(window_regex, win.window_name):
                continue

            result.append(win)

        return result

    @asyncify
    def connect_adb(
        self, path: Path, address: str, config: dict
    ) -> Tuple[bool, Optional[str]]:
        self.controller = AdbController(path, address, config=config)
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {path} {address}"

        return True, None

    @asyncify
    def connect_win32hwnd(
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
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return False, f"Failed to connect {hex(_hwnd)}"

        return True, None

    @asyncify
    def load_resource(self, dir: List[Path]) -> Tuple[bool, Optional[str]]:
        if not self.resource:
            self.resource = Resource()

        self.resource.clear()
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
    def create_agent(self, identifier: str) -> Optional[str]:
        if not self.resource:
            self.resource = Resource()

        self.agent = AgentClient(identifier)
        self.agent.bind(self.resource)

        return self.agent.identifier

    @asyncify
    def connect_agent(self) -> Tuple[bool, Optional[str]]:
        if self.agent and self.agent.connect():
            return True, None
        else:
            return False, "Failed to connect agent."

    @asyncify
    def run_task(
        self, entry: str, pipeline_override: dict = {}
    ) -> Tuple[bool, Optional[str]]:
        if not self.event_sink:
            assert ValueError("EventSink is None.")

        if not self.tasker:
            self.tasker = Tasker()
            self.tasker.add_context_sink(self.event_sink)

        if not self.resource:
            return False, "Resource is not initialized."

        if not self.controller:
            return False, "Controller is not initialized."

        self.tasker.bind(self.resource, self.controller)
        if not self.tasker.inited:
            return False, "Failed to initialize Tasker."

        return self.tasker.post_task(entry, pipeline_override).wait().succeeded, None

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


class MyContextEventSink(ContextEventSink):
    def __init__(
        self,
        on_next_list_starting: Optional[Callable],
        on_recognized: Optional[Callable],
    ) -> None:
        self.on_next_list_starting = on_next_list_starting
        self.on_recognized = on_recognized

    def on_node_next_list(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: ContextEventSink.NodeNextListDetail,
    ):
        if noti_type != NotificationType.Starting:
            return

        if self.on_next_list_starting is not None:
            self.on_next_list_starting(detail.name, detail.next_list)

    def on_node_recognition(
        self,
        _: Context,
        noti_type: NotificationType,
        detail: ContextEventSink.NodeRecognitionDetail,
    ):
        if (
            noti_type != NotificationType.Succeeded
            and noti_type != NotificationType.Failed
        ):
            return

        if self.on_recognized is not None:
            self.on_recognized(
                detail.reco_id, detail.name, noti_type == NotificationType.Succeeded
            )


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
