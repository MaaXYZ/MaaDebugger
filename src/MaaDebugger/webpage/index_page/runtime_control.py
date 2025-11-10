from collections import defaultdict
from dataclasses import dataclass
from typing import Optional
import asyncio
import os

from nicegui import app, ui
from nicegui.binding import bindable_dataclass

from ...maafw import maafw, MyContextEventSink
from ...webpage.components.status_indicator import Status, StatusIndicator
from ...webpage.reco_page import RecoData
from .global_status import GlobalStatus


STORAGE = app.storage.general

PAGINATION_DOCS_URL = "https://github.com/MaaXYZ/MaaDebugger/discussions/120"
# Set None to disable pagination or warning
PER_PAGE_ITEM_NUM: Optional[int] = (
    int(os.getenv("MAADBG_PER_PAGE_ITEM_NUM") or 0) or None
)
# When the pagination is disabled and the item number reaches this value, a warning will be displayed.
ITEM_NUMBER_WARING: Optional[int] = 400


@bindable_dataclass
class ItemData:
    col: int
    row: int
    name: str
    reco_id: int = 0
    status: Status = Status.PENDING


@dataclass
class ListData:
    row_len: int
    current: str
    list_to_reco: list[str]


def main():
    reco_data = RecognitionRow()
    reco_data.init_elements()


class RecognitionRow:
    def __init__(self) -> None:
        self.row_len = 0
        self.data = defaultdict(dict)
        self.list_data_map: dict[int, ListData] = {}

        self.register_notification_handler()

    def register_notification_handler(self):
        """Register the custom notification handler to maafw."""
        self.task_event_sink = MyContextEventSink(
            self.on_next_list_starting, self.on_recognized
        )

        maafw.event_sink = self.task_event_sink

    def init_elements(self):
        """Initialize the UI elements."""
        with ui.row():
            ui.button("Clear Items", icon="remove", on_click=self.clear_items).props(
                "no-caps"
            )
            ui.button(
                "Clear Items and Cache", icon="delete_forever", on_click=self.clear
            ).props("no-caps").bind_enabled_from(
                GlobalStatus, "task_running", lambda x: x != Status.RUNNING
            )
            self.reverse_switch = (
                ui.switch(
                    "Reverse",
                    value=STORAGE.get("items-reverse", True),
                    on_change=lambda x: self.on_reverse_switch_change(x.value),
                )
                .tooltip("Switch this will clear all items and cache.")
                .bind_enabled_from(
                    GlobalStatus, "task_running", lambda x: x != Status.RUNNING
                )
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

        if PER_PAGE_ITEM_NUM is None:
            self.pagination.set_visibility(False)

    async def on_reverse_switch_change(self, value: bool):
        await self.clear()
        STORAGE["items-reverse"] = value

    async def clear(self):
        await maafw.clear_cache()
        self.clear_items()

    def clear_items(self):
        self.row_len = 0
        self.data.clear()
        self.list_data_map.clear()

        self.homepage_row.clear()
        self.other_page_row.clear()
        self.pagination.max = 1
        self.pagination.set_value(1)

    def on_page_change(self, page: int):
        if PER_PAGE_ITEM_NUM is None:
            return

        self.other_page_row.clear()

        if page == 1:
            self.homepage_row.set_visibility(True)
            self.other_page_row.set_visibility(False)
        else:
            self.homepage_row.set_visibility(False)
            self.other_page_row.set_visibility(True)

            # Get all row_len in reverse order
            row_len_list = range(self.row_len, 0, -1)
            total_items = len(row_len_list)
            start_index = max((page - 1) * PER_PAGE_ITEM_NUM, 0)
            end_index = min(start_index + PER_PAGE_ITEM_NUM, total_items)

            # Get the row_len for the current page
            # NOTE: This row_len_list is reversed
            for row_len in row_len_list[start_index:end_index]:
                self.create_list(self.other_page_row, self.list_data_map[row_len])

    # maafw
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

        RecoData.data[reco_id] = name, hit, maafw.get_node_data(name)
        asyncio.run(maafw.screenshotter.refresh(False))

    # maafw
    def on_next_list_starting(self, current: str, list_to_reco: list[str]):
        self.row_len += 1

        list_data = ListData(self.row_len, current, list_to_reco)
        self.add_list_data(list_data)

        # 299/300 -> page:1 | 300/300 -> page:2
        if (
            PER_PAGE_ITEM_NUM is not None
            and self.row_len / PER_PAGE_ITEM_NUM >= self.pagination.max
        ):
            self.pagination.max += 1
            self.homepage_row.clear()

        # Display warning when pagination is disabled
        # In the future, we can consider to enable pagination by default
        if PER_PAGE_ITEM_NUM is None and ITEM_NUMBER_WARING is not None:
            if self.row_len == ITEM_NUMBER_WARING:
                with self.homepage_row:
                    ui.html(
                        "<style>.multi-line-notification { white-space: pre-line; }</style>"
                    )  # allow use \n in notification text
                    ui.notification(
                        f"Item number is reaching the {ITEM_NUMBER_WARING}! Please consider to enable pagination.\nFor more information, please see {PAGINATION_DOCS_URL}",
                        position="bottom-right",
                        type="warning",
                        timeout=None,
                        close_button=True,
                        multi_line=True,
                        classes="multi-line-notification",
                        actions=[
                            {
                                "label": "VIEW",
                                "color": "white",
                                ":handler": f"() => emitEvent('pagination-docs-clicked')",
                            }
                        ],
                    )
                    ui.on(
                        "pagination-docs-clicked",
                        lambda: ui.navigate.to(PAGINATION_DOCS_URL, new_tab=True),
                    )

        self.create_list(self.homepage_row, list_data)

        asyncio.run(maafw.screenshotter.refresh(False))

    def add_list_data(self, data: ListData):
        self.list_data_map[data.row_len] = data
        for index in range(len(data.list_to_reco)):
            name = data.list_to_reco[index]
            self.add_item_data(index, name, data.row_len)

    def add_item_data(self, index, name, row_len: int):
        data = ItemData(row_len, index, name)
        self.data[row_len][index] = data

    def create_list(self, row: ui.row, data: ListData):
        reverse: bool = self.reverse_switch.value

        with row:
            with ui.list().props("bordered separator") as ls:
                ls.set_visibility(False)  # The list will be hidden until prepared

                # reverse
                if row == self.homepage_row and reverse:
                    ls.move(row, 0)
                elif row == self.other_page_row and not reverse:
                    # As
                    # When in other page, we need to reverse the reverse logic
                    ls.move(row, 0)

                ui.item_label(data.current).props("header").classes("text-bold")
                ui.separator()

                for index in range(len(data.list_to_reco)):
                    name = data.list_to_reco[index]
                    self.create_items(index, name, data.row_len)

                ls.set_visibility(True)

    def create_items(self, index: int, name: str, row_len: int):
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
        if data.reco_id == 0:
            return
        else:
            print(
                f"on_click_item ({data.col}, {data.row}): {data.name} ({data.reco_id})"
            )
            ui.navigate.to(f"reco/{data.reco_id}", new_tab=True)
