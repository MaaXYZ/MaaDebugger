from nicegui import ui

from .check import check_update, CheckStatus


async def main():
    result = await check_update()
    if result is None:
        ui.notification(
            "No stable update available.",
            type="info",
            position="bottom-right",
            timeout=5,
            close_button=True,
        )

    if result == CheckStatus.SKIPPED:
        ui.notification(
            "Update check skipped.",
            type="info",
            position="bottom-right",
            timeout=None,
            close_button=True,
        )
    elif result == CheckStatus.FAILED:
        ui.notification(
            "Update check failed. Please check your network connection.",
            type="negative",
            position="bottom-right",
            timeout=None,
            close_button=True,
        )
    elif type(result) == str:
        ui.notification(
            f"New version {result} is available!",
            type="positive",
            position="bottom-right",
            timeout=None,
            close_button=True,
        )
