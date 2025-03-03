from nicegui import ui

from .webpage import index_page
from .webpage import reco_page

from .utils import args
from .utils import ci_test


def main():
    host = args.get_host()
    port = args.get_port()
    show = args.get_hide()
    dark = args.get_dark()

    ui.run(
        title="Maa Debugger",
        storage_secret="maadbg",
        reload=False,
        host=host,
        port=port,
        show=show,
        dark=dark,
        # root_path="/proxy/8011",
    )
