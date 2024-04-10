from pathlib import Path
from typing import List, Optional, Callable
from asyncer import asyncify
from PIL import Image
import sys
import asyncio
import threading
import time


class MaaFW:

    def __init__(
        self,
        on_list_to_recognize: Callable = None,
        on_recognition_result: Callable = None,
    ):
        self.resource = None
        self.controller = None
        self.instance = None

        self.screenshotter = Screenshotter(self.screencap)

        self.on_list_to_recognize = on_list_to_recognize
        self.on_recognition_result = on_recognition_result

    @staticmethod
    async def import_maa(pybinding_dir: Path, bin_dir: Path) -> bool:
        if not pybinding_dir.exists():
            print("Python binding dir does not exist")
            return False

        if not bin_dir.exists():
            print("Bin dir does not exist")
            return False

        pybinding_dir = str(pybinding_dir)
        if pybinding_dir not in sys.path:
            sys.path.insert(0, pybinding_dir)

        try:
            from maa.library import Library
            from maa.toolkit import Toolkit
            from maa.instance import Instance
        except ModuleNotFoundError as err:
            print(err)
            return False

        version = await asyncify(Library.open)(bin_dir)
        if not version:
            print("Failed to open MaaFramework")
            return False

        print(f"Import MAA successfully, version: {version}")

        Toolkit.init_option("./")
        Instance.set_debug_message(True)

        return True

    @staticmethod
    async def detect_adb() -> List["AdbDevice"]:
        from maa.toolkit import Toolkit

        return await Toolkit.adb_devices()

    async def connect_adb(self, path: Path, address: str) -> bool:
        from maa.controller import AdbController

        self.controller = AdbController(path, address)
        connected = await self.controller.connect()
        if not connected:
            print(f"Failed to connect {path} {address}")
            return False

        return True

    async def load_resource(self, dir: Path) -> bool:
        from maa.resource import Resource

        if not self.resource:
            self.resource = Resource()

        return self.resource.clear() and await self.resource.load(dir)

    async def run_task(self, entry: str, param: dict = {}) -> bool:

        from maa.instance import Instance

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

        pil = Image.fromarray(im)
        b, g, r = pil.split()
        return Image.merge("RGB", (r, g, b))

    async def click(self, x, y) -> None:
        if not self.controller:
            return None

        await self.controller.click(x, y)

    def _inst_callback(self, msg: str, detail: dict, arg):
        match msg:
            case "Task.Debug.ListToRecognize":
                print(f"Task.Debug.ListToRecognize: {detail}")
                self.screenshotter.refresh(False)
                if self.on_list_to_recognize:
                    self.on_list_to_recognize(detail["pre_hit_task"], detail["list"])

            case "Task.Debug.RecognitionResult":
                print(f"Task.Debug.RecognitionResult: {detail}")
                reco_detail = self.instance.query_recognition_detail(
                    detail["recognition"]["id"]
                )

                if self.on_recognition_result and reco_detail:
                    self.on_recognition_result(reco_detail)


class Screenshotter(threading.Thread):
    def __init__(self, screencap_func: Callable):
        super().__init__()
        self.source = None
        self.active = False
        self.screencap_func = screencap_func

    def __del__(self):
        self.active = False
        self.source = None

    def run(self):
        while self.active:
            self.refresh()
            time.sleep(0)

    def refresh(self, capture: bool = True):
        im = asyncio.run(self.screencap_func(capture))
        if not im:
            return

        self.source = im

    def start(self):
        self.active = True
        super().start()

    def stop(self):
        self.active = False
