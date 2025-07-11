from nicegui import app, ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page
from .utils import args, update_checker
from .webpage.traceback_page import on_exception
from .assets import FAVICON_PATH


def main():
    host = args.get_host()
    port = args.get_port() or 8011
    show = args.get_hide()
    dark = args.get_dark()

    index_page.index()

    app.on_exception(on_exception)
    ui.timer(2, update_checker.main, once=True)  # Check update

    ui.run(
        title="Maa Debugger",
        storage_secret="maadbg",
        favicon=FAVICON_PATH,
        reload=False,
        host=host,
        port=find_open_port(port, end_port=port + 100),
        show=show,
        dark=dark,
        # root_path="/proxy/8011",
    )
