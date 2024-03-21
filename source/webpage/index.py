from nicegui import app, ui
from pathlib import Path

from source.control.runner import *


@ui.page("/")
async def index():
    ui.dark_mode()  # auto dark mode

    with ui.row().style("align-items: center;"):
        await import_maa_control()

    ui.separator()

    with ui.row().style("align-items: center;"):
        await connect_adb_control()
    with ui.row().style("align-items: center;"):
        await load_resource_control()
    with ui.row().style("align-items: center;"):
        await run_task_control()

    ui.separator()


class StatusIndicator:
    def __init__(self):
        self.pending()
        ui.label().bind_text(self, "text")

    def pending(self):
        self.text = "🟡"
        return self

    def success(self):
        self.text = "✅"
        return self

    def failure(self):
        self.text = "❌"
        return self

    def running(self):
        self.text = "👀"
        return self

    def hide(self):
        self.text = "  "
        return self


async def import_maa_control():
    status = StatusIndicator()

    pybinding_input = (
        ui.input(
            "MaaFramework Python Binding Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64/binding/Python",
            on_change=lambda: status.pending(),
        )
        .props("size=60")
        .bind_value(app.storage.general, "maa_pybinding")
    )
    bin_input = (
        ui.input(
            "MaaFramework Binary Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64/bin",
            on_change=lambda: status.pending(),
        )
        .props("size=60")
        .bind_value(app.storage.general, "maa_bin")
    )

    import_button = ui.button("Import", on_click=lambda: on_click_import())

    async def on_click_import():
        status.running()

        if not pybinding_input.value or not bin_input.value:
            status.failure()
            return

        imported = await import_maa(Path(pybinding_input.value), Path(bin_input.value))
        if not imported:
            status.failure()

        status.success()

        pybinding_input.disable()
        bin_input.disable()
        import_button.disable()


async def connect_adb_control():
    status = StatusIndicator()

    adb_path_input = (
        ui.input(
            "ADB Path",
            placeholder="eg: C:/adb.exe",
            on_change=lambda: status.pending(),
        )
        .props("size=60")
        .bind_value(app.storage.general, "adb_path")
    )
    adb_address_input = (
        ui.input(
            "ADB Address",
            placeholder="eg: 127.0.0.1:5555",
            on_change=lambda: status.pending(),
        )
        .props("size=30")
        .bind_value(app.storage.general, "adb_address")
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_connect(),
    )
    ui.button(
        "Detect",
        on_click=lambda: on_click_detect(),
    )
    devices_select = ui.select({}, on_change=lambda e: on_change_devices_select(e))
    detect_status = StatusIndicator().hide()

    async def on_click_connect():
        status.running()

        if not adb_path_input.value or not adb_address_input.value:
            status.failure()
            return

        connected = await connect_adb(
            Path(adb_path_input.value), adb_address_input.value
        )
        if not connected:
            status.failure()

        status.success()

    async def on_click_detect():
        detect_status.running()

        devices = await detect_adb()
        options = {}
        for d in devices:
            v = (d.adb_path, d.address)
            l = d.name + " " + d.address
            options[v] = l

        devices_select.options = options
        devices_select.update()
        if options:
            devices_select.value = next(iter(options))

        detect_status.hide()

    def on_change_devices_select(e):
        adb_path_input.value = str(e.value[0])
        adb_address_input.value = e.value[1]


async def load_resource_control():
    status = StatusIndicator()

    dir_input = (
        ui.input(
            "Resource Directory",
            placeholder="eg: C:/M9A/assets/resource/base",
            on_change=lambda: status.pending(),
        )
        .props("size=60")
        .bind_value(app.storage.general, "resource_dir")
    )

    ui.button(
        "Load",
        on_click=lambda: on_click_load(),
    )

    async def on_click_load():
        status.running()

        if not dir_input.value:
            status.failure()
            return

        loaded = await load_resource(Path(dir_input.value))
        if not loaded:
            status.failure()

        status.success()


async def run_task_control():
    status = StatusIndicator()

    entry_input = (
        ui.input(
            "Task Entry",
            placeholder="eg: StartUp",
            on_change=lambda: status.pending(),
        )
        .props("size=30")
        .bind_value(app.storage.general, "task_entry")
    )

    ui.button("Start", on_click=lambda: on_click_start())
    ui.button("Stop", on_click=lambda: on_click_stop())

    async def on_click_start():
        status.running()

        if not entry_input.value:
            status.failure()
            return

        run = await run_task(entry_input.value)
        if not run:
            status.failure()

        status.success()

    async def on_click_stop():
        stopped = await stop_task()
        if not stopped:
            status.failure()

        status.pending()
