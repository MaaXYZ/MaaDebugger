import sys
import traceback
from datetime import datetime

from nicegui import ui
from nicegui.elements.mixins.value_element import ValueElement


class TracebackData:
    """
    This class is used to store the tracebacks data.

    :data: { key: (name,value,details,record_time) }
    """

    # data
    data: dict[int, tuple[str, str, str, str]] = {}
    id: int = -1
    max_display: int = 100

    # display
    reverse: bool = True
    auto_update: bool = False


class TracaBackElement(ValueElement):
    """
    This class is used to `bind_value_from` `TracebackData`.\n
    When the value of `TracebackData` changes, do something.
    """

    pass


def clear_traceback_data():
    TracebackData.data = {}
    TracebackData.id = 0
    ui.navigate.reload()


def get_data(num: int) -> dict:
    """
    Get the specified amount of data.
    """

    i, data = 0, {}

    for key in sorted(TracebackData.data.keys(), reverse=True):
        i += 1
        if i > num:
            break
        data[key] = TracebackData.data[key]

    return data


def auto_update_page():
    if TracebackData.auto_update:
        ui.navigate.reload()


@ui.page("/traceback/all")
def creata_traceback_all_page():

    TracaBackElement(value=None).bind_value_from(TracebackData, "id").on_value_change(
        auto_update_page
    )

    ui.page_title("Tracebacks")
    with ui.row(align_items="center").classes("w-full"):
        ui.number("Maximum Results to Show", min=1).bind_value(
            TracebackData, "max_display"
        )
        ui.switch("Reverse").bind_value(TracebackData, "reverse").on_value_change(
            ui.navigate.reload
        )
        ui.switch("Auto Update").bind_value(TracebackData, "auto_update").tooltip(
            "When a new exception is recorded, update the page."
        )
        ui.button("Clear Traceback Cache", on_click=clear_traceback_data).props(
            "no-caps"
        )
        ui.separator()

    if not TracebackData.data:
        ui.markdown("## No Tracebacks")
        return

    with ui.list().props("bordered separator").classes("w-full"):
        for key in sorted(
            get_data(TracebackData.max_display), reverse=TracebackData.reverse
        ):
            name, value, _, record_time = TracebackData.data[key]

            with ui.item(
                on_click=lambda id=key: ui.navigate.to(f"/traceback/{id}", True)
            ).classes("w-full"):
                with ui.item_section().props("side"):
                    ui.item_label(str(key))
                with ui.item_section():
                    ui.item_label(record_time)
                with ui.item_section():
                    ui.item_label(name)
                with ui.item_section():
                    ui.item_label(value)


@ui.page("/traceback/{tb_id}")
def create_traceback_page(tb_id: int):
    if tb_id not in TracebackData.data:
        ui.markdown("## Not Found")
        ui.link("Back to Tracebacks Page", "/traceback/all")
        return

    title = f"Traceback ({tb_id})"
    name, value, details, record_time = TracebackData.data[tb_id]

    ui.page_title(title)
    ui.markdown(f"## {title}")
    ui.separator()
    ui.markdown(f"#### {name}\n\n##### {value}")
    ui.code(details, language="shell").classes("w-full")

    ui.separator()

    ui.markdown(f"Recorded at: *{record_time}*")
    ui.link("Back to Tracebacks Page", "/traceback/all")


def on_exception(e: Exception):  # When an exception is raised, this function is called
    try:  # Avoid recursion
        TracebackData.id += 1

        details = ""

        page_path = f"/traceback/{TracebackData.id}"
        record_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        e_type, e_value, e_traceback = sys.exc_info()
        tb_info = traceback.format_exception(e_type, e_value, e_traceback)
        for t in tb_info:
            details += t

        TracebackData.data[TracebackData.id] = (
            e_type.__name__ if e_type else "UnknownError",
            str(e_value),
            details,
            record_time,
        )

        ui.notification(
            f"{e.__class__.__name__}: {e_value}",
            close_button=True,
            type="negative",
            position="bottom-right",
            timeout=10,
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
        pass
