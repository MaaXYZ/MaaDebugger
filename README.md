<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<div align="center">

# MaaDebugger

<a href="https://pypi.org/project/MaaDebugger/" target="_blank"><img alt="pypi" src="https://img.shields.io/pypi/dm/MaaDebugger?logo=pypi&label=PyPI"></a>

<a href="https://github.com/MaaXYZ/MaaDebugger/releases/latest" target="_blank"><img alt="release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?label=Release"></a>
<a href="https://github.com/MaaXYZ/MaaDebugger/releases" target="_blank"><img alt="pre-release" src="https://img.shields.io/github/v/release/MaaXYZ/MaaDebugger?include_prereleases&label=Pre-Release"></a>
<a href="https://github.com/MaaXYZ/MaaDebugger/commits/main/" target="_blank"><img alt="activity" src="https://img.shields.io/github/commit-activity/m/MaaXYZ/MaaDebugger?color=%23ff69b4&label=Commit+Activity"></a>

**[简体中文](./README.md) | [English](./README-en.md)**

</div>

## 需求版本

- Python >= 3.9,<= 3.13
- nicegui >= 2.21,< 3.0

## 快速开始

### 使用 uv 
提供更快的安装速度与环境隔离。

- **安装为全局工具**:
  ```bash
  uv tool install MaaDebugger
  MaaDebugger
  ```
- **临时运行 (免安装)**:
  ```bash
  uvx MaaDebugger
  ```
- **更新**:
  ```bash
  uv tool upgrade MaaDebugger
  ```

---

### 使用 Python (pip)
传统的 Python 安装方式。

- **安装**:
  ```bash
  python -m pip install MaaDebugger
  ```
- **运行**:
  ```bash
  python -m MaaDebugger
  ```
- **更新**:
  ```bash
  python -m pip install MaaDebugger MaaFW --upgrade
  ```

---

### 指定端口

MaaDebugger 默认使用端口 **8011**。你可以通过使用 --port [port] 选项来指定 MaaDebugger 运行的端口。例如，要在端口 **8080** 上运行 MaaDebugger：
无论使用哪种方式，均支持 `--port` 参数。例如在端口 **8080** 上运行：

```bash
# uv 
MaaDebugger --port 8080

# uvx (免安装)
uvx MaaDebugger --port 8080

# python
python -m MaaDebugger --port 8080
```

## 开发 MaaDebugger

```bash
cd src
python -m MaaDebugger
```

或者

使用 VSCode，在项目目录中按下 `F5`
