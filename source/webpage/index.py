from nicegui import app, ui


@ui.page("/")
def index():
    adb_path = (
        ui.input("ADB Path", placeholder="eg: C:/adb.exe")
        .props("size=60")
        .bind_value(app.storage.user, "adb_path")
    )

    adb_address = (
        ui.input("ADB Address", placeholder="eg: 127.0.0.1:5555")
        .props("size=60")
        .bind_value(app.storage.user, "adb_address")
    )

    resource_path = (
        ui.input("Resource Path", placeholder="eg: C:/my_proj/resource")
        .props("size=60")
        .bind_value(app.storage.user, "resource_path")
    )

    with ui.row():
        task_input = (
            ui.input("Task", placeholder="Enter the task")
            .props("size=30")
            .bind_value(app.storage.user, "task")
        )

        run_button = ui.button("Run")
