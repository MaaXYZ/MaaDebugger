from nicegui import ui, app

from src.MaaDebugger import MaaDebugger

ui.timer(3, lambda: app.shutdown(), once=True)

MaaDebugger.run_in_cli()
