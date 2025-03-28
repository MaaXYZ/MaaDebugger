import argparse
import json
from pathlib import Path
from typing import Optional

from ..port_checker import PortChecker


class ArgParser:
    def __init__(self) -> None:
        # MaaDebugger/args.json
        self.ARGS_PATH = Path(__file__).parent.parent.parent / "args.json"

        self.data = {}
        if self.ARGS_PATH.exists():
            with open(self.ARGS_PATH, "r") as f:
                self.data: dict = json.load(f)
                _data: dict = self.data.copy()  # A copy of data

        self.parser = argparse.ArgumentParser(
            description="A debugger specifically for MaaFramework."
        )

        self._add_argument()
        self._add_dark_group()
        self._add_check_update_group()

        self.args = self.parser.parse_args()

        # The args to be stored should be used through variables rather than functions.
        self.check_update: bool = self.store_check_update()

        # If data has changes, write it to the args.json
        if self.data != _data:
            with open(self.ARGS_PATH, "w") as f:
                json.dump(self.data, f, indent=4)

    def _add_argument(self):
        """
        Add command line arguments to the parser.
        """
        self.parser.add_argument(
            "--port",
            type=int,
            help="Run on which port",
            default=None,
        )
        self.parser.add_argument(
            "--host",
            type=str,
            help="When the value is 'localhost', only the local machine can access it. If you want to change this, you can set value as '0.0.0.0'. (Default: localhost)",
            default="localhost",
        )
        self.parser.add_argument(
            "--hide",
            action="store_true",
            help="DON'T automatically open the UI in a browser tab. (Default: False)",
            default=False,
        )

    def _add_dark_group(self):
        """
        Add command line arguments about dark_mode to the parser.
        """
        group = self.parser.add_mutually_exclusive_group()

        group.add_argument(
            "--dark",
            help="Enable dark mode. (Default: Auto)",
            action="store_true",
            default=None,
        )
        group.add_argument(
            "--light",
            help="Disable dark mode. (Default: Auto)",
            action="store_true",
            default=None,
        )

    def _add_check_update_group(self):
        """
        Add command line arguments about acheck_update to the parser.
        """
        group = self.parser.add_mutually_exclusive_group()

        group.add_argument(
            "--enable_update",
            help="Enable update checking",
            action="store_true",
            default=None,
        )
        group.add_argument(
            "--disable_update",
            help="Disable update checking",
            action="store_true",
            default=None,
        )

    def get_port(self) -> int:
        """
        Determine the port to use based on the provided arguments.
        """
        specified_port: Optional[int] = self.args.port

        if specified_port is not None:
            if PortChecker.is_port_in_use(specified_port):
                print(f"Specified port {specified_port} is in use.")
                return -1
            else:
                port = specified_port
        else:
            port = PortChecker.find_available_port(8011)
        return port

    def get_host(self) -> str:
        """
        When the value is 'localhost' (or '127.0.0.1'), only the local machine can access it.
        If you want to change this,you can set value as '0.0.0.0'
        """

        host = self.args.host

        if host in ["localhost", "127.0.0.1"]:
            print("NOTICE: Only the local machine can access MaaDebugger.")
        else:
            print("WARNING: All devices in the LAN can access MaaDebugger.")
        return host

    def get_hide(self) -> bool:
        """
        NOTICE: ui.run(show = not self.args.hide)
        """
        hide = self.args.hide

        return not bool(hide)

    def get_dark(self) -> Optional[bool]:
        dark = self.args.dark
        light = self.args.light

        if dark:
            return True
        elif light:
            return False
        else:
            return None

    def store_check_update(self) -> bool:
        if self.args.enable_update:
            self.data["update"] = True
            return True

        elif self.args.disable_update:
            self.data["update"] = False
            return False

        else:
            return self.data.get("update", True)
