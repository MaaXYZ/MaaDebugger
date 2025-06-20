import asyncio
import json
from pathlib import Path
from typing import Optional, List

from maa.define import MaaWin32ScreencapMethodEnum, MaaWin32InputMethodEnum
from nicegui import app, binding, ui
from nicegui.elements.mixins.value_element import ValueElement

from ...maafw import maafw
from ...utils import input_checker as ic
from ...utils import update_checker
from ...webpage.components.status_indicator import Status, StatusIndicator
from .global_status import GlobalStatus

binding.MAX_PROPAGATION_TIME = 1
STORAGE = app.storage.general


NodeListElement = ValueElement(value=[])


def main():
    with ui.row():
        with ui.column():
            connect_control()
            with ui.row(align_items="center").classes("w-full"):
                load_resource_control()
            with ui.row(align_items="center").classes("w-full"):
                agent_control()
            with ui.row(align_items="center").classes("w-full"):
                run_task_control()

        with ui.column():
            screenshot_control()


def connect_control():
    with ui.tabs() as tabs:
        adb = ui.tab("Adb")
        win32 = ui.tab("Win32")

    with ui.tab_panels(tabs, value="Adb").bind_value(STORAGE, "controller_type"):
        with ui.tab_panel(adb):
            with ui.row(align_items="center").classes("w-full"):
                connect_adb_control()
        with ui.tab_panel(win32):
            with ui.row(align_items="center").classes("w-full"):
                connect_win32_control()


