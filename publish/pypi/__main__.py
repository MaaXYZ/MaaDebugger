import importlib.util
from pathlib import Path

spec = importlib.util.find_spec("maa")
if spec and spec.origin:
    module_path = spec.origin
    node_path = Path(module_path).parent / "bin"
    print(node_path)
else:
    raise ModuleNotFoundError("Module 'maa' is not found")
