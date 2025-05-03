from pathlib import Path
from typing import Optional

from nicegui import ui
from nicegui.native.native_mode import find_open_port

from .webpage import index_page
from .webpage import reco_page


def run(
    *,
    host: str = "localhost",
    port: int = 8011,
    show: bool = True,
    dark: Optional[bool] = None,
):

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
