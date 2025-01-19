# MaaDebugger

**[简体中文](./README.md) | [English](./README-en.md)**

## Requirement

Python >= 3.9

## Installation

```bash
python -m pip install MaaDebugger
```

## Update

```bash
python -m pip install MaaDebugger MaaFW --upgrade
```

## Usage

```bash
python -m MaaDebugger
```

### Specifying a Port

MaaDebugger uses port **8011** by default. You can specify a port to run MaaDebugger on by using the `--port [port]` option. For example, to run MaaDebugger on port **8080**:

```bash
python -m MaaDebugger --port 8080
```
## Custom Action/Recognition

- The automatically loaded directory is the custom folder at the same level as the pipeline.

- It needs to have action and recognition folders inside.

- Custom programs should be placed in the specified folders according to their types.

- A custom program should be a folder, and the folder name should be equal to the custom program name and the call name inside the pipeline.

- The entry point of a custom program should be main.py.

- If you use a library that does not exist in maadebugger, you need to place it in the custom program folder for local calling.

- Example:
  
```tree
resource
├pipeline/
├image/
├model/
└custom/
  ├── action/
  │   ├── action1/
  │   │    └── main.py
  │   └── action2/
  │        └── main.py
  └── Recognition/
      ├── Recognition1/
      │    └── main.py
      └── Recognition2/
           └── main.py
```

- Among them, "Action 1", "Action 2", "Recognizer 1", and "Recognizer 2" are the names used in the pipeline, for example:

```json
"MyRecognition": {
        "recognition": "Custom",
        "custom_recognition": "Recognition1",
        "action": "Custom",
        "custom_action": "action1"
    }
```

- In ```main.py```, it is required that the object name be the same as the folder name, for example:

```python
  class Recognition1(CustomRecognition):
    def analyze(context, ...):
        # Obtain the image and perform custom image processing
        image = context.tasker.controller.cached_image
        # Return image analysis result
        return AnalyzeResult(box=(10, 10, 100, 100))

 ```

## Development of MaaDebugger itself

```bash
cd src
python -m MaaDebugger
```

or

Using VSCode, press F5 in the project directory.
