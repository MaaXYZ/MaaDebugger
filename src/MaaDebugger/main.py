from nicegui import ui

from .webpage import index_page
from .webpage import reco_page


def main():
    ui.dark_mode()  # auto dark mode
    ui.run(
        title="Maa Debugger", storage_secret="maadbg", reload=False
    )  # , root_path="/proxy/8080")
