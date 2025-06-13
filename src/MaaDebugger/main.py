from pathlib import Path

from nicegui import app, ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page
from .utils import args, update_checker
from .webpage.traceback_page import on_exception


def main():
    host = args.get_host()
    port = args.get_port() or 8011
    show = args.get_hide()
    dark = args.get_dark()

    index_page.index()

    app.on_exception(on_exception)
    ui.timer(0.5, update_checker.main, once=True)  # Check update

    ui.run(
        title="Maa Debugger",
        storage_secret="maadbg",
        reload=False,
        host=host,
        port=find_open_port(port, end_port=port + 100),
        show=show,
        dark=dark,
        favicon=Path(__file__).parent / "maa.ico",
        # root_path="/proxy/8011",
    )
