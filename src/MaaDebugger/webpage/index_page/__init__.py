from nicegui import ui

from .master_control import main as master_control
from .runtime_control import main as runtime_control


@ui.page("/")
async def index():

    await master_control()

    ui.separator()

    await runtime_control()
