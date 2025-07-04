from nicegui import app, ui

from . import master_control, runtime_control
from ...webpage.traceback_page import on_exception


def main():
    master_control.main()
    ui.separator()
    runtime_control.main()

    app.on_exception(on_exception)
