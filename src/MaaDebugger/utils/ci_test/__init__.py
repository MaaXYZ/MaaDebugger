from nicegui import ui, app

from ..arg_parser import ArgParser


def quit() -> None:
    print("CI test passed")
    app.shutdown()


ci = ArgParser().get_CI()
if ci:
    print("WARNING: CI mode is enabled.")
    ui.timer(interval=2, callback=quit, once=True)
