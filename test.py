from nicegui import ui, app

from src.MaaDebugger.main import main

ui.timer(3, lambda: app.shutdown(), once=True)

main()
