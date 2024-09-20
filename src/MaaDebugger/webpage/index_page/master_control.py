import json
import asyncio
from pathlib import Path

from maa.define import MaaWin32ScreencapMethodEnum, MaaWin32InputMethodEnum
from nicegui import app, binding, ui

from ...maafw import maafw
from ...webpage.components.status_indicator import Status, StatusIndicator

binding.MAX_PROPAGATION_TIME = 1


class GlobalStatus:
    ctrl_connecting: Status = Status.PENDING
    ctrl_detecting: Status = Status.PENDING  # not required
    res_loading: Status = Status.PENDING
    task_running: Status = Status.PENDING


async def main():
    with ui.row():
        with ui.column():
            await connect_control()
            with ui.row().style("align-items: center;"):
                await load_resource_control()
            with ui.row().style("align-items: center;"):
                await run_task_control()

        with ui.column():
            await screenshot_control()


async def connect_control():
    with ui.tabs() as tabs:
        adb = ui.tab("Adb")
        win32 = ui.tab("Win32")

    with ui.tab_panels(tabs, value="Adb").bind_value(
        app.storage.general, "controller_type"
    ):
        with ui.tab_panel(adb):
            with ui.row().style("align-items: center;"):
                await connect_adb_control()
        with ui.tab_panel(win32):
            with ui.row().style("align-items: center;"):
                await connect_win32_control()


async def connect_adb_control():
    StatusIndicator(GlobalStatus, "ctrl_connecting")

    adb_path_input = (
        ui.input(
            "ADB Path",
            placeholder="eg: C:/adb.exe",
        )
        .props("size=60")
        .bind_value(app.storage.general, "adb_path")
    )
    adb_address_input = (
        ui.input(
            "ADB Address",
            placeholder="eg: 127.0.0.1:5555",
        )
        .props("size=30")
        .bind_value(app.storage.general, "adb_address")
    )
    adb_config_input = (
        ui.input(
            "Extras",
            placeholder="eg: {}",
        )
        .props("size=30")
        .bind_value(app.storage.general, "adb_config")
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_connect(),
    )

    ui.button(
        icon="wifi_find",
        on_click=lambda: on_click_detect(),
    )

    devices_select = ui.select(
        {}, on_change=lambda e: on_change_devices_select(e)
    ).bind_visibility_from(
        GlobalStatus,
        "ctrl_detecting",
        backward=lambda s: s == Status.SUCCEEDED,
    )

    StatusIndicator(GlobalStatus, "ctrl_detecting").label().bind_visibility_from(
        GlobalStatus,
        "ctrl_detecting",
        backward=lambda s: s == Status.RUNNING or s == Status.FAILED,
    )

    async def on_click_connect():
        GlobalStatus.ctrl_connecting = Status.RUNNING

        if not adb_path_input.value or not adb_address_input.value:
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        try:
            config = json.loads(adb_config_input.value)
        except json.JSONDecodeError as e:
            print("Error parsing extras:", e)
            config = {}

        connected = await maafw.connect_adb(
            Path(adb_path_input.value), adb_address_input.value, config
        )
        if not connected:
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        GlobalStatus.ctrl_detecting = Status.PENDING

        await maafw.screenshotter.refresh(True)

    async def on_click_detect():
        GlobalStatus.ctrl_detecting = Status.RUNNING

        devices = await maafw.detect_adb()
        options = {}
        for d in devices:
            v = (d.adb_path, d.address, str(d.config))
            l = d.name + " " + d.address
            options[v] = l

        devices_select.options = options
        devices_select.update()
        if not options:
            GlobalStatus.ctrl_detecting = Status.FAILED
            return

        devices_select.value = next(iter(options))
        on_change_devices_select(devices_select.value)
        GlobalStatus.ctrl_detecting = Status.SUCCEEDED

    def on_change_devices_select(e):
        adb_path_input.value = str(e.value[0])
        adb_address_input.value = e.value[1]
        adb_config_input.value = str(e.value[2])


