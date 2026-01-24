import asyncio
import io
import json
from pathlib import Path
from typing import Any, Optional, List, Union, Literal

from maa.define import (
    MaaWin32ScreencapMethodEnum,
    MaaWin32InputMethodEnum,
    MaaGamepadTypeEnum,
)
from nicegui import app, binding, ui
from nicegui.elements.mixins.value_element import ValueElement
from PIL.Image import Image

from ...maafw import maafw
from ...utils import input_checker as ic
from ...utils import system, js
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

    screenshot_control()


def connect_control():
    with ui.tabs() as tabs:
        adb = ui.tab("Adb")
        win32 = ui.tab("Win32")
        gamepad = ui.tab("Gamepad")
        playcover = ui.tab("PlayCover")
        custom = ui.tab("Custom")

    tab_panels = (
        ui.tab_panels(tabs, value="Adb")
        .bind_value(STORAGE, "controller_type")
        .props("no-caps")
    )
    with tab_panels:
        with ui.tab_panel(adb):
            with ui.row(align_items="center").classes("w-full"):
                adb_control()
        with ui.tab_panel(win32):
            with ui.row(align_items="center").classes("w-full"):
                win32_control("win32")
        with ui.tab_panel(playcover):
            with ui.row(align_items="center").classes("w-full"):
                playcover_control()
        with ui.tab_panel(gamepad):
            with ui.row(align_items="center").classes("w-full"):
                win32_control("gamepad")
        with ui.tab_panel(custom):
            with ui.row(align_items="center").classes("w-full"):
                custom_control()

    os_type = system.get_os_type()
    if os_type != system.OSTypeEnum.Windows:
        win32.set_visibility(False)
        gamepad.set_visibility(False)
        tab_panels.set_value("Adb")
    if os_type != system.OSTypeEnum.macOS:
        playcover.set_visibility(False)
        tab_panels.set_value("Adb")


def adb_control():
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
                validation=ic.ascii_str,
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


