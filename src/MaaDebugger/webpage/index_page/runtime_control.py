import asyncio
import os
from collections import defaultdict
from dataclasses import dataclass
from typing import Optional, Any, Dict, List
from queue import Queue
from threading import Lock

from nicegui import app, ui, background_tasks
from maa.resource import Resource, NotificationType

from ...maafw import (
    maafw,
    MyResourceEventSink,
    LaunchGraphContextEventSink,
    LaunchGraphTaskerEventSink,
    LaunchGraphManager,
    LaunchGraph,
)
from ...webpage.components.status_indicator import Status, StatusIndicator
from ...webpage.reco_page import RecoData
from .global_status import GlobalStatus


# 全局状态机管理器实例
launch_graph_manager = LaunchGraphManager()


STORAGE = app.storage.general

PAGINATION_DOCS_URL = "https://github.com/MaaXYZ/MaaDebugger/discussions/120"
# Set None to disable pagination or warning
PER_PAGE_ITEM_NUM: Optional[int] = (
    int(os.getenv("MAADBG_PER_PAGE_ITEM_NUM") or 0) or None
)
# When the pagination is disabled and the item number reaches this value, a warning will be displayed.
ITEM_NUMBER_WARING: Optional[int] = 400


@dataclass
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
    next_list: List[str]


def main():
    reco_data = RecognitionRow()
    reco_data.init_elements()


