from typing import Any, Literal, Optional, Union

from nicegui import ui, app


@ui.refreshable
def send(
    message: Any,
    with_print: bool = True,
    position: Literal[
        "top-left",
        "top-right",
        "bottom-left",
        "bottom-right",
        "top",
        "bottom",
        "left",
        "right",
        "center",
    ] = "bottom-right",
    close_button: Union[bool, str] = False,
    type: Literal[
        "positive", "negative", "warning", "info", "ongoing", None
    ] = "negative",
    color: Optional[str] = None,
    multi_line: bool = False,
) -> None:
    if not message:
        return

    if with_print:
        print(message)

    ui.notify(
        message,
        position=position,
        close_button=close_button,
        type=type,
        color=color,
        multi_line=multi_line,
    )
