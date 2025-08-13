import platform
from enum import Enum, auto


class OSTypeEnum(Enum):
    Windows = auto()
    Linux = auto()
    macOS = auto()
    Unknown = auto()


def get_os_type() -> OSTypeEnum:
    system = platform.system()
    if system == "Windows":
        return OSTypeEnum.Windows
    elif system == "Linux":
        return OSTypeEnum.Linux
    elif system == "Darwin":
        return OSTypeEnum.macOS
    else:
        return OSTypeEnum.Unknown
