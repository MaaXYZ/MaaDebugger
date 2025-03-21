from nicegui import ui

from .master_control import main as master_control
from .runtime_control import main as runtime_control


def index():

    master_control()

    ui.separator()

    runtime_control()
