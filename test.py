from nicegui import ui, app

from src.MaaDebugger.main import run

ui.timer(3, lambda: app.shutdown(), once=True)

run()
