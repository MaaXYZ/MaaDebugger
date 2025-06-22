from nicegui import app, ui

from . import master_control, runtime_control
from ...utils import update_checker
from ...webpage.traceback_page import on_exception


def index():
    master_control.main()
    ui.separator()
    runtime_control.main()

    app.on_exception(on_exception)
    ui.timer(0.5, update_checker.main, once=True)  # Check update
