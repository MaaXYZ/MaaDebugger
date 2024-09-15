import re
from asyncify import asyncify
from pathlib import Path
from typing import Callable, List, Optional

from maa.controller import AdbController, Win32Controller
from maa.tasker import Tasker, RecognitionDetail
from maa.resource import Resource
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from PIL import Image

from ..utils import cvmat_to_image


class MaaFW:

    tasker: Tasker

    def __init__(
        self,
        on_list_to_recognize: Callable = None,
        on_miss_all: Callable = None,
        on_recognition_result: Callable = None,
    ):
        Toolkit.init_option("./")
        Tasker.set_debug_message(True)

        self.resource = None
        self.controller = None
        self.tasker = None

        self.screenshotter = Screenshotter(self.screencap)

        self.on_list_to_recognize = on_list_to_recognize
        self.on_miss_all = on_miss_all
        self.on_recognition_result = on_recognition_result

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
        connected = self.controller.post_connection().wait().success()
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
        connected = self.controller.post_connection().wait().success()
        if not connected:
            print(f"Failed to connect {hwnd}")
            return False

        return True

    @asyncify
    def load_resource(self, dir: Path) -> bool:
        if not self.resource:
            self.resource = Resource()

        return self.resource.clear() and self.resource.post_path(dir).wait().success()

    @asyncify
    def run_task(self, entry: str, pipeline_override: dict = {}) -> bool:
        if not self.tasker:
            self.tasker = Tasker(callback=self._tasker_callback)

        self.tasker.bind(self.resource, self.controller)
        if not self.tasker.inited:
            print("Failed to init MaaFramework instance")
            return False

        return self.tasker.post_pipeline(entry, pipeline_override).wait()

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
    def click(self, x, y) -> None:
        if not self.controller:
            return None

        self.controller.post_click(x, y).wait()

    @asyncify
    def get_reco_detail(self, reco_id: int) -> Optional[RecognitionDetail]:
        if not self.tasker:
            return None

        return self.tasker._get_recognition_detail(reco_id)

    def _tasker_callback(self, msg: str, detail: dict, arg):
        if msg == "Task.Debug.ListToRecognize":
            self.screenshotter.refresh(False)
            if self.on_list_to_recognize:
                self.on_list_to_recognize(detail["current"], detail["list"])

        elif msg == "Task.Debug.MissAll":
            if self.on_miss_all:
                self.on_miss_all(detail["current"], detail["list"])

        elif msg == "Task.Debug.RecognitionResult":
            reco = detail["recognition"]
            reco_id = reco["reco_id"]
            name = reco["name"]
            hit = reco["box"] is not None

            if self.on_recognition_result:
                self.on_recognition_result(reco_id, name, hit)


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
