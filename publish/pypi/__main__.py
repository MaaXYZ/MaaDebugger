import os
import platform
import importlib.util
import subprocess
from pathlib import Path

spec = importlib.util.find_spec("maa")
if spec and spec.origin:
    module_path = spec.origin
    LIB_PATH = Path(module_path).parent / "bin"
else:
    print("Module 'maa' is not found")
    print("Please run `pip install maafw`")
    raise ModuleNotFoundError("Module 'maa' is not found")

OS_TYPE = platform.system()
EXE_PATH = Path(__file__).parent
if OS_TYPE == "Windows":
    EXE_PATH = EXE_PATH / "MaaDebugger.exe"
else:
    EXE_PATH = EXE_PATH / "MaaDebugger"

if not LIB_PATH.exists():
    raise FileNotFoundError(f"Library path not found: {LIB_PATH}")
if not EXE_PATH.exists():
    raise FileNotFoundError(f"Executable not found: {EXE_PATH}")

ENV = {
    **os.environ,
    "MAADBG_CHANNEL": "pypi",
    "MAADBG_CHANNEL_LIB_PATH": str(LIB_PATH),
}

try:
    subprocess.run(EXE_PATH, env=ENV)
except Exception as e:
    # TODO:: Catch
    raise e
