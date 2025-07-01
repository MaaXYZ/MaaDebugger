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

## 需求版本

Python >= 3.9

## 安装

```bash
python -m pip install MaaDebugger
```

## 更新

```bash
python -m pip install MaaDebugger MaaFW --upgrade
```

## 使用

```bash
python -m MaaDebugger
```

### 指定端口

MaaDebugger 默认使用端口 **8011**。你可以通过使用 --port [port] 选项来指定 MaaDebugger 运行的端口。例如，要在端口 **8080** 上运行 MaaDebugger

```bash
python -m MaaDebugger --port 8080
```

## 开发 MaaDebugger

```bash
cd src
python -m MaaDebugger
```

或者

使用 VSCode，在项目目录中按下 F5
