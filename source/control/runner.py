from pathlib import Path
import sys

specified_binding_dir = None
specified_bin_dir = None

if len(sys.argv) > 2:
    specified_binding_dir = Path(sys.argv[1])
    specified_bin_dir = Path(sys.argv[2])
    print(
        f"specified binding_dir: {specified_binding_dir}, bin_dir: {specified_bin_dir}"
    )

latest_install_dir = None
latest_adb_path = None
latest_adb_address = None
latest_resource_dir = None

resource = None
controller = None
instance = None


async def run_task(
    install_dir: Path, adb_path: Path, adb_address: str, resource_dir: Path, task: str
):
    binding_dir = (
        specified_binding_dir
        and specified_binding_dir
        or (install_dir / "binding" / "Python")
    )
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

    global latest_install_dir, latest_adb_path, latest_adb_address

    if latest_install_dir != install_dir:
        bin_dir = specified_bin_dir and specified_bin_dir or (install_dir / "bin")
        version = Library.open(bin_dir)
        if not version:
            return "Failed to open MaaFramework"
        print(f"MaaFw Version: {version}")

    latest_install_dir = install_dir

    Toolkit.init_option("~/.maafw")

    global resource, controller, instance

    if latest_adb_path != adb_path or latest_adb_address != adb_address:
        controller = AdbController(adb_path, adb_address)
        connected = await controller.connect()
        if not connected:
            return "Failed to connect to ADB"

    latest_adb_path = adb_path
    latest_adb_address = adb_address

    if not resource:
        resource = Resource()

    # reload every time
    loaded = resource.clear() and await resource.load(resource_dir)
    if not loaded:
        return "Failed to load resource"

    if not instance:
        instance = Instance()

    instance.bind(resource, controller)
    inited = instance.inited
    if not inited:
        return "Failed to init MaaFramework instance"

    ret = await instance.run_task(task, {})
    return f"Task returned: {ret}"


async def stop_task():
    if instance:
        await instance.stop()
