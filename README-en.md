<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<div align="center">

# MaaDebugger

<a href="https://pypi.org/project/MaaDebugger/" target="_blank"><img alt="pypi" src="https://img.shields.io/pypi/dm/MaaDebugger?logo=pypi&label=PyPI"></a>

<a href="https://github.com/MaaXYZ/MaaDebugger/releases/latest" target="_blank"><img alt="release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?label=Release"></a>
<a href="https://github.com/MaaXYZ/MaaDebugger/releases" target="_blank"><img alt="pre-release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?include_prereleases&label=Pre-Release"></a>
<a href="https://github.com/MaaXYZ/MaaDebugger/commits/main/" target="_blank"><img alt="activity" src="https://img.shields.io/github/commit-activity/m/MaaXYZ/MaaDebugger?color=%23ff69b4&label=Commit+Activity"></a>
s

**[简体中文](./README.md) | [English](./README-en.md)**

</div>

> [!NOTE]
> We are currently refactoring MaaDebugger. You can visit https://github.com/MaaXYZ/MaaDebugger/issues/163 for more details and to get an early preview version.

## Requirement

- Python >= 3.9, <= 3.13
- nicegui >= 2.21,< 3.0

## Quick Start

### Using uv

Provides faster installation speed and environment isolation.

- **Install as a global tool**:
  ```bash
  uv tool install MaaDebugger
  MaaDebugger
  ```
- **Run temporarily (No installation required)**:
  ```bash
  uvx MaaDebugger
  ```
- **Update**:
  ```bash
  uv tool upgrade MaaDebugger
  ```

---

### Using Python (pip)

Traditional Python installation method.

- **Install**:
  ```bash
  python -m pip install MaaDebugger
  ```
- **Run**:
  ```bash
  python -m MaaDebugger
  ```
- **Update**:
  ```bash
  python -m pip install MaaDebugger MaaFW --upgrade
  ```

---

### Specify Port

MaaDebugger uses port **8011** by default. You can specify a port to run MaaDebugger on by using the `--port [port]` option. For example, to run MaaDebugger on port **8080**:
The `--port` parameter is supported across all methods. For example, to run on port **8080**:

```bash
# uv
MaaDebugger --port 8080

# uvx (No installation required)
uvx MaaDebugger --port 8080

# python
python -m MaaDebugger --port 8080
```

## Development of MaaDebugger itself

```bash
cd src
python -m MaaDebugger
```

or

Using VSCode, press `F5` in the project directory.
