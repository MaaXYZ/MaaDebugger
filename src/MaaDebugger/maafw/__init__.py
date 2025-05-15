import re
from pathlib import Path
from typing import Callable, List, Optional, Tuple, Union

from asyncify import asyncify
from PIL import Image
from maa.controller import AdbController, Win32Controller
from maa.tasker import Tasker, RecognitionDetail, NotificationHandler
from maa.resource import Resource
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from maa.agent_client import AgentClient

from ..utils import cvmat_to_image


class MaaFW:

    resource: Optional[Resource]
    controller: Union[AdbController, Win32Controller, None]
    tasker: Optional[Tasker]
    agent: Optional[AgentClient]
    notification_handler: Optional[NotificationHandler]

    def __init__(self):
        Toolkit.init_option("./")
        Tasker.set_debug_mode(True)

        self.resource = None
        self.controller = None
        self.tasker = None
        self.agent = None

        self.screenshotter = Screenshotter(self.screencap)
        self.notification_handler = None

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
        self, hwnd: Union[int, str], screencap_method: int, input_method: int
    ) -> Tuple[bool, Optional[str]]:
        if isinstance(hwnd, str):
            hwnd = int(hwnd, 16)

        self.controller = Win32Controller(
            hwnd, screencap_method=screencap_method, input_method=input_method
        )
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            return (False, f"Failed to connect {hex(hwnd)}")

        return (True, None)

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
                    "Fail to load resource,please check the outputs of CLI.",
                )
        return (True, None)

    @asyncify
    def create_agent(self, identifier: str) -> str:
        if not self.resource:
            self.resource = Resource()

        if not self.agent or self.agent.identifier != identifier:
            self.agent = AgentClient(identifier)
            self.agent.bind(self.resource)

        return self.agent.identifier

    @asyncify
    def connect_agent(self) -> Tuple[bool, Optional[str]]:
        if self.agent and self.agent.connect():
            return True, None
        else:
            return False, "Failed to connect agent"

    @asyncify
    def run_task(
        self, entry: str, pipeline_override: dict = {}
    ) -> Tuple[bool, Optional[str]]:
        if not self.tasker:
            self.tasker = Tasker(notification_handler=self.notification_handler)

        if not self.resource or not self.controller:
            return False, "Resource or Controller not initialized"

        self.tasker.bind(self.resource, self.controller)
        if not self.tasker.inited:
            return False, "Failed to init MaaFramework tasker"

        return self.tasker.post_task(entry, pipeline_override).wait().succeeded, None

    @asyncify
    def stop_task(self):
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


# class Screenshotter(threading.Thread):
class Screenshotter:
    def __init__(self, screencap_func: Callable):
        super().__init__()
        self.source = None
        self.screencap_func = screencap_func
        # self.active = False

    def __del__(self):
        self.source = None
        # self.active = False

    async def refresh(self, capture: bool = True):
        im = await self.screencap_func(capture)
        if not im:
            return

        self.source = im

    # def run(self):
    #     while self.active:
    #         self.refresh()
    #         time.sleep(0)

    # def start(self):
    #     self.active = True
    #     super().start()

    # def stop(self):
    #     self.active = False


maafw = MaaFW()