async def connect_win32_control():
    StatusIndicator(GlobalStatus, "ctrl_connecting")

    hwnd_input = (
        ui.input("HWND").props("size=30").bind_value(app.storage.general, "hwnd")
    )

    SCREENCAP_DICT = {
        MaaWin32ScreencapMethodEnum.GDI: "Screencap_GDI",
        MaaWin32ScreencapMethodEnum.DXGI_DesktopDup: "Screencap_DXGI_DesktopDup",
        MaaWin32ScreencapMethodEnum.FramePool: "Screencap_DXGI_FramePool",
    }
    screencap_select = ui.select(
        SCREENCAP_DICT, value=MaaWin32ScreencapMethodEnum.DXGI_DesktopDup
    ).bind_value(app.storage.general, "win32_screencap")

    INPUT_DICT = {
        MaaWin32InputMethodEnum.SendMessage: "Input_SendMessage",
        MaaWin32InputMethodEnum.Seize: "Input_Seize",
    }
    input_select = ui.select(
        INPUT_DICT,
        value=MaaWin32InputMethodEnum.Seize,
    ).bind_value(app.storage.general, "win32_input")

    ui.button(
        "Connect",
        on_click=lambda: on_click_connect(),
    )
    window_name_input = (
        ui.input("Search Window Name", placeholder="Supports regex, eg: File Explorer")
        .props("size=30")
        .bind_value(app.storage.general, "window_name")
    )
    ui.button(
        icon="wifi_find",
        on_click=lambda: on_click_detect(),
    )

    devices_select = ui.select(
        {}, on_change=lambda e: on_change_devices_select(e)
    ).bind_visibility_from(
        GlobalStatus,
        "ctrl_detecting",
        backward=lambda s: s == Status.SUCCEEDED,
    )

    StatusIndicator(GlobalStatus, "ctrl_detecting").label().bind_visibility_from(
        GlobalStatus,
        "ctrl_detecting",
        backward=lambda s: s == Status.RUNNING or s == Status.FAILED,
    )

    async def on_click_connect():
        GlobalStatus.ctrl_connecting = Status.RUNNING

        if not hwnd_input.value:
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        connected = await maafw.connect_win32hwnd(
            hwnd_input.value, screencap_select.value, input_select.value
        )
        if not connected:
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        GlobalStatus.ctrl_detecting = Status.PENDING

        await maafw.screenshotter.refresh(True)

    async def on_click_detect():
        GlobalStatus.ctrl_detecting = Status.RUNNING

        windows = await maafw.detect_win32hwnd(window_name_input.value)
        options = {}
        for w in windows:
            options[hex(w.hwnd)] = hex(w.hwnd) + " " + w.window_name

        devices_select.options = options
        devices_select.update()
        if not options:
            GlobalStatus.ctrl_detecting = Status.FAILED
            return

        devices_select.value = next(iter(options))
        GlobalStatus.ctrl_detecting = Status.SUCCEEDED

    def on_change_devices_select(e):
        hwnd_input.value = e.value


async def screenshot_control():
    with ui.row().style("align-items: flex-end;"):
        with ui.card().tight():
            ui.interactive_image(
                cross="green",
                on_mouse=lambda e: on_click_image(int(e.image_x), int(e.image_y)),
            ).bind_source_from(maafw.screenshotter, "source").style(
                "height: 200px;"
            ).bind_visibility_from(
                GlobalStatus,
                "ctrl_connecting",
                backward=lambda s: s == Status.SUCCEEDED,
            )

        ui.button(
            icon="refresh", on_click=lambda: on_click_refresh()
        ).bind_visibility_from(
            GlobalStatus, "ctrl_connecting", backward=lambda s: s == Status.SUCCEEDED
        ).bind_enabled_from(
            GlobalStatus, "task_running", backward=lambda s: s != Status.RUNNING
        )

    async def on_click_image(x, y):
        print(f"on_click_image: {x}, {y}")
        await maafw.click(x, y)
        await asyncio.sleep(1)
        await on_click_refresh()

    async def on_click_refresh():
        await maafw.screenshotter.refresh(True)


async def load_resource_control():
    StatusIndicator(GlobalStatus, "res_loading")

    dir_input = (
        ui.input(
            "Resource Directory",
            placeholder="eg: C:/M9A/assets/resource/base",
        )
        .props("size=60")
        .bind_value(app.storage.general, "resource_dir")
    )

    ui.button(
        "Load",
        on_click=lambda: on_click_load(),
    )

    async def on_click_load():
        GlobalStatus.res_loading = Status.RUNNING

        if not dir_input.value:
            GlobalStatus.res_loading = Status.FAILED
            return

        loaded = await maafw.load_resource(Path(dir_input.value))
        if not loaded:
            GlobalStatus.res_loading = Status.FAILED
            return

        GlobalStatus.res_loading = Status.SUCCEEDED


async def run_task_control():
    StatusIndicator(GlobalStatus, "task_running")

    entry_input = (
        ui.input(
            "Task Entry",
            placeholder="eg: StartUp",
        )
        .props("size=30")
        .bind_value(app.storage.general, "task_entry")
    )

    ui.button("Start", on_click=lambda: on_click_start())
    ui.button("Stop", on_click=lambda: on_click_stop())

    async def on_click_start():
        GlobalStatus.task_running = Status.RUNNING

        if not entry_input.value:
            GlobalStatus.task_running = Status.FAILED
            return

        run = await maafw.run_task(entry_input.value)
        if not run:
            GlobalStatus.task_running = Status.FAILED
            return

        GlobalStatus.task_running = Status.SUCCEEDED

    async def on_click_stop():
        stopped = await maafw.stop_task()
        if not stopped:
            GlobalStatus.task_running = Status.FAILED
            return

        GlobalStatus.task_running = Status.PENDING
        await maafw.screenshotter.refresh(True)
