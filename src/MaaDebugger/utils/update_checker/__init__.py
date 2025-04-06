from enum import auto, Enum
from typing import Optional, Union

import httpx
import semver
from packaging.version import parse as parse_version

from ... import __version__

GITHUB_API = "https://api.github.com/repos/MaaXYZ/MaaDebugger/releases/latest"
PYPI_API = "https://pypi.org/pypi/MaaDebugger/json"


class CheckStatus(Enum):
    FAILED = auto()
    SKIPPED = auto()


async def get_github() -> Optional[str]:  # -> 'v1.8.0-beta.1'
    try:
        async with httpx.AsyncClient(trust_env=False) as client:
            req = await client.get(GITHUB_API, timeout=5)
            if req.status_code == 200:
                return req.json().get("tag_name")
    except Exception as e:
        print(f"WARNING: Failed to GET Github API", e)


async def get_pypi() -> Optional[str]:  # -> '1.8.0b1'
    try:
        async with httpx.AsyncClient(trust_env=False) as client:
            req = await client.get(PYPI_API, timeout=5)
            if req.status_code == 200:
                return req.json().get("info", {}).get("version")
    except Exception as e:
        print(f"WARNING: Failed to GET PyPi API", e)


def compare_tag_name(tag_name: str) -> bool:
    if semver.compare(tag_name, __version__.tag_name.lstrip("v")) == 1:
        return True
    else:
        return False


def compare_version(version: str) -> bool:
    remote_ver = parse_version(version)
    current_ver = parse_version(__version__.version)

    if remote_ver > current_ver:
        return True
    else:
        return False


async def check_update() -> Union[CheckStatus, str, None]:  # PyPi -> Github -> FAILED
    """
    If updatable, return a version str (like 1.8.0b1); else return None\n
    If checking was skipped, return 'SKIPPED'\n
    If checking is failed, return 'FAILED'
    """
    if __version__.version == "DEBUG" or __version__.tag_name == "DEBUG":
        return CheckStatus.SKIPPED

    if pypi := await get_pypi():
        if compare_version(pypi):
            return pypi
        else:
            return None

    if github := await get_github():
        if compare_tag_name(github.lstrip("v")):
            return str(parse_version(github.lstrip("v")))
        else:
            return None

    return CheckStatus.FAILED
