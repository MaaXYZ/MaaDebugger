import re
from asyncify import asyncify
from pathlib import Path
from typing import Callable, List, Optional, Union, Dict

from maa.controller import AdbController, Win32Controller
from maa.tasker import Tasker, RecognitionDetail, NotificationHandler
from maa.resource import Resource
from maa.toolkit import Toolkit, AdbDevice, DesktopWindow
from PIL import Image

from ..utils import cvmat_to_image

import importlib.util


class MaaFW:

    resource: Optional[Resource]
    controller: Union[AdbController, Win32Controller, None]
    tasker: Optional[Tasker]
    notification_handler: Optional[NotificationHandler]

    def __init__(self):
        Toolkit.init_option("./")
        Tasker.set_debug_mode(True)

        self.resource = None
        self.controller = None
        self.tasker = None

        self.screenshotter = Screenshotter(self.screencap)
        self.notification_handler = None

        self.custom_list: Dict[str, List] = None  # 自定义动作\识别器列表

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
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            print(f"Failed to connect {path} {address}")
            return False

        return True

    @asyncify
    def connect_win32hwnd(
        self, hwnd: Union[int, str], screencap_method: int, input_method: int
    ) -> bool:
        if isinstance(hwnd, str):
            hwnd = int(hwnd, 16)

        self.controller = Win32Controller(
            hwnd, screencap_method=screencap_method, input_method=input_method
        )
        connected = self.controller.post_connection().wait().succeeded
        if not connected:
            print(f"Failed to connect {hwnd}")
            return False

        return True

    @asyncify
    def load_resource(self, dir: List[Path]) -> bool:
        if not self.resource:
            self.resource = Resource()

        self.resource.clear()
        for d in dir:
            if not d.exists():
                return False

            status = self.resource.post_bundle(d).wait().succeeded
            if not status:
                return False
        self.resource_dir = dir
        return True

    @asyncify
    def run_task(self, entry: str, pipeline_override: dict = {}) -> bool:
        self.custom_list = {"action": [], "recognition": []}
        if not self.tasker:
            self.tasker = Tasker(notification_handler=self.notification_handler)

        if not self.resource or not self.controller:
            print("Resource or Controller not initialized")
            return False

        self.tasker.bind(self.resource, self.controller)

        self.resource.clear_custom_recognition()
        self.resource.clear_custom_action()
        for d in self.resource_dir:
            d = d / "custom"
            if not d.exists():
                continue

            self.load_custom_objects(d)
        print(self.custom_list)
        if not self.tasker.inited:
            print("Failed to init MaaFramework tasker")
            return False

        return self.tasker.post_task(entry, pipeline_override).wait().succeeded

    def load_custom_objects(self, custom_dir):
        custom_path = Path(custom_dir)
        if not custom_path.exists():
            return False

        if not list(custom_path.iterdir()):
            return False

        errors = []
        for module_type in ["action", "recognition"]:
            module_type_path = custom_path / module_type
            if not module_type_path.exists():
                continue

            for subdir in module_type_path.iterdir():
                if subdir.is_dir():
                    entry_file = subdir / "main.py"
                    if not entry_file.exists():
                        continue

                    try:
                        module_name = subdir.name
                        if self._load_module(module_type, module_name, entry_file):
                            self.custom_list[module_type].append(module_name)
                        else:
                            errors.append(
                                f"Failed to load {module_name} {module_type}."
                            )
                    except Exception as e:
                        errors.append(f"Error loading {module_name} {module_type}: {e}")

        if errors:
            print("\n".join(errors))
            return False

        return True

    def _load_module(self, module_type, module_name, entry_file):
        spec = importlib.util.spec_from_file_location(module_name, entry_file)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        register_func = getattr(self.resource, f"register_custom_{module_type}")
        return register_func(f"{module_name}", getattr(module, module_name)())

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
