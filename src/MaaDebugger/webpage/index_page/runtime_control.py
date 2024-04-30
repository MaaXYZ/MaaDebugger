from collections import defaultdict
from dataclasses import dataclass

from nicegui import ui

from ...maafw import maafw
from ...webpage.components.status_indicator import Status, StatusIndicator
from ...webpage.reco_page import RecoData


async def main():
    Controls.recognition_row.register()


class RecognitionRow:

    def __init__(self) -> None:
        self.row_len = 0
        self.data = defaultdict(dict)

    def register(self):
        self.row = ui.row()

        maafw.on_list_to_recognize = self.on_list_to_recognize
        maafw.on_miss_all = self.on_miss_all
        maafw.on_recognition_result = self.on_recognition_result

    def on_list_to_recognize(self, pre_hit, list_to_reco):
        self.row_len = self.row_len + 1

        self.cur_list = list_to_reco
        self.next_reco_index = 0

        with self.row:
            self._add_list(pre_hit, list_to_reco)

    def _add_list(self, pre_hit, list_to_reco):
        with ui.list().props("bordered separator"):
            ui.item_label(pre_hit).props("header").classes("text-bold")
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

    def on_recognition_result(self, reco_id: int, name: str, hit: bool):
        target = None
        for item in self.data[self.row_len].values():
            if item.status == Status.PENDING and item.name == name:
                target = item
                break

        if not target:
            return

        target.reco_id = reco_id
        target.status = hit and Status.SUCCESS or Status.FAILURE

        RecoData.data[reco_id] = name, hit

    def on_miss_all(self, pre_hit, list_to_reco):
        pass


class Controls:
    recognition_row = RecognitionRow()
