# pip install MaaDebugger pyinstaller

import os
from pathlib import Path


os.environ["MAAFW_BINARY_PATH"] = str(Path.cwd() / "bin")

if __name__ == "__main__":
    from MaaDebugger import MaaDebugger

    MaaDebugger.run()
