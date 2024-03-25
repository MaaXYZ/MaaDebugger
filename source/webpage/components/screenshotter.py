import asyncio
import threading

import source.interaction.maafw as maafw


class Screenshotter(threading.Thread):
    def __init__(self):
        super().__init__()
        self.source = None
        self.active = False

    def __del__(self):
        self.active = False
        self.source = None

    def run(self):
        while self.active:
            im = asyncio.run(maafw.screencap())
            if not im:
                continue

            self.source = im

    def start(self):
        self.active = True
        super().start()

    def stop(self):
        self.active = False
