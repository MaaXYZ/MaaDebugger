from typing import Dict, Tuple, Optional

from nicegui import ui

from ...utils import cvmat_to_image
from ...maafw import maafw, RecognitionDetail


class RecoData:
    data: Dict[int, Tuple[str, bool, dict]] = {}


@ui.page("/reco/{reco_id}")
async def reco_page(reco_id: int):
    if reco_id == 0 or not reco_id in RecoData.data:
        ui.markdown("## Not Found")
        return

    name, hit, node_data = RecoData.data[reco_id]
    status = hit and "✅" or "❌"
    title = f"{status} {name} ({reco_id})"

    ui.page_title(title)
    ui.markdown(f"## {title}")

    ui.separator()

    details: Optional[RecognitionDetail] = await maafw.get_reco_detail(reco_id)
    if not details:
        ui.markdown("## Not Found")
        return

    ui.markdown(f"#### {details.algorithm}")
    ui.markdown(f"#### {details.best_result}")

    for draw in details.draw_images:
        ui.image(cvmat_to_image(draw)).props("fit=scale-down")

    with ui.row():
        ui.json_editor({"content": {"json": details.raw_detail}, "readOnly": True})
        ui.json_editor({"content": {"json": node_data}, "readOnly": True})
