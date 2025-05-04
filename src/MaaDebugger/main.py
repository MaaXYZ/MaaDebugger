from pathlib import Path

from nicegui import ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page

from .utils import args


def main():
    host = args.get_host()
    port = args.get_port() or 8011
    show = args.get_hide()
    dark = args.get_dark()

    index_page.index()

    ui.run(
        title="Maa Debugger",
        storage_secret="maadbg",
        reload=False,
        host=host,
        port=find_open_port(port),
        show=show,
        dark=dark,
        favicon=Path(__file__).parent / "maa.ico",
        # root_path="/proxy/8011",
    )
