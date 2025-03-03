from nicegui import ui, app

from ..arg_parser import ArgParser


def quit() -> None:
    print("CI is enabled, shutting down...")
    app.shutdown()


ci = ArgParser().get_CI()
if ci:
    print("WARNING: CI mode is enabled.")
    ui.timer(interval=5, callback=quit, once=True)
