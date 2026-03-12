import os
import sys
import platform
import importlib.util
import subprocess
from pathlib import Path


def main() -> int:
    spec = importlib.util.find_spec("maa")
    if spec and spec.origin:
        module_path = spec.origin
        LIB_PATH = Path(module_path).parent / "bin"
    else:
        print("Please run `pip install maafw`")
        raise ModuleNotFoundError("Module 'maa' is not found")

    OS_TYPE = platform.system()
    EXE_PATH = Path(__file__).parent
    if OS_TYPE == "Windows":
        EXE_PATH = EXE_PATH / "MaaDebugger.exe"
    else:
        EXE_PATH = EXE_PATH / "MaaDebugger"

    if not LIB_PATH.is_dir():
        raise FileNotFoundError(f"Library path not found: {LIB_PATH}")
    if not EXE_PATH.is_file():
        raise FileNotFoundError(f"Executable not found: {EXE_PATH}")

    ENV = {
        **os.environ,
        "MAADBG_CHANNEL": "pypi",
        "MAADBG_CHANNEL_MANAGER": "pypi",
        "MAADBG_CHANNEL_LIB_PATH": str(LIB_PATH),
    }

    try:
        result = subprocess.run(
            [str(EXE_PATH), *sys.argv[1:]],
            env=ENV,
            shell=False,
            check=False,
        )
    except Exception as e:
        print(f"System: {OS_TYPE}", file=sys.stderr)
        print(f"Lib Path: {LIB_PATH}", file=sys.stderr)
        print(f"Exe Path: {EXE_PATH}", file=sys.stderr)
        raise e

    finally:
        sys.exit(result.returncode)


if __name__ == "__main__":
    main()
