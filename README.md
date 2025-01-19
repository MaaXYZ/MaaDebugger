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

### 自定义动作/识别器

- 自动加载的目录为和`pipeline`同级别的`custom`文件夹
- 其内部需要有`action`和`recognition`文件夹
- custom程序应根据类型放置在指定文件夹内
- custom程序应为一个文件夹,`文件夹名`==`custom程序名`==`pipeline内部调用名`
- custom程序的入口应为`main.py`
- 如果使用了maadebugger中不存在的库,需要自行放置在custom程序文件夹内进行本地调用

- 示例:

```tree
resource
├pipeline/
├image/
├model/
└custom/
  ├── action/
  │   ├── 动作1/
  │   │    └── main.py
  │   └── 动作2/
  │        └── main.py
  └── Recognition/
      ├── 识别器1/
      │    └── main.py
      └── 识别器2/
           └── main.py
```

- 其中,动作1,动作2,识别器1,识别器2为在pipeline中所使用的名字,比如

```json
"我的自定义任务": {
        "recognition": "Custom",
        "custom_recognition": "识别器1",
        "action": "Custom",
        "custom_action": "动作1"
    }
```

- `main.py`中要求对象名和文件夹相同,比如

```python
  class 识别器1(CustomRecognition):
    def analyze(context, ...):
        # 获取图片，然后进行自己的图像操作
        image = context.tasker.controller.cached_image
        # 返回图像分析结果
        return AnalyzeResult(box=(10, 10, 100, 100))

 ```


## 开发 MaaDebugger 

```bash
cd src
python -m MaaDebugger
```

或者

使用 VSCode，在项目目录中按下 F5
