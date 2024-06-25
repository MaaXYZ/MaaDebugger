import asyncio
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, List, Optional

from maa.controller import AdbController, Win32Controller
from maa.instance import Instance
from maa.resource import Resource
from maa.toolkit import Toolkit
from PIL import Image

from ..utils import cvmat_to_image


class MaaFW:
    def __init__(
        self,
        on_list_to_recognize: Callable = None,
        on_miss_all: Callable = None,
        on_recognition_result: Callable = None,
    ):
        Toolkit.init_option("./")
        Instance.set_debug_message(True)

        self.resource = None
        self.controller = None
        self.instance = None

        self.screenshotter = Screenshotter(self.screencap)

        self.on_list_to_recognize = on_list_to_recognize
        self.on_miss_all = on_miss_all
        self.on_recognition_result = on_recognition_result

    @staticmethod
    async def detect_adb() -> List["AdbDevice"]:
        return await Toolkit.adb_devices()

    @dataclass
    class Window:
        hwnd: int
        class_name: str
        window_name: str

    @staticmethod
    async def detect_win32hwnd(class_regex: str, window_regex: str) -> List[Window]:
        hwnds = Toolkit.search_window(class_regex, window_regex)
        windows = []
        for hwnd in hwnds:
            class_name = Toolkit.get_class_name(hwnd)
            window_name = Toolkit.get_window_name(hwnd)
            windows.append(MaaFW.Window(hwnd, class_name, window_name))

        return windows

    async def connect_adb(self, path: Path, address: str, config: dict) -> bool:
        self.controller = AdbController(path, address, config=config)
        connected = await self.controller.connect()
        if not connected:
            print(f"Failed to connect {path} {address}")
            return False

        return True

    async def connect_win32hwnd(
        self, hwnd: int | str, screencap_type: int, input_type: int
    ) -> bool:
        if isinstance(hwnd, str):
            hwnd = int(hwnd, 16)

        self.controller = Win32Controller(
            hwnd, screencap_type=screencap_type, touch_type=input_type, key_type=0
        )
        connected = await self.controller.connect()
        if not connected:
            print(f"Failed to connect {hwnd}")
            return False

        return True

    async def load_resource(self, dir: Path) -> bool:
        if not self.resource:
            self.resource = Resource()

        return self.resource.clear() and await self.resource.load(dir)

    async def run_task(self, entry: str, param: dict = {}) -> bool:
        if not self.instance:
            self.instance = Instance(callback=self._inst_callback)

        self.instance.bind(self.resource, self.controller)
        if not self.instance.inited:
            print("Failed to init MaaFramework instance")
            return False

        return await self.instance.run_task(entry, param)

    async def stop_task(self):
        if not self.instance:
            return

        await self.instance.stop()

    async def screencap(self, capture: bool = True) -> Optional[Image.Image]:
        if not self.controller:
            return None

        im = await self.controller.screencap(capture)
        if im is None:
            return None

        return cvmat_to_image(im)

    async def click(self, x, y) -> None:
        if not self.controller:
            return None

        await self.controller.click(x, y)

    def _inst_callback(self, msg: str, detail: dict, arg):
        match msg:
            case "Task.Debug.ListToRecognize":
                asyncio.run(self.screenshotter.refresh(False))
                if self.on_list_to_recognize:
                    self.on_list_to_recognize(detail["pre_hit_task"], detail["list"])

            case "Task.Debug.MissAll":
                if self.on_miss_all:
                    self.on_miss_all(detail["pre_hit_task"], detail["list"])

            case "Task.Debug.RecognitionResult":
                reco_id = detail["recognition"]["id"]
                name = detail["name"]
                hit = detail["recognition"]["hit"]

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
