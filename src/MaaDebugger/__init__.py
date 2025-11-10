from pathlib import Path
from typing import Optional


from nicegui import app, ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page  # noqa: F401
from .webpage.traceback_page import on_exception
from .utils import update_checker
from .maafw import maafw
from .__version__ import version
from .assets import FAVICON_PATH


APP_TITLE = f"Maa Debugger (v{version})"


class MaaDebugger:
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
        print(f"MaaFramework version: {maafw.version}")
        print(f"Log located at {Path.cwd()/'debug'/'maa.log'}\n")

        if host in ["localhost", "127.0.0.1", "::1"]:
            print("NOTICE: Only the local machine can access MaaDebugger.")
        else:
            print("WARNING: All devices in the LAN can access MaaDebugger.")

        index_page.main()
        ui.timer(2, update_checker.main, once=True)  # Check update
        app.on_exception(on_exception)

        ui.run(
            title=APP_TITLE,
            storage_secret="maadbg",
            reload=False,
            favicon=FAVICON_PATH,
            host=host,
            port=find_open_port(port, end_port=port + 100),
            show=show,
            dark=dark,
            **kwargs,
        )
