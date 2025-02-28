import json
from pathlib import Path
from typing import Optional


def hwnd(data: str) -> Optional[str]:
    if not data:
        return

    try:
        int(data, 16)
    except:
        return "Please input hexadecimal numbers."


def json_style_str(data: str) -> Optional[str]:
    if not data:
        return

    try:
        json.loads(data)
    except json.decoder.JSONDecodeError as e:
        return f"JSONDecodeError: {e}"
    except Exception as e:
        return str(e)


def paths_exist(data: str) -> Optional[str]:
    if not data:
        return

    not_exist_paths = []
    msg = ""

    paths = [Path(p) for p in data.split("\n") if p]
    for p in paths:
        if not p.exists():
            not_exist_paths.append(str(p))

    if not_exist_paths:
        if len(not_exist_paths) == 1:
            msg += f"Path not exist: {not_exist_paths[0]}"
        else:
            msg += f"Paths not exist: {not_exist_paths}"

    if msg:
        return msg