def connect_adb_control():
    with ui.row(align_items="baseline"):
        StatusIndicator(GlobalStatus, "ctrl_connecting")
        adb_path_input = (
            ui.input(
                "ADB Path",
                placeholder="eg: C:/adb.exe",
            )
            .props("size=60")
            .bind_value(STORAGE, "adb_path")
        )
        adb_address_input = (
            ui.input(
                "ADB Address",
                placeholder="eg: 127.0.0.1:5555",
            )
            .props("size=20")
            .bind_value(STORAGE, "adb_address")
        )

        adb_config_input = (
            ui.input(
                "Extras",
                value="{}",
                placeholder="eg: {}",
                validation=ic.json_style_str,
            )
            .props("size=20")
            .bind_value(STORAGE, "adb_config")
        )
        ui.button(
            "Connect",
            on_click=lambda: on_click_connect(),
        )
        ui.button(
            "Scan",
            on_click=lambda: on_click_scan(),
        )

        device_select = ui.select(
            {},
            label="Devices",
            on_change=lambda e: on_change_device_select(e.value),
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
        if not adb_config_input.value:
            adb_config_input.value = "{}"
        try:
            config = json.loads(adb_config_input.value)
        except json.JSONDecodeError as e:
            ui.notify(
                f"Error parsing extras: {e}", position="bottom-right", type="negative"
            )
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        connected, error = await maafw.connect_adb(
            Path(adb_path_input.value), adb_address_input.value, config
        )
        if not connected:
            GlobalStatus.ctrl_connecting = Status.FAILED
            ui.notify(error, position="bottom-right", type="negative")
            print(error)
            return

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        GlobalStatus.ctrl_detecting = Status.PENDING

        await maafw.screenshotter.refresh(True)

    async def on_click_scan():
        GlobalStatus.ctrl_detecting = Status.RUNNING

        devices = await maafw.detect_adb()
        options = {}
        for d in devices:
            v = (d.adb_path, d.address, json.dumps(d.config))
            l = d.name + " " + d.address
            options[v] = l

        device_select.set_options(options)
        device_select.update()

        if not options:
            GlobalStatus.ctrl_detecting = Status.FAILED
            return

        device_select.set_value(next(iter(options)))
        on_change_device_select(device_select.value)
        GlobalStatus.ctrl_detecting = Status.SUCCEEDED

    def on_change_device_select(value: Optional[List[str]]):
        if not value:
            return

        adb_path_input.value = str(value[0])
        adb_address_input.value = value[1]
        adb_config_input.value = value[2]


def connect_win32_control():
    SCREENCAP_DICT = {
        MaaWin32ScreencapMethodEnum.GDI: "Screencap_GDI",
        MaaWin32ScreencapMethodEnum.DXGI_DesktopDup: "Screencap_DXGI_DesktopDup",
        MaaWin32ScreencapMethodEnum.FramePool: "Screencap_DXGI_FramePool",
    }

    INPUT_DICT = {
        MaaWin32InputMethodEnum.SendMessage: "Input_SendMessage",
        MaaWin32InputMethodEnum.Seize: "Input_Seize",
    }

    with ui.row(align_items="baseline"):
        StatusIndicator(GlobalStatus, "ctrl_connecting")
        hwnd_input = (
            ui.input("HWND", placeholder="0x11451", validation=ic.hwnd)
            .props("size=30")
            .bind_value(STORAGE, "hwnd")
            .on("keydown.enter", lambda: on_click_connect())
        )
        screencap_select = ui.select(
            SCREENCAP_DICT,
            label="Screencap Method",
            value=MaaWin32ScreencapMethodEnum.DXGI_DesktopDup,
        ).bind_value(STORAGE, "win32_screencap")
        input_select = ui.select(
            INPUT_DICT,
            label="Input Method",
            value=MaaWin32InputMethodEnum.Seize,
        ).bind_value(STORAGE, "win32_input")

        ui.button(
            "Connect",
            on_click=lambda: on_click_connect(),
        )
        window_name_input = (
            ui.input(
                "Search Window Name", placeholder="Supports regex, eg: File Explorer"
            )
            .props("size=30")
            .bind_value(STORAGE, "window_name")
            .on("keydown.enter", lambda: on_click_scan())
        )
        ui.button(
            "Scan",
            on_click=lambda: on_click_scan(),
        )

        hwnd_select = ui.select(
            {}, label="Windows", on_change=lambda e: on_change_hwnd_select(e.value)
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

        connected, error = await maafw.connect_win32hwnd(
            hwnd_input.value, screencap_select.value, input_select.value  # type:ignore
        )
        if not connected:
            GlobalStatus.ctrl_connecting = Status.FAILED
            ui.notify(error, position="bottom-right", type="negative")
            return

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        GlobalStatus.ctrl_detecting = Status.PENDING

        await maafw.screenshotter.refresh(True)

    async def on_click_scan():
        GlobalStatus.ctrl_detecting = Status.RUNNING

        windows = await maafw.detect_win32hwnd(window_name_input.value)
        options = {}
        for w in windows:
            options[hex(w.hwnd)] = hex(w.hwnd) + " " + w.window_name

        hwnd_select.set_options(options)
        hwnd_select.update()
        if not options:
            GlobalStatus.ctrl_detecting = Status.FAILED
            return

        hwnd_select.set_value(next(iter(options)))
        on_change_hwnd_select(hwnd_select.value)
        GlobalStatus.ctrl_detecting = Status.SUCCEEDED

    def on_change_hwnd_select(value: Optional[str]):
        if not value:
            return

        hwnd_input.value = value


def screenshot_control():
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


def load_resource_control():
    StatusIndicator(GlobalStatus, "res_loading")

    with ui.row(align_items="baseline").classes("w-3/4"):
        dir_input = (
            ui.textarea(
                "Resource Directory",
                placeholder="Separate with newline, eg: C:/M9A/assets/resource/base",
                validation=ic.paths_exist,
            )
            .props("input-class=h-7")
            .style("width: 500px;")
            .bind_value(STORAGE, "resource_dir")
            .tooltip("Directorise are separated by newline characters.")
        )

        ui.button(
            "Load",
            on_click=lambda: on_click_resource_load(dir_input.value),
        )


def agent_control():
    StatusIndicator(GlobalStatus, "agent_connecting")

    agent_identifier_input = (
        ui.input(
            "Agent Identifier",
        )
        .props("size=20")
        .bind_value(STORAGE, "agent_identifier")
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_agent(),
    )

    async def on_click_agent():
        GlobalStatus.agent_connecting = Status.RUNNING

        identifier = await maafw.create_agent(agent_identifier_input.value)
        agent_identifier_input.value = identifier

        connected, error = await maafw.connect_agent()
        if not connected:
            GlobalStatus.agent_connecting = Status.FAILED
            ui.notify(error, position="bottom-right", type="negative")
            print(error)
            return

        GlobalStatus.agent_connecting = Status.SUCCEEDED


async def on_click_resource_load(values: Optional[str]):
    GlobalStatus.res_loading = Status.RUNNING

    if not values:
        GlobalStatus.res_loading = Status.FAILED
        return

    paths = [Path(p) for p in values.split("\n") if p]

    loaded, error = await maafw.load_resource(paths)
    if not loaded:
        NodeListElement.value = []
        GlobalStatus.res_loading = Status.FAILED
        ui.notify(error, position="bottom-right", type="negative")
        print(error)
        return
    else:
        GlobalStatus.res_loading = Status.SUCCEEDED
        node_list = sorted(await maafw.get_node_list())
        NodeListElement.value = node_list


def run_task_control():
    StatusIndicator(GlobalStatus, "task_running")

    with ui.row(align_items="baseline"):
        entry_select = (
            ui.select(
                [],
                label="Task Entry",
                with_input=True,
            )
            .props("size=30")
            .bind_value(STORAGE, "task_entry")
        )

        pipeline_override_input = (
            ui.input(
                "Pipeline Override",
                value="{}",
                placeholder="eg: {}",
                validation=ic.json_style_str,
            )
            .props("size=60")
            .bind_value(STORAGE, "task_pipeline_override")
        )

        ui.button("Start", on_click=lambda: on_click_start())
        ui.button("Stop", on_click=lambda: on_click_stop())

        task_entry = STORAGE.get("task_entry", None)

        NodeListElement.on_value_change(
            lambda: entry_select.set_options(NodeListElement.value)
        )
        NodeListElement.on_value_change(
            lambda: entry_select.set_value(
                check_entry_node(task_entry, NodeListElement.value)
                or NodeListElement.value[0]
                if NodeListElement.value
                else None
            )
        )

    async def on_click_start():
        GlobalStatus.task_running = Status.RUNNING

        if not entry_select.value:
            GlobalStatus.task_running = Status.FAILED
            return
        if not pipeline_override_input.value:
            pipeline_override_input.value = "{}"
        try:
            pipeline_override = json.loads(pipeline_override_input.value)
        except json.JSONDecodeError as e:
            ui.notify(
                f"Error parsing pipeline override: {e}",
                position="bottom-right",
                type="negative",
            )
            GlobalStatus.task_running = Status.FAILED
            return

        await on_click_resource_load(STORAGE.get("resource_dir"))  # type:ignore

        run, error = await maafw.run_task(entry_select.value, pipeline_override)
        if not run:
            GlobalStatus.task_running = Status.FAILED
            print(error)
            if error is not None:
                ui.notify(error, position="bottom-right", type="negative")
            return

        GlobalStatus.task_running = Status.SUCCEEDED

    async def on_click_stop():
        stopped = await maafw.stop_task()
        if not stopped:
            GlobalStatus.task_running = Status.FAILED
            return

        GlobalStatus.task_running = Status.PENDING
        await maafw.screenshotter.refresh(True)


def check_entry_node(entry: Optional[str], node_list: List[str]) -> Optional[str]:
    """
    Check if the entry node is in the node list.
    """
    if not entry or not node_list or entry not in node_list:
        return None
    else:
        return entry
