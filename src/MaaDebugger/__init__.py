from typing import Optional

from maa.custom_action import CustomAction
from maa.custom_recognition import CustomRecognition

from . import main
from .maafw import maafw


class MaaDebugger:

    @staticmethod
    def run(
        *,
        host: str = "localhost",
        port: int = 8011,
        show: bool = True,
        dark: Optional[bool] = None,
    ):
        main.run(
            host=host,
            port=port,
            show=show,
            dark=dark,
        )

    @staticmethod
    def register_custom_action(name: str, action: CustomAction):
        maafw.register_custom_action(name, action)

    @staticmethod
    def register_custom_recognition(name: str, recognition: CustomRecognition):
        maafw.register_custom_recognition(name, recognition)
