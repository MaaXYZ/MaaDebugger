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
        async with httpx.AsyncClient() as client:
            req = await client.get(url, timeout=5)
            if req.status_code == 200:
                return req.json().get("info", {}).get("version", None)
            else:
                return None
    except Exception as e:
        print(f"WARNING: Failed to check update from {url}", e)
        return None


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
    When succeed, if updatable, return a version str (like 1.8.0b1); else return None\n
    If checking was skipped, return 'SKIPPED'\n
    If checking was failed, return 'FAILED'
    """
    if "PASS" in (__version__.tag_name, __version__.version):
        return CheckStatus.SKIPPED
    elif "FAILED" in (__version__.tag_name, __version__.version):
        return CheckStatus.FAILED

    else:
        pypi = get_from_pypi(PYPI_API)
        tsinghua_pypi = get_from_pypi(TSINGHUA_PYPI_API)

        vers = await asyncio.gather(pypi, tsinghua_pypi, return_exceptions=True)

        check_number = len(vers)
        check_succeed_number = 0

        for ver in vers:
            if type(ver) == str:
                check_succeed_number += 1
        print(
            f"NOTICE: Update Check: {check_number}, Succeed: {check_succeed_number}, Failed: {check_number-check_succeed_number}"
        )

        if all(ver is None for ver in vers):
            return CheckStatus.FAILED

        for ver in vers:
            if type(ver) == str:
                if compare_version(ver):
                    return ver
        return None
