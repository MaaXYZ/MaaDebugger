from nicegui import ui
from dataclasses import dataclass
from collections import defaultdict
from PIL import Image

from source.maafw import maafw
from source.webpage.components.status_indicator import Status, StatusIndicator
from source.utils import cvmat_to_image


async def main():
    Controls.recognition_row.register()
    Controls.recognition_dialog.register()


@dataclass
class RecoItemData:
    col: int
    row: int
    name: str
    reco_id: int = 0
    status: Status = Status.PENDING
    hit_box: tuple[int, int, int, int] | None = None
    detail: dict | None = None
    draws: list[Image.Image] | None = None


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

    def _add_item(self, index, name):
        data = RecoItemData(self.row_len, index, name)
        self.data[self.row_len][index] = data

        with ui.item(on_click=lambda data=data: self.on_click_item(data)):
            with ui.item_section().props("side"):
                StatusIndicator(data, "status")

            with ui.item_section():
                ui.item_label(name)

    def on_click_item(self, data: RecoItemData):
        print(f"on_click_item ({data.col}, {data.row}): {data.name} ({data.reco_id})")

        Controls.recognition_dialog.set_data(data)
        Controls.recognition_dialog.open()

    def on_recognition_result(
        self, reco_id: int, name: str, hit: bool, reco_detail: "RecognitionDetail"
    ):
        target = None
        for item in self.data[self.row_len].values():
            if item.status == Status.PENDING and item.name == name:
                target = item
                break

        if not target:
            return

        target.reco_id = reco_id
        target.status = hit and Status.SUCCESS or Status.FAILURE
        target.hit_box = reco_detail.hit_box
        target.detail = reco_detail.detail
        target.draws = [cvmat_to_image(d) for d in reco_detail.draws]

    def on_miss_all(self, pre_hit, list_to_reco):
        pass


class RecognitionDialog:

    def __init__(self) -> None:
        self.image_source = None
        self.text = None

    def register(self):
        with ui.dialog() as self.dialog, ui.card():
            ui.image().props('fit=scale-down').bind_source_from(self, "image_source")
            ui.label().bind_text_from(self, "text")

    def set_data(self, data: RecoItemData):
        self.data = data

        self.title = f"{self.data.name} ({self.data.reco_id})"

        if self.data.draws:
            # TODO: show all draws
            self.image_source = self.data.draws[0]

        self.text = str(self.data.detail)

    def open(self):
        self.dialog.open()


class Controls:
    recognition_row = RecognitionRow()
    recognition_dialog = RecognitionDialog()
