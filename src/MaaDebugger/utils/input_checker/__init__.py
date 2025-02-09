import json
from typing import Optional


def json_style_str(data: str) -> Optional[str]:
    if not data:
        return

    try:
        json.loads(data)
    except json.decoder.JSONDecodeError as e:
        return f"JSONDecodeError: {e}"
    except:
        return "Unknown error"
