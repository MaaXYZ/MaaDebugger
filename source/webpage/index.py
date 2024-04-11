from nicegui import ui

# from source.maafw import maafw

from .master_control import main as master_control
from .runtime_control import main as runtime_control


ui.dark_mode()  # auto dark mode


@ui.page("/")
async def index():

    await master_control()

    ui.separator()

    await runtime_control()
