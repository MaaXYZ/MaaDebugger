from pathlib import Path
from typing import List, Optional
from asyncer import asyncify
from PIL import Image
import sys
import asyncio
import threading
import time


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
    except ModuleNotFoundError as err:
        print(err)
        return False

    version = await asyncify(Library.open)(bin_dir)
    if not version:
        print("Failed to open MaaFramework")
        return False

    print(f"Import MAA successfully, version: {version}")

    Toolkit.init_option("./")
    Library.set_debug_message(True)

    return True


async def detect_adb() -> List["AdbDevice"]:
    from maa.toolkit import Toolkit

    return await Toolkit.adb_devices()


resource = None
controller = None
instance = None


async def connect_adb(path: Path, address: str) -> bool:
    global controller

    from maa.controller import AdbController

    controller = AdbController(path, address)
    connected = await controller.connect()
    if not connected:
        print(f"Failed to connect {path} {address}")
        return False

    return True


async def load_resource(dir: Path) -> bool:
    global resource

    from maa.resource import Resource

    if not resource:
        resource = Resource()

    return resource.clear() and await resource.load(dir)


ui_callback = None


def inst_callback(msg: str, detail: dict, arg):
    match msg:
        case "Task.Debug.ListToRecognize":
            print(f"Task.Debug.ListToRecognize: {detail}")
            screenshotter.refresh(False)
            if ui_callback:
                ui_callback(detail["latest_hit"], detail["list"])


async def run_task(entry: str, param: dict = {}) -> bool:
    global controller, resource, instance

    from maa.instance import Instance

    if not instance:
        instance = Instance(callback=inst_callback)

    instance.bind(resource, controller)
    if not instance.inited:
        print("Failed to init MaaFramework instance")
        return False

    return await instance.run_task(entry, param)


async def stop_task():
    global instance

    if not instance:
        return

    await instance.stop()


async def screencap(capture: bool = True) -> Optional[Image.Image]:
    global controller
    if not controller:
        return None

    im = await controller.screencap(capture)
    if im is None:
        return None

    pil = Image.fromarray(im)
    b, g, r = pil.split()
    return Image.merge("RGB", (r, g, b))


class Screenshotter(threading.Thread):
    def __init__(self):
        super().__init__()
        self.source = None
        self.active = False

    def __del__(self):
        self.active = False
        self.source = None

    def run(self):
        while self.active:
            self.refresh()
            time.sleep(0)

    def refresh(self, capture: bool = True):
        im = asyncio.run(screencap(capture))
        if not im:
            return

        self.source = im

    def start(self):
        self.active = True
        super().start()

    def stop(self):
        self.active = False


screenshotter = Screenshotter()


async def click(x, y) -> None:
    global controller
    if not controller:
        return None

    await controller.click(x, y)
