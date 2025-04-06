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
    id: int = 0

    # display
    max_display: int = 100
    reverse: bool = True
    auto_update: bool = True

    @classmethod
    def reset(cls):
        """Reset the data and id."""
        cls.data.clear()
        cls.id = 0


class TraceBackElement(ValueElement):
    """
    This class is used to `bind_value_from` `TracebackData`.\n
    When the value of `TracebackData` changes, do something.
    """

    pass


def clear_traceback_data():
    TracebackData.reset()
    create_traceback_list.refresh()
    ui.notify("Cleared.", position="bottom-right", type="info")


def get_data(num: int) -> dict:
    """
    Get the specified amount of data.
    """
    if num is None:
        num = 0

    i, data = 0, {}

    for key in sorted(TracebackData.data.keys(), reverse=True):
        i += 1
        if i > num:
            break
        data[key] = TracebackData.data[key]

    return data


def auto_update_list():
    if TracebackData.auto_update:
        create_traceback_list.refresh()


@ui.refreshable
def create_traceback_list():
    if not TracebackData.data:
        ui.markdown("### No Tracebacks")
        return

    with ui.list().props("bordered separator").classes("w-full"):
        for key in sorted(
            get_data(TracebackData.max_display), reverse=TracebackData.reverse
        ):
            name, value, _, record_time = TracebackData.data[key]

            with (
                ui.link(target=f"/traceback/{key}", new_tab=True)
                .classes("!no-underline")  # Hide the underline
                .classes(
                    "dark: text-black"  # In light mode, set the font color to black (it is blue by default due to ui.link)
                ),
                ui.item().classes("w-full"),
            ):
                with ui.item_section().props("side"):
                    ui.item_label(str(key))
                with ui.item_section():
                    ui.item_label(record_time)
                with ui.item_section():
                    ui.item_label(name)
                with ui.item_section():
                    ui.item_label(value)


@ui.page("/traceback")
def create_all_traceback_page():
    TraceBackElement(value=None).bind_value_from(TracebackData, "id").on_value_change(
        auto_update_list
    )

    ui.page_title("Tracebacks")
    ui.markdown("## Tracebacks")

    with ui.row(align_items="center").classes("w-full"):
        ui.number("Maximum Results to Show", min=1, precision=0).bind_value(
            TracebackData, "max_display"
        ).on_value_change(create_traceback_list.refresh)
        ui.switch("Reverse").bind_value(TracebackData, "reverse").on_value_change(
            create_traceback_list.refresh
        )
        ui.switch("Auto Update").bind_value(TracebackData, "auto_update").tooltip(
            "When a new exception is recorded, should the list be automatically updated"
        )
        ui.button("Clear Traceback Cache", on_click=clear_traceback_data).props(
            "no-caps"
        )
        ui.separator()
        create_traceback_list()


@ui.page("/traceback/{tb_id}")
def create_traceback_page(tb_id: int):
    if tb_id not in TracebackData.data:
        ui.markdown("## Not Found")
        ui.link("Back to Tracebacks Page", "/traceback")
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
    ui.link("Back to Tracebacks Page", "/traceback")


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
        TracebackData.id += 1