class RecognitionRow:
    def __init__(self) -> None:
        self.row_len = 0
        self.data = defaultdict(dict)
        self.list_data_map: dict[int, ListData] = {}
        self._pending_messages: Queue = Queue()
        self._lock = Lock()

        self.register_sink()

    def register_sink(self):
        """Register the custom notification handler to maafw."""
        # 使用状态机驱动的 EventSink
        context_event_sink = LaunchGraphContextEventSink(
            graph_manager=launch_graph_manager
        )
        resource_event_sink = MyResourceEventSink(self.on_resource_loading)
        # 注册 TaskerEventSink 以捕获 Task 级别事件
        tasker_event_sink = LaunchGraphTaskerEventSink(
            graph_manager=launch_graph_manager
        )

        maafw.context_event_sink = context_event_sink
        maafw.resource_event_sink = resource_event_sink
        maafw.tasker_event_sink = tasker_event_sink

        # 订阅状态机变化（增量处理方式）
        launch_graph_manager.subscribe(self.on_graph_change)

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
        RecoData.data.clear()
        self.data.clear()
        self.list_data_map.clear()
        # 重置状态机
        launch_graph_manager.reset()

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

                for index in range(len(data.next_list)):
                    name = data.next_list[index]
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

    def on_graph_change(self, graph: LaunchGraph, msg: Dict[str, Any]):
        """
        状态机变化时的回调（增量处理）
        将消息放入队列，由 NiceGUI 主线程处理
        """
        msg_type = msg.get("msg", "")
        print(f"[DEBUG on_graph_change] msg_type={msg_type}, msg={msg}")

        # 将消息放入队列
        self._pending_messages.put(msg)

        # 使用 background_tasks 在主线程中处理消息
        background_tasks.create(self._process_pending_messages())

    async def _process_pending_messages(self):
        """在主线程中处理待处理的消息"""
        while not self._pending_messages.empty():
            try:
                msg = self._pending_messages.get_nowait()
                await self._handle_message(msg)
            except Exception as e:
                print(f"[ERROR] Failed to process message: {e}")

    async def _handle_message(self, msg: Dict[str, Any]):
        """处理单个消息"""
        msg_type = msg.get("msg", "")

        # 处理 NextList.Starting - 添加新列表
        if msg_type == "NextList.Starting":
            name = msg.get("name", "")
            next_names = msg.get("next_list", [])
            print(f"[DEBUG] NextList.Starting: name={name}, next_list={next_names}")
            self._on_next_list_starting(name, next_names)
            await maafw.screenshotter.refresh(False)

        # TODO: 等待嵌套处理 - RecognitionNode.Starting 创建独立 UI 条目
        # RecognitionNode 内部会触发 Recognition，但这些 Recognition 不属于任何 NextList
        # 需要为它们创建独立的 UI 条目
        elif msg_type == "RecognitionNode.Starting":
            name = msg.get("name", "")
            node_id = msg.get("node_id", 0)
            print(f"[DEBUG] RecognitionNode.Starting: name={name}, node_id={node_id}")
            self._on_reco_node_starting(name, node_id)
            await maafw.screenshotter.refresh(False)

        # 处理 Recognition.Succeeded/Failed - 更新识别状态
        elif msg_type in ("Recognition.Succeeded", "Recognition.Failed"):
            name = msg.get("name", "")
            reco_id = msg.get("reco_id", 0)
            hit = msg_type == "Recognition.Succeeded"
            print(f"[DEBUG] Recognition: name={name}, reco_id={reco_id}, hit={hit}")
            self._on_recognized(reco_id, name, hit)
            await maafw.screenshotter.refresh(False)

    def _on_recognized(self, reco_id: int, name: str, hit: bool):
        """处理识别完成事件"""
        found = False
        for item in self.data[self.row_len].values():
            if item.status == Status.PENDING and item.name == name:
                item.reco_id = reco_id
                item.status = Status.SUCCEEDED if hit else Status.FAILED
                found = True
                break

        # TODO: 等待嵌套处理 - 如果在当前 row 没找到匹配项，检查 RecognitionNode 创建的条目
        if not found:
            # 遍历所有 row 寻找匹配的 pending 条目
            for row_data in self.data.values():
                for item in row_data.values():
                    if item.status == Status.PENDING and item.name == name:
                        item.reco_id = reco_id
                        item.status = Status.SUCCEEDED if hit else Status.FAILED
                        found = True
                        break
                if found:
                    break

        RecoData.data[reco_id] = name, hit, maafw.get_node_data(name)

    def _on_reco_node_starting(self, name: str, node_id: int):
        """
        处理 RecognitionNode 开始事件
        TODO: 等待嵌套处理 - 为 RecognitionNode 内部的 Recognition 创建独立 UI 条目

        RecognitionNode 是在 custom recognizer/action 内部调用 ctx.run_recognition() 时触发的
        它不属于任何 NextList，所以需要创建独立的 UI 列表来展示
        """
        self.row_len += 1

        # 创建一个只包含单个识别项的列表
        # 列表标题使用 "RecognitionNode:" 前缀以区分
        list_data = ListData(self.row_len, f"RecognitionNode: {name}", [name])
        self.add_list_data(list_data)

        # 分页处理
        if (
            PER_PAGE_ITEM_NUM is not None
            and self.row_len / PER_PAGE_ITEM_NUM >= self.pagination.max
        ):
            self.pagination.max += 1
            self.homepage_row.clear()

        self.create_list(self.homepage_row, list_data)

    def _on_next_list_starting(self, current: str, next_list: List[str]):
        """处理 NextList 开始事件"""
        self.row_len += 1

        list_data = ListData(self.row_len, current, next_list)
        self.add_list_data(list_data)

        # 299/300 -> page:1 | 300/300 -> page:2
        if (
            PER_PAGE_ITEM_NUM is not None
            and self.row_len / PER_PAGE_ITEM_NUM >= self.pagination.max
        ):
            self.pagination.max += 1
            self.homepage_row.clear()

        # Display warning when pagination is disabled
        if PER_PAGE_ITEM_NUM is None and ITEM_NUMBER_WARING is not None:
            if self.row_len == ITEM_NUMBER_WARING:
                with self.homepage_row:
                    ui.html(
                        "<style>.multi-line-notification { white-space: pre-line; }</style>"
                    )
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

    def add_list_data(self, data: ListData):
        self.list_data_map[data.row_len] = data
        for index in range(len(data.next_list)):
            name = data.next_list[index]
            self.add_item_data(index, name, data.row_len)

    def on_resource_loading(
        self,
        _: Resource,
        noti_type: NotificationType,
        detail: Any,
    ):
        if noti_type == NotificationType.Failed:
            with self.homepage_row:
                ui.notification(
                    f"Fail to load: {detail.path}",
                    position="bottom-right",
                    type="negative",
                    timeout=20,
                    close_button=True,
                )
