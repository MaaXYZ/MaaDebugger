from collections import defaultdict
from dataclasses import dataclass
from typing import Callable
import asyncio

from nicegui import ui
from maa.notification_handler import NotificationHandler, NotificationType

from ...maafw import maafw
from ...webpage.components.status_indicator import Status, StatusIndicator
from ...webpage.reco_page import RecoData


def main():
    Controls.recognition_row.register()


class RecognitionRow:

    class MyNotificationHandler(NotificationHandler):

        def __init__(self) -> None:
            super().__init__()

            self.on_next_list_starting: Callable = None
            self.on_recognized: Callable = None

        def on_node_next_list(
            self,
            noti_type: NotificationType,
            detail: NotificationHandler.NodeNextListDetail,
        ):
            if noti_type != NotificationType.Starting:
                return

            self.on_next_list_starting(detail.name, detail.next_list)

        def on_node_recognition(
            self,
            noti_type: NotificationType,
            detail: NotificationHandler.NodeRecognitionDetail,
        ):
            if (
                noti_type != NotificationType.Succeeded
                and noti_type != NotificationType.Failed
            ):
                return

            self.on_recognized(
                detail.reco_id, detail.name, noti_type == NotificationType.Succeeded
            )

    def __init__(self) -> None:
        self.row_len = 0
        self.data = defaultdict(dict)

    def register(self):
        self.row = ui.row()

        self.notification_handler = self.MyNotificationHandler()
        self.notification_handler.on_next_list_starting = self.on_next_list_starting
        self.notification_handler.on_recognized = self.on_recognized

        maafw.notification_handler = self.notification_handler

    def on_next_list_starting(self, current: str, list_to_reco: list[str]):
        self.row_len = self.row_len + 1

        self.cur_list = list_to_reco
        self.next_reco_index = 0

        with self.row:
            self._add_list(current, list_to_reco)

        asyncio.run(maafw.screenshotter.refresh(False))

    def _add_list(self, current: str, list_to_reco: list[str]):
        with ui.list().props("bordered separator"):
            ui.item_label(current).props("header").classes("text-bold")
            ui.separator()

            for index in range(len(list_to_reco)):
                name = list_to_reco[index]
                self._add_item(index, name)

    @dataclass
    class ItemData:
        col: int
        row: int
        name: str
        reco_id: int = 0
        status: Status = Status.PENDING

    def _add_item(self, index, name):
        data = RecognitionRow.ItemData(self.row_len, index, name)
        self.data[self.row_len][index] = data

        with ui.item(on_click=lambda data=data: self.on_click_item(data)):
            with ui.item_section().props("side"):
                StatusIndicator(data, "status")

            with ui.item_section():
                ui.item_label(name)

            with ui.item_section().props("side"):
                ui.item_label().bind_text_from(data, "reco_id").bind_visibility_from(
                    data, "reco_id", backward=lambda i: i != 0
                ).props("caption")

    def on_click_item(self, data: ItemData):
        print(f"on_click_item ({data.col}, {data.row}): {data.name} ({data.reco_id})")

        ui.navigate.to(f"reco/{data.reco_id}", new_tab=True)

    def on_recognized(self, reco_id: int, name: str, hit: bool):
        print(f"on_recognized: {reco_id}, {name}, {hit}")

        target = None
        for item in self.data[self.row_len].values():
            if item.status == Status.PENDING and item.name == name:
                target = item
                break

        if not target:
            return

        target.reco_id = reco_id
        target.status = hit and Status.SUCCEEDED or Status.FAILED

        RecoData.data[reco_id] = name, hit
        asyncio.run(maafw.screenshotter.refresh(False))


class Controls:
    recognition_row = RecognitionRow()
