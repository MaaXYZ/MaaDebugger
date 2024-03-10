from pathlib import Path
import asyncio
import sys


async def run_task(
    install_dir: Path, adb_path: Path, adb_address: str, resource_dir: Path, task: str
):
    binding_dir = install_dir / "binding" / "Python"
    if not binding_dir.exists():
        return "Binding directory does not exist"

    binding_dir = str(binding_dir)
    if binding_dir not in sys.path:
        sys.path.insert(0, binding_dir)

    from maa.library import Library
    from maa.resource import Resource
    from maa.controller import AdbController
    from maa.instance import Instance
    from maa.toolkit import Toolkit

    version = Library.open(install_dir / "bin")
    if not version:
        return "Failed to open MaaFramework"

    print(f"MaaFw Version: {version}")

    Toolkit.init_config()

    controller = AdbController(adb_path, adb_address)
    connected = await controller.connect()
    if not connected:
        return "Failed to connect to ADB"

    resource = Resource()
    loaded = await resource.load(resource_dir)
    if not loaded:
        return "Failed to load resource"

    maa_inst = Instance()
    maa_inst.bind(resource, controller)
    inited = maa_inst.inited
    if not inited:
        return "Failed to init MaaFramework instance"

    ret = await maa_inst.run_task(task, {})
    return f"Task returned: {ret}"
