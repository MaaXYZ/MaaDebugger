import argparse
from typing import Optional


class ArgParser:
    def __init__(self) -> None:
        self.parser = argparse.ArgumentParser(
            description="A debugger specifically for MaaFramework."
        )

        self._add_argument()
        self._add_dark_group()

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

    def get_port(self) -> int:
        """
        Determine the port to use based on the provided arguments.
        """
        return self.args.port or 8011

    def get_host(self) -> str:
        """
        When the value is 'localhost' (or '127.0.0.1'), only the local machine can access it.
        If you want to change this,you can set value as '0.0.0.0'
        """
        return self.args.host

    def get_show(self) -> bool:
        """
        NOTICE: ui.run(show = not self.args.hide)
        """
        return not bool(self.args.hide)

    def get_dark(self) -> Optional[bool]:
        dark = self.args.dark
        light = self.args.light

        if dark:
            return True
        elif light:
            return False
        else:
            return None
