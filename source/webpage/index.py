from nicegui import app, ui
from pathlib import Path

from source.control.runner import *
from .components.status_indicator import Status, StatusIndicator


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


class GlobalStatus:
    maa_imported: Status = Status.PENDING
    adb_connected: Status = Status.PENDING
    adb_detected: Status = Status.PENDING  # not required
    res_loaded: Status = Status.PENDING
    task_started: Status = Status.PENDING


async def import_maa_control():

    StatusIndicator(GlobalStatus, "maa_imported")

    pybinding_input = (
        ui.input(
            "MaaFramework Python Binding Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64/binding/Python",
        )
        .props("size=60")
        .bind_value(app.storage.general, "maa_pybinding")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s != Status.SUCCESS
        )
    )
    bin_input = (
        ui.input(
            "MaaFramework Binary Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64/bin",
        )
        .props("size=60")
        .bind_value(app.storage.general, "maa_bin")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s != Status.SUCCESS
        )
    )

    import_button = ui.button(
        "Import", on_click=lambda: on_click_import()
    ).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s != Status.SUCCESS
    )

    async def on_click_import():
        GlobalStatus.maa_imported = Status.RUNNING

        if not pybinding_input.value or not bin_input.value:
            GlobalStatus.maa_imported = Status.FAILURE
            return

        imported = await import_maa(Path(pybinding_input.value), Path(bin_input.value))
        if not imported:
            GlobalStatus.maa_imported = Status.FAILURE
            return

        GlobalStatus.maa_imported = Status.SUCCESS


async def connect_adb_control():
    StatusIndicator(GlobalStatus, "adb_connected")

    adb_path_input = (
        ui.input(
            "ADB Path",
            placeholder="eg: C:/adb.exe",
        )
        .props("size=60")
        .bind_value(app.storage.general, "adb_path")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
        )
    )
    adb_address_input = (
        ui.input(
            "ADB Address",
            placeholder="eg: 127.0.0.1:5555",
        )
        .props("size=30")
        .bind_value(app.storage.general, "adb_address")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
        )
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_connect(),
    ).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )

    ui.button(
        "Detect",
        on_click=lambda: on_click_detect(),
    ).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )

    devices_select = ui.select(
        {}, on_change=lambda e: on_change_devices_select(e)
    ).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )
    StatusIndicator(GlobalStatus, "adb_detected").label().bind_visibility_from(
        GlobalStatus,
        "adb_detected",
        backward=lambda s: s == Status.RUNNING or s == Status.FAILURE,
    )

    async def on_click_connect():
        GlobalStatus.adb_connected = Status.RUNNING

        if not adb_path_input.value or not adb_address_input.value:
            GlobalStatus.adb_connected = Status.FAILURE
            return

        connected = await connect_adb(
            Path(adb_path_input.value), adb_address_input.value
        )
        if not connected:
            GlobalStatus.adb_connected = Status.FAILURE
            return

        GlobalStatus.adb_connected = Status.SUCCESS

    async def on_click_detect():
        GlobalStatus.adb_detected = Status.RUNNING

        devices = await detect_adb()
        options = {}
        for d in devices:
            v = (d.adb_path, d.address)
            l = d.name + " " + d.address
            options[v] = l

        devices_select.options = options
        devices_select.update()
        if not options:
            GlobalStatus.adb_detected = Status.FAILURE

        devices_select.value = next(iter(options))
        GlobalStatus.adb_detected = Status.SUCCESS

    def on_change_devices_select(e):
        adb_path_input.value = str(e.value[0])
        adb_address_input.value = e.value[1]


async def load_resource_control():
    StatusIndicator(GlobalStatus, "res_loaded")

    dir_input = (
        ui.input(
            "Resource Directory",
            placeholder="eg: C:/M9A/assets/resource/base",
        )
        .props("size=60")
        .bind_value(app.storage.general, "resource_dir")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
        )
    )

    ui.button(
        "Load",
        on_click=lambda: on_click_load(),
    ).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )

    async def on_click_load():
        GlobalStatus.res_loaded = Status.RUNNING

        if not dir_input.value:
            GlobalStatus.res_loaded = Status.FAILURE
            return

        loaded = await load_resource(Path(dir_input.value))
        if not loaded:
            GlobalStatus.res_loaded = Status.FAILURE
            return

        GlobalStatus.res_loaded = Status.SUCCESS


async def run_task_control():
    StatusIndicator(GlobalStatus, "task_started")

    entry_input = (
        ui.input(
            "Task Entry",
            placeholder="eg: StartUp",
        )
        .props("size=30")
        .bind_value(app.storage.general, "task_entry")
        .bind_enabled_from(
            GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
        )
    )

    ui.button("Start", on_click=lambda: on_click_start()).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )
    ui.button("Stop", on_click=lambda: on_click_stop()).bind_enabled_from(
        GlobalStatus, "maa_imported", backward=lambda s: s == Status.SUCCESS
    )

    async def on_click_start():
        GlobalStatus.task_started = Status.RUNNING

        if not entry_input.value:
            GlobalStatus.task_started = Status.FAILURE
            return

        run = await run_task(entry_input.value)
        if not run:
            GlobalStatus.task_started = Status.FAILURE
            return

        GlobalStatus.task_started = Status.SUCCESS

    async def on_click_stop():
        stopped = await stop_task()
        if not stopped:
            GlobalStatus.task_started = Status.FAILURE
            return

        GlobalStatus.task_started = Status.PENDING
