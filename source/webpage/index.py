from nicegui import app, ui
from pathlib import Path

from source.control.runner import run_task, stop_task


@ui.page("/")
def index():
    maafw_install_dir_input = (
        ui.input(
            "MaaFramework Release Directory",
            placeholder="eg: C:/Downloads/MAA-win-x86_64",
        )
        .props("size=60")
        .bind_value(app.storage.general, "maafw_install_dir")
    )

    adb_path_input = (
        ui.input("ADB Path", placeholder="eg: C:/adb.exe")
        .props("size=60")
        .bind_value(app.storage.general, "adb_path")
    )

    adb_address_input = (
        ui.input("ADB Address", placeholder="eg: 127.0.0.1:5555")
        .props("size=60")
        .bind_value(app.storage.general, "adb_address")
    )

    resource_dir_input = (
        ui.input("Resource Directory", placeholder="eg: C:/M9A/assets/resource")
        .props("size=60")
        .bind_value(app.storage.general, "resource_dir")
    )

    with ui.row():
        task_input = (
            ui.input("Task", placeholder="Enter the task")
            .props("size=30")
            .bind_value(app.storage.general, "task")
        )

        run_button = ui.button("Run", on_click=on_run_button_click)
        # stop 有 bug，回头修修
        # stop_button = ui.button("Stop", on_click=on_stop_button_click)


async def on_run_button_click():
    maafw_install_dir = Path(app.storage.general.get("maafw_install_dir"))
    adb_path = Path(app.storage.general.get("adb_path"))
    adb_address = app.storage.general.get("adb_address")
    resource_dir = Path(app.storage.general.get("resource_dir"))
    task = app.storage.general.get("task")

    print(
        f"on_run_button_click: maafw_install_dir: {maafw_install_dir}, adb_path: {adb_path}, adb_address: {adb_address}, resource_dir: {resource_dir}, task: {task}"
    )

    message = await run_task(
        maafw_install_dir, adb_path, adb_address, resource_dir, task
    )
    ui.notify(message)


async def on_stop_button_click():
    await stop_task()
    ui.notify("Task stopped")
