from nicegui import ui


class PipelineRow:
    def __init__(self) -> None:
        self.row = ui.row()
        self.row_len = 0

    def on_list_to_recognize(self, pre_hit, list_to_reco):
        self.row_len = self.row_len + 1

        self.cur_list = list_to_reco
        self.next_reco_index = 0

        with self.row:
            with ui.list().props("bordered separator"):
                ui.item_label(pre_hit).props("header").classes("text-bold")
                ui.separator()

                for index in range(len(list_to_reco)):
                    name = list_to_reco[index]
                    self._add_item(index, name)

    def _add_item(self, index, name):
        data = (self.row_len, index, name)

        with ui.item(on_click=lambda data=data: self.on_click(data)):
            with ui.item_section():
                ui.item_label(name)

    def on_click(self, data):
        print(f"Clicked on {data}")

    def on_recognition_result(self, result):
        pass
