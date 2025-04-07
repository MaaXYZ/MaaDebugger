import sys
import platform
from pathlib import Path

import nicegui
import packaging
import packaging.version

from ...maafw import maafw
from ... import __version__


class Infos:
    CWD = Path.cwd()
    DEBUG_DIR = CWD / "debug"
    OS_TYPE = platform.system()
    OS_VERSION = platform.version()
    OS_INFO = f"{OS_TYPE} {OS_VERSION}"
    MACHINE = platform.machine()
    PYTHON_VERSION = (
        f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}"
    )
    DBG_VERSION = __version__.version
    MAAFW_VERSION = str(packaging.version.parse(maafw.version))
    NICEGUI_VERSION = nicegui.__version__
