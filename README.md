# MaaDebugger

**[简体中文](./README.md) | [English](./README-en.md)**

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
