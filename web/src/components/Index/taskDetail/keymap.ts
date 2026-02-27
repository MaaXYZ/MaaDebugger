export default function getKey(value: number, controller_type: string): string {
  if (controller_type === "win32") {
    return win32.get(value) ?? "";
  }
  if (controller_type === "adb") {
    return adb.get(value) ?? "";
  }
  return "Unknown";
}

const win32: Map<number, string> = new Map([
  // Mouse buttons
  [1, "Left Mouse"], // VK_LBUTTON 0x01
  [2, "Right Mouse"], // VK_RBUTTON 0x02
  [4, "Middle Mouse"], // VK_MBUTTON 0x04
  [5, "X1 Mouse"], // VK_XBUTTON1 0x05
  [6, "X2 Mouse"], // VK_XBUTTON2 0x06

  // Control keys
  [8, "Backspace"], // VK_BACK 0x08
  [9, "Tab"], // VK_TAB 0x09
  [13, "Enter"], // VK_RETURN 0x0D
  [16, "Shift"], // VK_SHIFT 0x10
  [17, "Ctrl"], // VK_CONTROL 0x11
  [18, "Alt"], // VK_MENU 0x12
  [19, "Pause"], // VK_PAUSE 0x13
  [20, "Caps Lock"], // VK_CAPITAL 0x14
  [27, "Esc"], // VK_ESCAPE 0x1B
  [32, "Space"], // VK_SPACE 0x20
  [33, "Page Up"], // VK_PRIOR 0x21
  [34, "Page Down"], // VK_NEXT 0x22
  [35, "End"], // VK_END 0x23
  [36, "Home"], // VK_HOME 0x24
  [37, "←"], // VK_LEFT 0x25
  [38, "↑"], // VK_UP 0x26
  [39, "→"], // VK_RIGHT 0x27
  [40, "↓"], // VK_DOWN 0x28
  [44, "Print Screen"], // VK_SNAPSHOT 0x2C
  [45, "Insert"], // VK_INSERT 0x2D
  [46, "Delete"], // VK_DELETE 0x2E

  // Number keys 0-9
  [48, "0"], // 0x30
  [49, "1"], // 0x31
  [50, "2"], // 0x32
  [51, "3"], // 0x33
  [52, "4"], // 0x34
  [53, "5"], // 0x35
  [54, "6"], // 0x36
  [55, "7"], // 0x37
  [56, "8"], // 0x38
  [57, "9"], // 0x39

  // Letter keys A-Z
  [65, "A"], // 0x41
  [66, "B"], // 0x42
  [67, "C"], // 0x43
  [68, "D"], // 0x44
  [69, "E"], // 0x45
  [70, "F"], // 0x46
  [71, "G"], // 0x47
  [72, "H"], // 0x48
  [73, "I"], // 0x49
  [74, "J"], // 0x4A
  [75, "K"], // 0x4B
  [76, "L"], // 0x4C
  [77, "M"], // 0x4D
  [78, "N"], // 0x4E
  [79, "O"], // 0x4F
  [80, "P"], // 0x50
  [81, "Q"], // 0x51
  [82, "R"], // 0x52
  [83, "S"], // 0x53
  [84, "T"], // 0x54
  [85, "U"], // 0x55
  [86, "V"], // 0x56
  [87, "W"], // 0x57
  [88, "X"], // 0x58
  [89, "Y"], // 0x59
  [90, "Z"], // 0x5A

  // Windows keys
  [91, "Left Win"], // VK_LWIN 0x5B
  [92, "Right Win"], // VK_RWIN 0x5C
  [93, "Menu"], // VK_APPS 0x5D

  // Numpad keys
  [96, "0"], // VK_NUMPAD0 0x60
  [97, "1"], // VK_NUMPAD1 0x61
  [98, "2"], // VK_NUMPAD2 0x62
  [99, "3"], // VK_NUMPAD3 0x63
  [100, "4"], // VK_NUMPAD4 0x64
  [101, "5"], // VK_NUMPAD5 0x65
  [102, "6"], // VK_NUMPAD6 0x66
  [103, "7"], // VK_NUMPAD7 0x67
  [104, "8"], // VK_NUMPAD8 0x68
  [105, "9"], // VK_NUMPAD9 0x69
  [106, "*"], // VK_MULTIPLY 0x6A
  [107, "+"], // VK_ADD 0x6B
  [109, "-"], // VK_SUBTRACT 0x6D
  [110, "."], // VK_DECIMAL 0x6E
  [111, "/"], // VK_DIVIDE 0x6F

  // Function keys
  [112, "F1"], // VK_F1 0x70
  [113, "F2"], // VK_F2 0x71
  [114, "F3"], // VK_F3 0x72
  [115, "F4"], // VK_F4 0x73
  [116, "F5"], // VK_F5 0x74
  [117, "F6"], // VK_F6 0x75
  [118, "F7"], // VK_F7 0x76
  [119, "F8"], // VK_F8 0x77
  [120, "F9"], // VK_F9 0x78
  [121, "F10"], // VK_F10 0x79
  [122, "F11"], // VK_F11 0x7A
  [123, "F12"], // VK_F12 0x7B

  // Lock keys
  [144, "Num Lock"], // VK_NUMLOCK 0x90
  [145, "Scroll Lock"], // VK_SCROLL 0x91

  // Modifier keys (left/right)
  [160, "Left Shift"], // VK_LSHIFT 0xA0
  [161, "Right Shift"], // VK_RSHIFT 0xA1
  [162, "Left Ctrl"], // VK_LCONTROL 0xA2
  [163, "Right Ctrl"], // VK_RCONTROL 0xA3
  [164, "Left Alt"], // VK_LMENU 0xA4
  [165, "Right Alt"], // VK_RMENU 0xA5

  // OEM keys (punctuation/symbols)
  [186, ";"], // VK_OEM_1 0xBA  ;:
  [187, "="], // VK_OEM_PLUS 0xBB  =+
  [188, ","], // VK_OEM_COMMA 0xBC  ,<
  [189, "-"], // VK_OEM_MINUS 0xBD  -_
  [190, "."], // VK_OEM_PERIOD 0xBE  .>
  [191, "/"], // VK_OEM_2 0xBF  /?
  [192, "`"], // VK_OEM_3 0xC0  `~
  [219, "["], // VK_OEM_4 0xDB  [{
  [220, "\\"], // VK_OEM_5 0xDC  \|
  [221, "]"], // VK_OEM_6 0xDD  ]}
  [222, "'"], // VK_OEM_7 0xDE  '"
]);

const adb: Map<number, string> = new Map([
  [0, "ACTION_DOWN"],
  [1, "ACTION_UP"],
]);
