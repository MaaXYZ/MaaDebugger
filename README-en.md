<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<div align="center">

# MaaDebugger

<a href="https://pypi.org/project/MaaDebugger/" target="_blank"><img alt="pypi" src="https://img.shields.io/badge/PyPI-3775A9?logo=pypi&logoColor=white"></a>
<img alt="pypi-downloads" src="https://img.shields.io/pypi/dm/MaaDebugger?label=Downloads">

<img alt="release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?label=Release">
<img alt=pre-release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?include_prereleases&label=Pre-Release">
<img alt="activity" src="https://img.shields.io/github/commit-activity/m/MaaXYZ/MaaDebugger?color=%23ff69b4&label=Commit+Activity">

**[简体中文](./README.md) | [English](./README-en.md)**

</div>

## Requirement

Python >= 3.9

## Installation

```bash
pip install MaaDebugger
```

## Update

```bash
pip install MaaDebugger MaaFW --upgrade
```

## Usage

### Start from CLI

```bash
python -m MaaDebugger
```

#### Specify port

MaaDebugger uses port **8011** by default. You can specify the port MaaDebugger runs on by using the --port [port] option. For example, to run MaaDebugger on port **8080**:

```bash
python -m MaaDebugger --port 8080
```

For other command line arguments, please use `python -m MaaDebugger -h` to view.

### Start with script

```python
from MaaDebugger import MaaDebugger

MaaDebugger.run()
```

## Development of MaaDebugger itself

```bash
cd src
python -m MaaDebugger
```

or

Using VSCode, press F5 in the project directory.
