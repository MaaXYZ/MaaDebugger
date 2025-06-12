import asyncio
from enum import auto, Enum
from typing import Any, Optional, Union

import httpx
from packaging.version import parse as parse_version

from ... import __version__

PYPI_API = "https://pypi.org/pypi/MaaDebugger/json"
TSINGHUA_PYPI_API = "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/json/maadebugger"


class CheckStatus(Enum):
    FAILED = auto()
    SKIPPED = auto()


async def get_from_pypi(url: str) -> Optional[str]:  # -> '1.8.0b1'
    try:
        async with httpx.AsyncClient(trust_env=False) as client:
            req = await client.get(url, timeout=5)
            if req.status_code == 200:
                return req.json().get("info", {}).get("version")
    except Exception as e:
        print(f"WARNING: Failed to GET PyPi API", e)


def compare_version(ver: Union[str, Any]) -> bool:
    if type(ver) != str:
        return False

    remote_ver = parse_version(ver)
    current_ver = parse_version(__version__.version)

    if remote_ver > current_ver:
        return True
    else:
        return False


async def check_update() -> Union[CheckStatus, str, None]:
    """
    If updatable, return a version str (like 1.8.0b1); else return None\n
    If checking was skipped, return 'SKIPPED'\n
    If checking is failed, return 'FAILED'
    """
    if "PASS" in [__version__.tag_name, __version__.version]:
        return CheckStatus.SKIPPED
    elif "FAILED" in [__version__.tag_name, __version__.version]:
        return CheckStatus.FAILED

    else:
        pypi = get_from_pypi(PYPI_API)
        tsinghua_pypi = get_from_pypi(TSINGHUA_PYPI_API)

        vers = await asyncio.gather(pypi, tsinghua_pypi, return_exceptions=True)

        if "DEBUG" in [__version__.tag_name, __version__.version]:
            for ver in vers:
                if ver and type(ver) == str:
                    return ver
            return CheckStatus.FAILED

        for ver in vers:
            if ver and type(ver) == str:
                if compare_version(ver):
                    return ver
        return CheckStatus.FAILED