def win32_control(type: Literal["win32", "gamepad"] = "win32"):
    SCREENCAP_DICT = {
        MaaWin32ScreencapMethodEnum.GDI: "GDI",
        MaaWin32ScreencapMethodEnum.FramePool: "FramePool",
        MaaWin32ScreencapMethodEnum.DXGI_DesktopDup: "DXGI_DesktopDup",
        MaaWin32ScreencapMethodEnum.DXGI_DesktopDup_Window: "DXGI_DesktopDup_Window",
        MaaWin32ScreencapMethodEnum.PrintWindow: "PrintWindow",
        MaaWin32ScreencapMethodEnum.ScreenDC: "ScreenDC",
    }

    if type == "win32":
        STORAGE_TARGET_PREFIX = "win32_"
        INPUT_DICT = {
            MaaWin32InputMethodEnum.Seize: "Seize",
            MaaWin32InputMethodEnum.SendMessage: "SendMessage",
            MaaWin32InputMethodEnum.PostMessage: "PostMessage",
            MaaWin32InputMethodEnum.LegacyEvent: "LegacyEvent",
            MaaWin32InputMethodEnum.PostThreadMessage: "PostThreadMessage",
            MaaWin32InputMethodEnum.SendMessageWithCursorPos: "SendMessageWithCursorPos",
            MaaWin32InputMethodEnum.PostMessageWithCursorPos: "PostMessageWithCursorPos",
        }
    else:  # elif type == "gamepad":
        STORAGE_TARGET_PREFIX = "gamepad_"
        GAMEPAD_TYPE_DICT = {
            MaaGamepadTypeEnum.Xbox360: "Xbox360",
            MaaGamepadTypeEnum.DualShock4: "DualShock4",
        }

    with ui.row(align_items="baseline"):
        StatusIndicator(GlobalStatus, "ctrl_connecting")
        hwnd_input = (
            ui.input("HWND", placeholder="0x11451", validation=ic.hwnd)
            .props("size=15")
            .bind_value(STORAGE, "hwnd")
            .on("keydown.enter", lambda: on_click_connect())
        )
        screencap_select = (
            ui.select(
                SCREENCAP_DICT,
                label="Screencap Method",
                value=MaaWin32ScreencapMethodEnum.DXGI_DesktopDup,
            )
            .style("min-width: 100px")
            .bind_value(STORAGE, STORAGE_TARGET_PREFIX + "screencap")
            .on_value_change(lambda: on_connect_params_change())
        )

        if type == "win32":
            mouse_select = (
                ui.select(
                    INPUT_DICT,
                    label="Mouse Input Method",
                    value=MaaWin32InputMethodEnum.Seize,
                )
                .style("min-width: 100px")
                .bind_value(STORAGE, "win32_mouse")
                .on_value_change(lambda: on_connect_params_change())
            )
            keyboard_select = (
                ui.select(
                    INPUT_DICT,
                    label="Keyboard Input Method",
                    value=MaaWin32InputMethodEnum.Seize,
                )
                .style("min-width: 100px")
                .bind_value(STORAGE, "win32_keyboard")
                .on_value_change(lambda: on_connect_params_change())
            )
        else:  # elif type == "gamepad":
            gamepad_type_select = (
                ui.select(GAMEPAD_TYPE_DICT, label="Gamepad Type")
                .style("min-width: 100px")
                .bind_value(STORAGE, "gamepad_keyboard")
                .on_value_change(lambda: on_connect_params_change())
            )

        ui.button(
            "Connect",
            on_click=lambda: on_click_connect(),
        )
        window_name_input = (
            ui.input("Window Name", placeholder="Supports regex")
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

    async def on_connect_params_change():
        if GlobalStatus.ctrl_connecting == Status.SUCCEEDED:
            await on_click_connect()

    async def on_click_connect():
        GlobalStatus.ctrl_connecting = Status.RUNNING

        if not hwnd_input.value:
            GlobalStatus.ctrl_connecting = Status.FAILED
            return

        if type == "win32":
            connected, error = await maafw.connect_win32(
                str(hwnd_input.value),
                int(screencap_select.value),  # type:ignore
                int(mouse_select.value),  # type:ignore
                int(keyboard_select.value),  # type:ignore
            )

        elif type == "gamepad":
            connected, error = await maafw.connect_gamepad(
                str(hwnd_input.value),  # type:ignore
                int(gamepad_type_select.value),  # type:ignore
                int(screencap_select.value),  # type:ignore
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
            options[hex(w.hwnd)] = hex(w.hwnd) + " " + w.window_name  # type:ignore

        hwnd_select.set_options(options)
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


def playcover_control():
    StatusIndicator(GlobalStatus, "ctrl_connecting")

    with ui.row(align_items="baseline"):
        address_input = (
            ui.input(label="Address")
            .tooltip("PlayTools service endpoint (host:port)")
            .props("size=30")
            .bind_value(STORAGE, "playcover_address")
        )
        uuid_input = (
            ui.input(label="UUID")
            .tooltip("Target app bundle identifier")
            .props("size=30")
            .bind_value(STORAGE, "playcover_uuid")
        )
        ui.button(
            "Connect",
            on_click=lambda: on_click_connect(),
        )

    async def on_click_connect():
        GlobalStatus.ctrl_connecting = Status.RUNNING
        connected, err = await maafw.connect_playcover(
            address_input.value, uuid_input.value
        )
        if not connected:
            GlobalStatus.ctrl_connecting = Status.FAILED
            ui.notify(
                f"Failed to connect PlayCover controller. {err}",
                position="bottom-right",
                type="negative",
            )
            return

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        GlobalStatus.ctrl_detecting = Status.PENDING

        await maafw.screenshotter.refresh(True)


def custom_control():
    StatusIndicator(GlobalStatus, "ctrl_connecting")
    with ui.row(align_items="baseline"):
        img_path_input = (
            ui.input(
                label="Image Path",
                validation=ic.is_file,
            )
            .props("size=80")
            .bind_value(STORAGE, "custom_controller_img_path")
        )
        ui.button("Load").on_click(lambda: on_load_img())

    async def on_load_img():
        if not img_path_input.value:
            GlobalStatus.ctrl_connecting = Status.FAILED
            ui.notify(
                "Image path cannot be empty.",
                position="bottom-right",
                type="negative",
            )
            return

        _path = Path(img_path_input.value)
        if not _path.is_file():
            GlobalStatus.ctrl_connecting = Status.FAILED
            ui.notify(
                "Please enter a valid image file path.",
                position="bottom-right",
                type="negative",
            )
            return

        GlobalStatus.ctrl_connecting = Status.RUNNING
        try:
            maafw.connect_custom_controller(_path)
        except Exception as e:
            GlobalStatus.ctrl_connecting = Status.FAILED
            raise e

        GlobalStatus.ctrl_connecting = Status.SUCCEEDED
        await maafw.screenshotter.refresh(True)


def screenshot_control():
    with (
        ui.row()
        .style("align-items: flex-end;")
        .bind_visibility_from(
            maafw.screenshotter, "source", backward=lambda x: x is not None
        )
    ):
        with ui.card().tight():
            img = (
                ui.interactive_image(
                    cross="green",
                    on_mouse=lambda e: on_click_image(int(e.image_x), int(e.image_y)),
                )
                .bind_source_from(maafw.screenshotter, "source")
                .style("height: 200px;")
            )

        ui.button(
            icon="refresh", on_click=lambda: on_click_refresh()
        ).bind_enabled_from(
            GlobalStatus, "task_running", backward=lambda s: s != Status.RUNNING
        )
        ui.button(
            icon="download",
            on_click=lambda: on_download_image(img.source),  # type:ignore
        ).bind_enabled_from(img, "source", lambda x: x is not None)

    async def on_click_image(x, y):
        if await maafw.click(x, y):
            print(f"on_click_image: {x}, {y}")
            await asyncio.sleep(1)
            await on_click_refresh()
        else:
            print(f"Failed to click at {x}, {y}")

    async def on_click_refresh():
        await maafw.screenshotter.refresh(True)

    def on_download_image(img: Union[Image, Any]):
        if not img or type(img) != Image:
            return

        # Image to Bytes
        bytes_io = io.BytesIO()
        img.save(bytes_io, format="PNG")
        img_bytes = bytes_io.getvalue()

        # Use hash of Bytes as filename
        ui.download(img_bytes, f"{hash(img_bytes)}.png")


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
        .props("size=40")
        .bind_value(STORAGE, "agent_identifier")
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_agent(),
    ).bind_enabled_from(GlobalStatus, "agent_connecting", lambda x: x != Status.RUNNING)

    async def on_click_agent():
        GlobalStatus.agent_connecting = Status.RUNNING

        created, error = await maafw.create_agent(agent_identifier_input.value)
        if not created:
            GlobalStatus.agent_connecting = Status.FAILED
            ui.notify(error, position="bottom-right", type="negative")
            print(error)
            return

        agent_identifier_input.value = maafw.agent_identifier
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
    else:
        GlobalStatus.res_loading = Status.SUCCEEDED
        node_list = sorted(await maafw.get_node_list())
        NodeListElement.value = node_list


def run_task_control():
    StatusIndicator(GlobalStatus, "task_running")

    with ui.row(align_items="baseline"):
        entry_select = (
            ui.select([], label="Task Entry", with_input=True)
            .props("size=30")
            .bind_value(STORAGE, "task_entry")
        )
        ui.timer(
            0.1,
            lambda: ui.run_javascript(js.select_filter(entry_select.id)),
            once=True,
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

        ui.button("Start", on_click=lambda: on_click_start()).bind_enabled_from(
            GlobalStatus, "task_running", backward=lambda s: s != Status.RUNNING
        )
        ui.button("Stop", on_click=lambda: on_click_stop())

        NodeListElement.on_value_change(
            lambda: entry_select.set_options(
                NodeListElement.value, value=get_entry_node()
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


def get_entry_node() -> Optional[str]:
    """
    Get a entry node value. (Current Node -> List[0] -> None)
    """
    node_list: List[str] = NodeListElement.value

    if not node_list:
        return None

    entry = STORAGE.get("task_entry", None)
    if entry is None or entry not in node_list:
        return node_list[0] or None
    else:
        return entry
