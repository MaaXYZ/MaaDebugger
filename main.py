from nicegui import ui

from source.webpage import index_page
from source.webpage import reco_page

ui.dark_mode()  # auto dark mode
ui.run(title="Maa Debugger", storage_secret="maadbg")  # , root_path="/proxy/8080")
