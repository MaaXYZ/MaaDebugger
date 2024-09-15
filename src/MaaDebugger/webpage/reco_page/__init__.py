from typing import Dict, Tuple

from nicegui import ui

from ...utils import cvmat_to_image
from ...maafw import maafw, RecognitionDetail

class RecoData:
    data: Dict[int, Tuple[str, bool]] = {}


@ui.page("/reco/{reco_id}")
async def reco_page(reco_id: int):
    if reco_id == 0 or not reco_id in RecoData.data:
        ui.markdown("## Not Found")
        return

    name, hit = RecoData.data[reco_id]
    status = hit and "✅" or "❌"
    title = f"{status} {name} ({reco_id})"

    ui.page_title(title)
    ui.markdown(f"## {title}")

    ui.separator()

    details: RecognitionDetail = await maafw.get_reco_detail(reco_id)
    if not details:
        ui.markdown("## Not Found")
        return

    ui.markdown(f"#### Hit: {str(details.box)}")

    for draw in details.draws:
        ui.image(cvmat_to_image(draw)).props("fit=scale-down")

    ui.json_editor({"content": {"json": details.detail}, "readOnly": True})
