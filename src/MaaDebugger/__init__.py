from pathlib import Path
from typing import Optional


from nicegui import ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page
from .webpage.index_page import runtime_control
from .maafw import maafw
from .utils import update_checker
from .__version__ import version

TITLE = f"Maa Debugger ({version})"


class MaaDebugger:
    check_update: bool = True

    @staticmethod
    def set_device(device_id: int = -1) -> None:
        """
        Set the device ID. For more information, please see `define.MaaInferenceDeviceEnum` .

        :param device_id: The ID of the device to be used.
        """
        if type(device_id) != int or device_id < -2:
            raise ValueError("device_id must be an integer greater than or equal to -2")

        maafw.device_id = device_id

    @staticmethod
    def set_pagination(per_page: Optional[int]) -> None:
        """
        Set the number of items per page for the pagination in the MaaDebugger UI.

        :param per_page: The number of items to display per page.
        """
        runtime_control.PER_PAGE_ITEM_NUM = per_page

    @classmethod
    def disable_update_checking(cls):
        """Disable the automatic update checking."""
        cls.check_update = False

    @classmethod
    def run(
        cls,
        *,
        host: str = "localhost",
        port: int = 8011,
        show: bool = True,
        dark: Optional[bool] = None,
        **kwargs,
    ) -> None:
        """
        Run MaaDebugger.

        :param host: The host to run the server on. When the value is `localhost`, only the **local machine** can access it. If you want to change this, you can set value as `0.0.0.0`.
        :param port: Run on which port.
        :param show: If automatically open the UI in a browser tab.
        :param dark: Enable dark mode. If set to `None`, it will be auto-detected based on the system settings.
        :param **kwargs: Additional keyword arguments to pass to `ui.run()`. For more information, please see https://nicegui.io/documentation/run#ui_run
        """
        index_page.main()

        if cls.check_update:
            ui.timer(2, update_checker.main, once=True)  # Check update

        ui.run(
            title=TITLE,
            storage_secret="maadbg",
            reload=False,
            favicon=Path(__file__).parent / "maa.ico",
            host=host,
            port=find_open_port(port, end_port=port + 100),
            show=show,
            dark=dark,
            **kwargs,
        )
