import re
from asyncify import asyncify
from pathlib import Path
from typing import Callable, List, Optional

from maa.controller import AdbController, Win32Controller
from maa.tasker import Tasker, RecognitionDetail, NotificationHandler
from maa.resource import Resource
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from PIL import Image

from ..utils import cvmat_to_image


class MaaFW:

    resource: Resource | None
    controller: AdbController | Win32Controller | None
    tasker: Tasker | None
    notification_handler: NotificationHandler | None

    def __init__(self):
        Toolkit.init_option("./")
        Tasker.set_debug_mode(True)

        self.resource = None
        self.controller = None
        self.tasker = None

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
    def connect_adb(self, path: Path, address: str, config: dict) -> bool:
        self.controller = AdbController(path, address, config=config)
        connected = self.controller.post_connection().wait().succeeded()
        if not connected:
            print(f"Failed to connect {path} {address}")
            return False

        return True

    @asyncify
    def connect_win32hwnd(
        self, hwnd: int | str, screencap_method: int, input_method: int
    ) -> bool:
        if isinstance(hwnd, str):
            hwnd = int(hwnd, 16)

        self.controller = Win32Controller(
            hwnd, screencap_method=screencap_method, input_method=input_method
        )
        connected = self.controller.post_connection().wait().succeeded()
        if not connected:
            print(f"Failed to connect {hwnd}")
            return False

        return True

    @asyncify
    def load_resource(self, dir: Path) -> bool:
        if not self.resource:
            self.resource = Resource()

        return self.resource.clear() and self.resource.post_path(dir).wait().succeeded()

    @asyncify
    def run_task(self, entry: str, pipeline_override: dict = {}) -> bool:
        if not self.tasker:
            self.tasker = Tasker(notification_handler=self.notification_handler)

        if not self.resource or not self.controller:
            print("Resource or Controller not initialized")
            return False

        self.tasker.bind(self.resource, self.controller)
        if not self.tasker.inited:
            print("Failed to init MaaFramework instance")
            return False

        return self.tasker.post_pipeline(entry, pipeline_override).wait().succeeded()

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

        return self.controller.post_click(x, y).wait().succeeded()

    @asyncify
    def get_reco_detail(self, reco_id: int) -> Optional[RecognitionDetail]:
        if not self.tasker:
            return None

        return self.tasker.get_recognition_detail(reco_id)

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
