from nicegui import app, ui, binding
from pathlib import Path

from source.maafw import MaaFW

from .components.status_indicator import Status, StatusIndicator
from .pipeline_row import PipelineRow


maafw = MaaFW()

binding.MAX_PROPAGATION_TIME = 1
ui.dark_mode()  # auto dark mode


@ui.page("/")
async def index():

    with ui.row().style("align-items: center;"):
        await import_maa_control()

    ui.separator()

    with ui.row():
        with ui.column():
            with ui.row().style("align-items: center;"):
                await connect_adb_control()
            with ui.row().style("align-items: center;"):
                await load_resource_control()
            with ui.row().style("align-items: center;"):
                await run_task_control()

        with ui.column():
            await screenshot_control()

    ui.separator()

    pipeline_control()

    ## for debug
    # append_list_to_reco("AAA", ["BBB", "CCC"])


class GlobalStatus:
    maa_importing: Status = Status.PENDING
    adb_connecting: Status = Status.PENDING
    adb_detecting: Status = Status.PENDING  # not required
    res_loading: Status = Status.PENDING
    task_running: Status = Status.PENDING


async def import_maa_control():

    StatusIndicator(GlobalStatus, "maa_importing")

    pybinding_input = (
        ui.input(
            "MaaFramework Python Binding Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64/binding/Python",
        )
        .props("size=60")
        .bind_value(app.storage.general, "maa_pybinding")
        .bind_enabled_from(
            GlobalStatus, "maa_importing", backward=lambda s: s != Status.SUCCESS
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
            GlobalStatus, "maa_importing", backward=lambda s: s != Status.SUCCESS
        )
    )

    import_button = ui.button(
        "Import", on_click=lambda: on_click_import()
    ).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s != Status.SUCCESS
    )

    async def on_click_import():
        GlobalStatus.maa_importing = Status.RUNNING

        if not pybinding_input.value or not bin_input.value:
            GlobalStatus.maa_importing = Status.FAILURE
            return

        imported = await maafw.import_maa(
            Path(pybinding_input.value), Path(bin_input.value)
        )
        if not imported:
            GlobalStatus.maa_importing = Status.FAILURE
            return

        GlobalStatus.maa_importing = Status.SUCCESS


async def connect_adb_control():
    StatusIndicator(GlobalStatus, "adb_connecting")

    adb_path_input = (
        ui.input(
            "ADB Path",
            placeholder="eg: C:/adb.exe",
        )
        .props("size=60")
        .bind_value(app.storage.general, "adb_path")
        .bind_enabled_from(
            GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
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
            GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
        )
    )
    ui.button(
        "Connect",
        on_click=lambda: on_click_connect(),
    ).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
    )

    ui.button(
        "Detect",
        on_click=lambda: on_click_detect(),
    ).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
    )

    devices_select = (
        ui.select({}, on_change=lambda e: on_change_devices_select(e))
        .bind_enabled_from(
            GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
        )
        .bind_visibility_from(
            GlobalStatus,
            "adb_detecting",
            backward=lambda s: s == Status.SUCCESS,
        )
    )

    StatusIndicator(GlobalStatus, "adb_detecting").label().bind_visibility_from(
        GlobalStatus,
        "adb_detecting",
        backward=lambda s: s == Status.RUNNING or s == Status.FAILURE,
    )

    async def on_click_connect():
        GlobalStatus.adb_connecting = Status.RUNNING

        if not adb_path_input.value or not adb_address_input.value:
            GlobalStatus.adb_connecting = Status.FAILURE
            return

        connected = await maafw.connect_adb(
            Path(adb_path_input.value), adb_address_input.value
        )
        if not connected:
            GlobalStatus.adb_connecting = Status.FAILURE
            return

        GlobalStatus.adb_connecting = Status.SUCCESS
        GlobalStatus.adb_detecting = Status.PENDING

        maafw.screenshotter.start()

    async def on_click_detect():
        GlobalStatus.adb_detecting = Status.RUNNING

        devices = await maafw.detect_adb()
        options = {}
        for d in devices:
            v = (d.adb_path, d.address)
            l = d.name + " " + d.address
            options[v] = l

        devices_select.options = options
        devices_select.update()
        if not options:
            GlobalStatus.adb_detecting = Status.FAILURE
            return

        devices_select.value = next(iter(options))
        GlobalStatus.adb_detecting = Status.SUCCESS

    def on_change_devices_select(e):
        adb_path_input.value = str(e.value[0])
        adb_address_input.value = e.value[1]


async def screenshot_control():
    with ui.card().tight():
        ui.interactive_image(
            cross="green",
            on_mouse=lambda e: on_click_image(int(e.image_x), int(e.image_y)),
        ).bind_source_from(maafw.screenshotter, "source").style(
            "height: 200px;"
        ).bind_visibility_from(
            GlobalStatus, "adb_connecting", backward=lambda s: s == Status.SUCCESS
        )

    async def on_click_image(x, y):
        print(f"on_click_image: {x}, {y}")
        await maafw.click(x, y)


async def load_resource_control():
    StatusIndicator(GlobalStatus, "res_loading")

    dir_input = (
        ui.input(
            "Resource Directory",
            placeholder="eg: C:/M9A/assets/resource/base",
        )
        .props("size=60")
        .bind_value(app.storage.general, "resource_dir")
        .bind_enabled_from(
            GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
        )
    )

    ui.button(
        "Load",
        on_click=lambda: on_click_load(),
    ).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
    )

    async def on_click_load():
        GlobalStatus.res_loading = Status.RUNNING

        if not dir_input.value:
            GlobalStatus.res_loading = Status.FAILURE
            return

        loaded = await maafw.load_resource(Path(dir_input.value))
        if not loaded:
            GlobalStatus.res_loading = Status.FAILURE
            return

        GlobalStatus.res_loading = Status.SUCCESS


async def run_task_control():
    StatusIndicator(GlobalStatus, "task_running")

    entry_input = (
        ui.input(
            "Task Entry",
            placeholder="eg: StartUp",
        )
        .props("size=30")
        .bind_value(app.storage.general, "task_entry")
        .bind_enabled_from(
            GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
        )
    )

    ui.button("Start", on_click=lambda: on_click_start()).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
    )
    ui.button("Stop", on_click=lambda: on_click_stop()).bind_enabled_from(
        GlobalStatus, "maa_importing", backward=lambda s: s == Status.SUCCESS
    )

    async def on_click_start():
        maafw.screenshotter.stop()

        GlobalStatus.task_running = Status.RUNNING

        if not entry_input.value:
            GlobalStatus.task_running = Status.FAILURE
            return

        run = await maafw.run_task(entry_input.value)
        if not run:
            GlobalStatus.task_running = Status.FAILURE
            return

        GlobalStatus.task_running = Status.SUCCESS

    async def on_click_stop():
        stopped = await maafw.stop_task()
        if not stopped:
            GlobalStatus.task_running = Status.FAILURE
            return

        GlobalStatus.task_running = Status.PENDING
        maafw.screenshotter.start()


pipeline_row = None


def pipeline_control():
    global pipeline_row

    pipeline_row = PipelineRow()
    maafw.on_list_to_recognize = pipeline_row.on_list_to_recognize
    maafw.on_recognition_result = pipeline_row.on_recognition_result

    # for debug
    pipeline_row.on_list_to_recognize("AAA", ["BBB", "CCC"])
    pipeline_row.on_list_to_recognize("BBB", ["DDD", "EEE"])