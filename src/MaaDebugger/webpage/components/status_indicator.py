from enum import Enum, auto

from nicegui import ui


class Status(Enum):
    PENDING = auto()
    RUNNING = auto()
    SUCCESS = auto()
    FAILURE = auto()


class StatusIndicator:
    def __init__(self, target_object, target_name):
        self._label = ui.label().bind_text_from(
            target_object,
            target_name,
            backward=lambda s: StatusIndicator._text_backward(s),
        )

    def label(self):
        return self._label

    @staticmethod
    def _text_backward(status: Status) -> str:
        match status:
            case Status.PENDING:
                return "🟡"
            case Status.RUNNING:
                return "👀"
            case Status.SUCCESS:
                return "✅"
            case Status.FAILURE:
                return "❌"
