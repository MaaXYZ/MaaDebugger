from nicegui import ui
import asyncio
import threading

from source.interaction.interaction import screencap


class Screenshotter(threading.Thread):
    def __init__(self):
        super().__init__()
        self.source = None
        self.active = False

    def __del__(self):
        self.active = False
        self.source = None
        super().__del__()

    def run(self):
        while self.active:
            im = asyncio.run(screencap())
            if not im:
                continue

            self.source = im

    def start(self):
        self.active = True
        super().start()

    def stop(self):
        self.active = False
