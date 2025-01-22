import argparse
from typing import Optional

from ..port_checker import PortChecker


class ArgParser:
    def __init__(self) -> None:
        self.parser = argparse.ArgumentParser(
            description="A debugger specifically for MaaFramework."
        )
        self._add_argument()
        self.args = self.parser.parse_args()

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
            help="When the value is 'localhost', only the local machine can access it. If you want to change this,you can set value as '0.0.0.0'. (Default: localhost)",
            default="localhost",
        )
        self.parser.add_argument(
            "--hide",
            type=str,
            help="Whether NOT automatically open the UI in a browser tab. (Default: False)",
            default=False,
        )
        self.parser.add_argument(
            "--dark",
            type=str,
            help="Whether enabel dark mode.None means auto. (Default: None)",
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
        NOTICE: ui.run(show=not self.args.hide)
        """
        hide = self.args.hide

        if hide in ["True", "true"]:
            return True
        else:
            return False

    def get_dark(self) -> Optional[bool]:
        dark = self.args.dark

        if dark in ["True", "true"]:
            return True
        elif dark in ["False", "false"]:
            return False
        else:
            return None
