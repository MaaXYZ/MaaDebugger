from nicegui import ui

from .webpage import index_page
from .webpage import reco_page

from .utils import args


def main():
    port = args.get_port()

    ui.dark_mode()  # auto dark mode
    ui.run(
        port=port, title="Maa Debugger", storage_secret="maadbg", reload=False
    )  # , root_path="/proxy/8080")
