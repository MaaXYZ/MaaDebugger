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


def main():
    set_toml_ver()


if __name__ == "__main__":
    main()
