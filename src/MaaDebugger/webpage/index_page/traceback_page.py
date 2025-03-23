import sys
import traceback
from datetime import datetime

from nicegui import ui


class TracebackData:
    """
    key: (name,value,details,record_time)
    """

    data: dict[int, tuple] = {}
    id: int = 0


@ui.page("/traceback/{tb_id}")
def create_traceback_page(tb_id: int):
    if tb_id not in TracebackData.data:
        ui.markdown("## Not Found")
        return

    title = f"Traceback ({tb_id})"
    name, value, details, record_time = TracebackData.data[tb_id]

    ui.page_title(title)
    ui.markdown(f"## {title}")
    ui.separator()
    ui.markdown(f"#### {name}\n\n##### {value}")
    ui.code(details, language="shell").classes("w-full")
    ui.markdown(f"Recorded at: *{record_time}*")


def on_exception(e: Exception):  # When an exception is raised, this function is called
    try:  # Avoid recursion
        details = ""

        page_path = f"/traceback/{TracebackData.id}"
        record_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        e_type, e_value, e_traceback = sys.exc_info()
        tb_info = traceback.format_exception(e_type, e_value, e_traceback)
        for t in tb_info:
            details += t

        TracebackData.data[TracebackData.id] = (
            e_type.__name__ if e_type else "UnknownError",
            e_value,
            details,
            record_time,
        )

        ui.notification(
            f"{e.__class__.__name__}: {e_value}",
            close_button=True,
            type="negative",
            position="bottom-right",
            timeout=None,
            actions=[
                {
                    "label": "VIEW",
                    "color": "white",
                    ":handler": f"() => emitEvent('{page_path}/view-clicked')",
                }
            ],
        )

        ui.on(
            f"{page_path}/view-clicked", lambda: ui.navigate.to(page_path, new_tab=True)
        )

    finally:
        TracebackData.id += 1
