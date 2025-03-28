import json
from pathlib import Path
from typing import Literal

from nicegui import ui

from ..traceback_page import TracebackData
from ...maafw import maafw
from ...utils.infos import Infos


class DebugData:
    auto_clear_maafw_cache: bool = True


def export_maafw_log():
    log_path = Path.cwd() / "debug/maa.log"
    if not log_path.exists():
        ui.notify(
            f"{log_path} does not exist.", position="bottom-right", type="warning"
        )
        return
    ui.download(log_path, "maa.log")


async def clear_maafw_cache():
    if await maafw.clear_cache():
        ui.notify("Cleared.", position="bottom-right", type="info")


def export_traceback_data():
    data = json.dumps(TracebackData.data, indent=4, ensure_ascii=False).encode()
    ui.download(data, "traceback_data.json")


def copy_data(data: str):
    ui.clipboard.write(data)
    ui.notify("Copied.", position="bottom-right", type="info")


def save_data(
    type: Literal["yaml", "json", "markdown"],
    mode: Literal["copy", "download"],
    elements: tuple[ui.input],
):
    data = ""
    if type == "yaml":
        for e in elements:
            data += f"{e.props['label']}: {e.value}\n"
    elif type == "json":
        for e in elements:
            _data = {e.props["label"]: e.value for e in elements}
        data = json.dumps(_data, indent=4, ensure_ascii=False)
    elif type == "markdown":
        for e in elements:
            data += f"#### {e.props['label']}\n{e.value}\n\n"

    if mode == "copy":
        copy_data(data)
    elif mode == "download":
        ui.download(data.encode(), f"infos.{type}")


def create_info_dialog() -> ui.dialog:
    dialog = ui.dialog()
    with dialog, ui.card():
        ui.markdown("#### Infos")
        ui.separator()

        with ui.row().classes("w-full"):
            cwd = ui.input("CWD").bind_value_from(
                Infos, "CWD", backward=lambda x: str(x)
            )
            debug_dir = ui.input("Debug Directory").bind_value_from(
                Infos, "DEBUG_DIR", backward=lambda x: str(x)
            )
        with ui.row().classes("w-full"):  # OS infos
            os_info = ui.input("OS Info").bind_value_from(Infos, "OS_INFO")
            machine = ui.input("Machine").bind_value_from(Infos, "MACHINE")
        with ui.row().classes("w-full"):
            py_ver = ui.input("Python Version").bind_value_from(Infos, "PYTHON_VERSION")
            dbg_ver = ui.input("Debugger Version").bind_value_from(Infos, "DBG_VERSION")
        with ui.row():  # deps version
            fw_ver = ui.input("MaaFW Version").bind_value_from(Infos, "MAAFW_VERSION")
            nice_ver = ui.input("NiceGUI Version").bind_value_from(
                Infos, "NICEGUI_VERSION"
            )

        elements = (cwd, debug_dir, os_info, machine, py_ver, dbg_ver, fw_ver, nice_ver)
        for e in elements:
            e.on("click", lambda e=e: copy_data(e.value))  # type: ignore

        with ui.row(align_items="center").classes("justify-center").classes("w-full"):
            with ui.dropdown_button(
                "Copy",
                split=True,
                on_click=lambda e=elements: save_data("json", "copy", e),  # type: ignore
            ).props("no-caps"):
                ui.item(
                    "Copy as Yaml",
                    on_click=lambda e=elements: save_data("yaml", "copy", e),  # type: ignore
                ).props("no-caps")
                ui.item(
                    "Copy as Markdown",
                    on_click=lambda e=elements: save_data("markdown", "copy", e),  # type: ignore
                ).props("no-caps")
            with ui.dropdown_button(
                "Export",
                split=True,
                on_click=lambda e=elements: save_data("json", "download", e),  # type: ignore
            ).props("no-caps"):
                ui.item(
                    "Export as Yaml",
                    on_click=lambda e=elements: save_data("yaml", "download", e),  # type: ignore
                ).props("no-caps")
                ui.item(
                    "Export as Markdown",
                    on_click=lambda e=elements: save_data("markdown", "download", e),  # type: ignore
                ).props("no-caps")

    return dialog


@ui.page("/debug")
def create_debug_page():
    ui.page_title("Debug")
    ui.markdown("## Debug Page")
    ui.separator()

    info_dialog = create_info_dialog()

    with ui.card().classes("w-full"):
        ui.markdown("#### MaaFramework")
        with ui.row():
            ui.button("Export Log", on_click=export_maafw_log).props("no-caps")
            ui.button("Clear Cache", on_click=clear_maafw_cache).props("no-caps")
            ui.switch("Auto Clear Cache").bind_value(
                DebugData, "auto_clear_maafw_cache"
            )

    with ui.card().classes("w-full"):
        ui.markdown("#### Feature Test")
        with ui.row():
            with ui.dropdown_button("Raise Error").props("no-caps"):
                ui.item("Raise Json Error", on_click=lambda: json.loads("{")).props(
                    "no-caps"
                )
                ui.item("Raise Path Error", on_click=lambda: Path(None)).props("no-caps")  # type: ignore
            ui.button("Export Traceback Data", on_click=export_traceback_data).props(
                "no-caps"
            )

    with ui.card().classes("w-full"):
        ui.markdown("#### Others")
        with ui.row():
            ui.button("Info", on_click=info_dialog.open).props("no-caps")
