import sys

import tomlkit


def set_toml_ver():
    TOML_FILE = "pyproject.toml"

    version = sys.argv[1]

    print("Setting version to:", version)

    with open(TOML_FILE, "r") as f:
        py_project = tomlkit.load(f)
        py_project["project"]["version"] = version

    with open(TOML_FILE, "w") as f:
        tomlkit.dump(py_project, f)


def set_python_ver():
    VER_FILE = "src/MaaDebugger/__version__.py"

    tag_name: str = sys.argv[1]
    ver = (
        tag_name.lstrip("v")
        .replace("-", "")
        .replace("alpha.", "a")
        .replace("beta.", "b")
    )

    with open(VER_FILE, "w", encoding="utf-8") as f:
        f.write(f'version = "{ver}"\n')
        f.write(f'tag_name = "{tag_name}"\n')

    print(f"version = {ver} | tag_name = {tag_name}")


def main():
    set_toml_ver()
    set_python_ver()


if __name__ == "__main__":
    main()
