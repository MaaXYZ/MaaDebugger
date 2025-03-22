import sys
import traceback
from datetime import datetime

from nicegui import ui

tb_number = 0


def on_exception(e: Exception):  # When an exception is raised, this function is called
    try:  # Avoid recursion
        global tb_number
        log = ""

        page_path = f"/traceback/{tb_number}"
        page_title = f"Traceback ({tb_number})"
        record_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        e_type, e_value, e_traceback = sys.exc_info()
        if not all([e_type, e_value, e_traceback]):
            return

        tb_info = traceback.format_exception(e_type, e_value, e_traceback)
        for t in tb_info:
            log += t

        @ui.page(page_path, title=page_title)
        def create_traceback_page():
            ui.markdown(f"## {page_title}")
            ui.separator()
            ui.markdown(f"#### {e_type.__name__}\n\n##### {e_value}")
            ui.code(log, language="shell").classes("w-full")
            ui.markdown(f"Recorded at: *{record_time}*")

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

    except:
        pass

    finally:
        tb_number += 1
