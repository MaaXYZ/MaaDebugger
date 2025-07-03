from nicegui import ui, app

from src.MaaDebugger import MaaDebugger

# Waiting for update checking
ui.timer(10, lambda: app.shutdown(), once=True)

MaaDebugger.run()
