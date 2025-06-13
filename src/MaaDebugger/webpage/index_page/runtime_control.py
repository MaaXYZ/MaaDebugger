from collections import defaultdict
from dataclasses import dataclass, field
from typing import Callable, Optional
import asyncio

from nicegui import ui
from nicegui.binding import bindable_dataclass
from maa.notification_handler import NotificationHandler, NotificationType

from ...maafw import maafw
from ...webpage.components.status_indicator import Status, StatusIndicator
from ...webpage.reco_page import RecoData
from .global_status import GlobalStatus

PER_PAGE_ITEMs_NUM = 200


@bindable_dataclass
class ItemData:
    col: int
    row: int
    name: str
    reco_id: int = 0
    status: Status = Status.PENDING


@bindable_dataclass
class ItemPageData:
    reverse: bool = True


@dataclass
class ListData:
    row_len: int
    current: str
    list_to_reco: list[str]


def main():
    reco_data = RecognitionRow()
    reco_data.init_elements()


class MyNotificationHandler(NotificationHandler):
    def __init__(self) -> None:
        super().__init__()

        self.on_next_list_starting: Optional[Callable] = None
        self.on_recognized: Optional[Callable] = None

    def on_node_next_list(
        self,
        noti_type: NotificationType,
        detail: NotificationHandler.NodeNextListDetail,
    ):
        if noti_type != NotificationType.Starting:
            return

        if self.on_next_list_starting is not None:
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

        if self.on_recognized is not None:
            self.on_recognized(
                detail.reco_id, detail.name, noti_type == NotificationType.Succeeded
            )


class RecognitionRow:
    def __init__(self) -> None:
        self.row_len = 0
        self.data = defaultdict(dict)
        self.ItemPageData = ItemPageData()
        self.lsdata_dict: dict[int, ListData] = {}

        self.register_notification_handler()

    def register_notification_handler(self):
        self.notification_handler = MyNotificationHandler()
        self.notification_handler.on_next_list_starting = self.on_next_list_starting
        self.notification_handler.on_recognized = self.on_recognized

        maafw.notification_handler = self.notification_handler

    def init_elements(self):
        with ui.row():
            ui.button("Clear Items", icon="remove", on_click=self.clear_items).props(
                "no-caps"
            )
            ui.button(
                "Clear Items and Cache", icon="delete_forever", on_click=self.clear
            ).props("no-caps").bind_enabled_from(
                GlobalStatus,
                "task_running",
                lambda x: x == Status.FAILED or x == Status.SUCCEEDED,
            )
            ui.switch("Reverse").bind_value(
                self.ItemPageData, "reverse"
            ).bind_enabled_from(
                GlobalStatus,
                "task_running",
                lambda x: x == Status.FAILED or x == Status.SUCCEEDED,
            )

        self.pagination = ui.pagination(1, 1)

        self.homepage_row = ui.row(align_items="start")
        self.other_page_row = ui.row(align_items="start")
        self.other_page_row.set_visibility(False)
        self.pagination.on_value_change(
            lambda: self.on_page_change(self.pagination.value)
        )
        self.pagination.bind_enabled_from(
            GlobalStatus,
            "task_running",
            lambda x: x == Status.FAILED or x == Status.SUCCEEDED,
        )

    async def clear(self):
        await maafw.clear_cache()
        self.clear_items()
        self.data.clear()

    def clear_items(self):
        self.row_len = 0
        self.data.clear()
        self.lsdata_dict.clear()

        self.homepage_row.clear()
        self.other_page_row.clear()
        self.pagination.max = 1
        self.pagination.set_value(1)

    def on_page_change(self, page: int):
        self.other_page_row.clear()

        if page == 1:
            self.homepage_row.set_visibility(True)
            self.other_page_row.set_visibility(False)
        else:
            self.homepage_row.set_visibility(False)
            self.other_page_row.set_visibility(True)
            self.other_page_row.clear()

            # 计算所有 row_len 的倒序列表
            row_len_list = range(self.row_len, 0, -1)
            # 计算当前页的起止索引，确保不会越界
            total_items = len(row_len_list)
            start_index = max((page - 1) * PER_PAGE_ITEMs_NUM, 0)
            end_index = min(start_index + PER_PAGE_ITEMs_NUM, total_items)
            # 切片获取当前页要显示的 row_len
            for row_len in row_len_list[start_index:end_index]:
                self.create_list(self.other_page_row, self.lsdata_dict[row_len])

    def on_recognized(self, reco_id: int, name: str, hit: bool):
        target_item = None
        for item in self.data[self.row_len].values():
            if item.status == Status.PENDING and item.name == name:
                target_item = item
                break

        if not target_item:
            return

        target_item.reco_id = reco_id
        target_item.status = hit and Status.SUCCEEDED or Status.FAILED

        RecoData.data[reco_id] = name, hit
        asyncio.run(maafw.screenshotter.refresh(False))

    def on_next_list_starting(self, current: str, list_to_reco: list[str]):
        self.row_len += 1

        list_data = ListData(self.row_len, current, list_to_reco)
        self.add_list_data(list_data)

        # ui
        # 299/300 -> page:1 | 300/300 -> page:2
        if self.row_len / PER_PAGE_ITEMs_NUM >= self.pagination.max:
            self.pagination.max += 1
            self.homepage_row.clear()

        self.create_list(self.homepage_row, list_data)
        asyncio.run(maafw.screenshotter.refresh(False))

    def add_list_data(self, data: ListData):
        self.lsdata_dict[data.row_len] = data
        for index in range(len(data.list_to_reco)):
            name = data.list_to_reco[index]
            self.add_item_data(index, name, data.row_len)

    def add_item_data(self, index, name, row_len: int):
        data = ItemData(row_len, index, name)
        self.data[row_len][index] = data

    def create_list(self, row: ui.row, data: ListData):
        with row:
            with ui.list().props("bordered separator") as ls:
                ui.item_label(data.current).props("header").classes("text-bold")
                ui.separator()

                for index in range(len(data.list_to_reco)):
                    name = data.list_to_reco[index]
                    self.create_item(index, name, data.row_len)

                if row == self.homepage_row and self.ItemPageData.reverse:
                    ls.move(row, 0)
                elif row == self.other_page_row and not self.ItemPageData.reverse:
                    ls.move(row, 0)

    def create_item(self, index: int, name: str, row_len: int):
        data: ItemData = self.data[row_len][index]

        with ui.item(on_click=lambda data=data: self.on_click_item(data)):  # type: ignore
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
