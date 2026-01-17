import argparse
from typing import Optional


class ArgParser:
    parser = argparse.ArgumentParser(
        description="A debugger specifically for MaaFramework."
    )

    @classmethod
    def init(cls):
        cls._add_argument()
        cls._add_dark_group()
        cls.args = cls.parser.parse_args()

    @classmethod
    def _add_argument(cls):
        """
        Add command line arguments to the parser.
        """
        cls.parser.add_argument(
            "--port",
            type=int,
            help="Run on which port",
            default=None,
        )
        cls.parser.add_argument(
            "--host",
            type=str,
            help="When the value is 'localhost', only the local machine can access it. If you want to change this, you can set value as '0.0.0.0'. (Default: localhost)",
            default="localhost",
        )
        cls.parser.add_argument(
            "--hide",
            action="store_true",
            help="DON'T automatically open the UI in a browser tab. (Default: False)",
            default=False,
        )
        cls.parser.add_argument(
            "--DEBUG",
            action="store_true",
            help="Enable Debug mode. (Default: False)",
            default=False,
        )

    @classmethod
    def _add_dark_group(cls):
        """
        Add command line arguments about dark_mode to the parser.
        """
        group = cls.parser.add_mutually_exclusive_group()

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

    @classmethod
    def get_port(cls) -> int:
        """
        Determine the port to use based on the provided arguments.
        """
        return cls.args.port or 8011

    @classmethod
    def get_host(cls) -> str:
        """
        When the value is 'localhost' (or '127.0.0.1'), only the local machine can access it.
        If you want to change this,you can set value as '0.0.0.0'
        """
        return cls.args.host

    @classmethod
    def get_show(cls) -> bool:
        """
        NOTICE: ui.run(show = not cls.args.hide)
        """
        return not bool(cls.args.hide)

    @classmethod
    def get_dark(cls) -> Optional[bool]:
        dark = cls.args.dark
        light = cls.args.light

        if dark:
            return True
        elif light:
            return False
        else:
            return None

    @classmethod
    def get_debug(cls) -> bool:
        return bool(cls.args.DEBUG)


ArgParser.init()
