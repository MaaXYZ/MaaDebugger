from pathlib import Path
from typing import Optional


from nicegui import ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page
from .webpage.index_page import runtime_control
from .utils import args
from .maafw import maafw
from .__version__ import version

TITLE = f"Maa Debugger ({version})"


class MaaDebugger:
    @staticmethod
    def init_page() -> None:
        index_page.main()

    @staticmethod
    def set_pagination(per_page: Optional[int]) -> None:
        """
        Set the number of items per page for the pagination in the MaaDebugger UI.

        :param per_page: The number of items to display per page.
        """
        runtime_control.PER_PAGE_ITEM_NUM = per_page

    @staticmethod
    def use_cpu() -> None:
        maafw.use_cpu = True

    @classmethod
    def run_in_cli(cls) -> None:
        """This is a **internal method** to run the MaaDebugger in CLI mode. For user, please use `MaaDebugger.run()` instead."""
        host = args.get_host()
        port = args.get_port() or 8011
        show = args.get_hide()
        dark = args.get_dark()

        cls.init_page()

        ui.run(
            title=TITLE,
            storage_secret="maadbg",
            reload=False,
            host=host,
            port=find_open_port(port, end_port=port + 100),
            show=show,
            dark=dark,
            favicon=Path(__file__).parent / "maa.ico",
        )

    @classmethod
    def run(
        cls,
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
        cls.init_page()

        ui.run(
            title=TITLE,
            storage_secret="maadbg",
            reload=False,
            host=host,
            port=find_open_port(port, end_port=port + 100),
            show=show,
            dark=dark,
            favicon=Path(__file__).parent / "maa.ico",
            **kwargs,
        )
