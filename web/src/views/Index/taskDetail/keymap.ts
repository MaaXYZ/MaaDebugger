export default function getKey(value: number, controller_type: string): string {
  switch (controller_type) {
    case "win32":
      return win32.get(value) ?? "";
    case "adb":
      return adb.get(value) ?? "";
    case "gamepad":
      return gamepad.get(value) ?? "";
  }
  return "NULL";
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
  [0, "Unknown"], // KEYCODE_UNKNOWN
  [1, "Soft Left"], // KEYCODE_SOFT_LEFT
  [2, "Soft Right"], // KEYCODE_SOFT_RIGHT
  [3, "Home"], // KEYCODE_HOME
  [4, "Back"], // KEYCODE_BACK
  [5, "Call"], // KEYCODE_CALL
  [6, "End Call"], // KEYCODE_ENDCALL
  [7, "0"], // KEYCODE_0
  [8, "1"], // KEYCODE_1
  [9, "2"], // KEYCODE_2
  [10, "3"], // KEYCODE_3
  [11, "4"], // KEYCODE_4
  [12, "5"], // KEYCODE_5
  [13, "6"], // KEYCODE_6
  [14, "7"], // KEYCODE_7
  [15, "8"], // KEYCODE_8
  [16, "9"], // KEYCODE_9
  [17, "*"], // KEYCODE_STAR
  [18, "#"], // KEYCODE_POUND
  [19, "↑"], // KEYCODE_DPAD_UP
  [20, "↓"], // KEYCODE_DPAD_DOWN
  [21, "←"], // KEYCODE_DPAD_LEFT
  [22, "→"], // KEYCODE_DPAD_RIGHT
  [23, "DPad Center"], // KEYCODE_DPAD_CENTER
  [24, "Volume Up"], // KEYCODE_VOLUME_UP
  [25, "Volume Down"], // KEYCODE_VOLUME_DOWN
  [26, "Power"], // KEYCODE_POWER
  [27, "Camera"], // KEYCODE_CAMERA
  [28, "Clear"], // KEYCODE_CLEAR
  [29, "A"], // KEYCODE_A
  [30, "B"], // KEYCODE_B
  [31, "C"], // KEYCODE_C
  [32, "D"], // KEYCODE_D
  [33, "E"], // KEYCODE_E
  [34, "F"], // KEYCODE_F
  [35, "G"], // KEYCODE_G
  [36, "H"], // KEYCODE_H
  [37, "I"], // KEYCODE_I
  [38, "J"], // KEYCODE_J
  [39, "K"], // KEYCODE_K
  [40, "L"], // KEYCODE_L
  [41, "M"], // KEYCODE_M
  [42, "N"], // KEYCODE_N
  [43, "O"], // KEYCODE_O
  [44, "P"], // KEYCODE_P
  [45, "Q"], // KEYCODE_Q
  [46, "R"], // KEYCODE_R
  [47, "S"], // KEYCODE_S
  [48, "T"], // KEYCODE_T
  [49, "U"], // KEYCODE_U
  [50, "V"], // KEYCODE_V
  [51, "W"], // KEYCODE_W
  [52, "X"], // KEYCODE_X
  [53, "Y"], // KEYCODE_Y
  [54, "Z"], // KEYCODE_Z
  [55, ","], // KEYCODE_COMMA
  [56, "."], // KEYCODE_PERIOD
  [57, "Left Alt"], // KEYCODE_ALT_LEFT
  [58, "Right Alt"], // KEYCODE_ALT_RIGHT
  [59, "Left Shift"], // KEYCODE_SHIFT_LEFT
  [60, "Right Shift"], // KEYCODE_SHIFT_RIGHT
  [61, "Tab"], // KEYCODE_TAB
  [62, "Space"], // KEYCODE_SPACE
  [63, "Sym"], // KEYCODE_SYM
  [64, "Explorer"], // KEYCODE_EXPLORER
  [65, "Envelope"], // KEYCODE_ENVELOPE
  [66, "Enter"], // KEYCODE_ENTER
  [67, "Backspace"], // KEYCODE_DEL
  [68, "`"], // KEYCODE_GRAVE
  [69, "-"], // KEYCODE_MINUS
  [70, "="], // KEYCODE_EQUALS
  [71, "["], // KEYCODE_LEFT_BRACKET
  [72, "]"], // KEYCODE_RIGHT_BRACKET
  [73, "\\"], // KEYCODE_BACKSLASH
  [74, ";"], // KEYCODE_SEMICOLON
  [75, "'"], // KEYCODE_APOSTROPHE
  [76, "/"], // KEYCODE_SLASH
  [77, "@"], // KEYCODE_AT
  [78, "Num"], // KEYCODE_NUM
  [79, "Headset Hook"], // KEYCODE_HEADSETHOOK
  [80, "Focus"], // KEYCODE_FOCUS
  [81, "+"], // KEYCODE_PLUS
  [82, "Menu"], // KEYCODE_MENU
  [83, "Notification"], // KEYCODE_NOTIFICATION
  [84, "Search"], // KEYCODE_SEARCH
  [85, "Media Play/Pause"], // KEYCODE_MEDIA_PLAY_PAUSE
  [86, "Media Stop"], // KEYCODE_MEDIA_STOP
  [87, "Media Next"], // KEYCODE_MEDIA_NEXT
  [88, "Media Previous"], // KEYCODE_MEDIA_PREVIOUS
  [89, "Media Rewind"], // KEYCODE_MEDIA_REWIND
  [90, "Media Fast Forward"], // KEYCODE_MEDIA_FAST_FORWARD
  [91, "Mute"], // KEYCODE_MUTE
  [92, "Page Up"], // KEYCODE_PAGE_UP
  [93, "Page Down"], // KEYCODE_PAGE_DOWN
  [94, "PictSymbols"], // KEYCODE_PICTSYMBOLS
  [95, "Switch Charset"], // KEYCODE_SWITCH_CHARSET
  [96, "Button A"], // KEYCODE_BUTTON_A
  [97, "Button B"], // KEYCODE_BUTTON_B
  [98, "Button C"], // KEYCODE_BUTTON_C
  [99, "Button X"], // KEYCODE_BUTTON_X
  [100, "Button Y"], // KEYCODE_BUTTON_Y
  [101, "Button Z"], // KEYCODE_BUTTON_Z
  [102, "Button L1"], // KEYCODE_BUTTON_L1
  [103, "Button R1"], // KEYCODE_BUTTON_R1
  [104, "Button L2"], // KEYCODE_BUTTON_L2
  [105, "Button R2"], // KEYCODE_BUTTON_R2
  [106, "Button Thumb L"], // KEYCODE_BUTTON_THUMBL
  [107, "Button Thumb R"], // KEYCODE_BUTTON_THUMBR
  [108, "Button Start"], // KEYCODE_BUTTON_START
  [109, "Button Select"], // KEYCODE_BUTTON_SELECT
  [110, "Button Mode"], // KEYCODE_BUTTON_MODE
  [111, "Esc"], // KEYCODE_ESCAPE
  [112, "Left Ctrl"], // KEYCODE_CTRL_LEFT
  [113, "Right Ctrl"], // KEYCODE_CTRL_RIGHT
  [114, "Caps Lock"], // KEYCODE_CAPS_LOCK
  [115, "Scroll Lock"], // KEYCODE_SCROLL_LOCK
  [116, "Left Meta"], // KEYCODE_META_LEFT
  [117, "Right Meta"], // KEYCODE_META_RIGHT
  [118, "Function"], // KEYCODE_FUNCTION
  [119, "SysRq"], // KEYCODE_SYSRQ
  [120, "Break"], // KEYCODE_BREAK
  [121, "Move Home"], // KEYCODE_MOVE_HOME
  [122, "Move End"], // KEYCODE_MOVE_END
  [123, "Insert"], // KEYCODE_INSERT
  [124, "Forward"], // KEYCODE_FORWARD
  [125, "Media Play"], // KEYCODE_MEDIA_PLAY
  [126, "Media Pause"], // KEYCODE_MEDIA_PAUSE
  [127, "Media Close"], // KEYCODE_MEDIA_CLOSE
  [128, "Media Eject"], // KEYCODE_MEDIA_EJECT
  [129, "Media Record"], // KEYCODE_MEDIA_RECORD
  [130, "F1"], // KEYCODE_F1
  [131, "F2"], // KEYCODE_F2
  [132, "F3"], // KEYCODE_F3
  [133, "F4"], // KEYCODE_F4
  [134, "F5"], // KEYCODE_F5
  [135, "F6"], // KEYCODE_F6
  [136, "F7"], // KEYCODE_F7
  [137, "F8"], // KEYCODE_F8
  [138, "F9"], // KEYCODE_F9
  [139, "F10"], // KEYCODE_F10
  [140, "F11"], // KEYCODE_F11
  [141, "F12"], // KEYCODE_F12
  [142, "Num Lock"], // KEYCODE_NUM_LOCK
  [143, "Numpad 0"], // KEYCODE_NUMPAD_0
  [144, "Numpad 1"], // KEYCODE_NUMPAD_1
  [145, "Numpad 2"], // KEYCODE_NUMPAD_2
  [146, "Numpad 3"], // KEYCODE_NUMPAD_3
  [147, "Numpad 4"], // KEYCODE_NUMPAD_4
  [148, "Numpad 5"], // KEYCODE_NUMPAD_5
  [149, "Numpad 6"], // KEYCODE_NUMPAD_6
  [150, "Numpad 7"], // KEYCODE_NUMPAD_7
  [151, "Numpad 8"], // KEYCODE_NUMPAD_8
  [152, "Numpad 9"], // KEYCODE_NUMPAD_9
  [153, "Numpad /"], // KEYCODE_NUMPAD_DIVIDE
  [154, "Numpad *"], // KEYCODE_NUMPAD_MULTIPLY
  [155, "Numpad -"], // KEYCODE_NUMPAD_SUBTRACT
  [156, "Numpad +"], // KEYCODE_NUMPAD_ADD
  [157, "Numpad ."], // KEYCODE_NUMPAD_DOT
  [158, "Numpad ,"], // KEYCODE_NUMPAD_COMMA
  [159, "Numpad Enter"], // KEYCODE_NUMPAD_ENTER
  [160, "Numpad ="], // KEYCODE_NUMPAD_EQUALS
  [161, "Numpad ("], // KEYCODE_NUMPAD_LEFT_PAREN
  [162, "Numpad )"], // KEYCODE_NUMPAD_RIGHT_PAREN
  [163, "Volume Mute"], // KEYCODE_VOLUME_MUTE
  [164, "Info"], // KEYCODE_INFO
  [165, "Channel Up"], // KEYCODE_CHANNEL_UP
  [166, "Channel Down"], // KEYCODE_CHANNEL_DOWN
  [167, "Zoom In"], // KEYCODE_ZOOM_IN
  [168, "Zoom Out"], // KEYCODE_ZOOM_OUT
  [169, "TV"], // KEYCODE_TV
  [170, "Window"], // KEYCODE_WINDOW
  [171, "Guide"], // KEYCODE_GUIDE
  [172, "DVR"], // KEYCODE_DVR
  [173, "Bookmark"], // KEYCODE_BOOKMARK
  [174, "Captions"], // KEYCODE_CAPTIONS
  [175, "Settings"], // KEYCODE_SETTINGS
  [176, "TV Power"], // KEYCODE_TV_POWER
  [177, "TV Input"], // KEYCODE_TV_INPUT
  [178, "STB Power"], // KEYCODE_STB_POWER
  [179, "STB Input"], // KEYCODE_STB_INPUT
  [180, "AVR Power"], // KEYCODE_AVR_POWER
  [181, "AVR Input"], // KEYCODE_AVR_INPUT
  [182, "Program Red"], // KEYCODE_PROG_RED
  [183, "Program Green"], // KEYCODE_PROG_GREEN
  [184, "Program Yellow"], // KEYCODE_PROG_YELLOW
  [185, "Program Blue"], // KEYCODE_PROG_BLUE
  [186, "App Switch"], // KEYCODE_APP_SWITCH
  [187, "Button 1"], // KEYCODE_BUTTON_1
  [188, "Button 2"], // KEYCODE_BUTTON_2
  [189, "Button 3"], // KEYCODE_BUTTON_3
  [190, "Button 4"], // KEYCODE_BUTTON_4
  [191, "Button 5"], // KEYCODE_BUTTON_5
  [192, "Button 6"], // KEYCODE_BUTTON_6
  [193, "Button 7"], // KEYCODE_BUTTON_7
  [194, "Button 8"], // KEYCODE_BUTTON_8
  [195, "Button 9"], // KEYCODE_BUTTON_9
  [196, "Button 10"], // KEYCODE_BUTTON_10
  [197, "Button 11"], // KEYCODE_BUTTON_11
  [198, "Button 12"], // KEYCODE_BUTTON_12
  [199, "Button 13"], // KEYCODE_BUTTON_13
  [200, "Button 14"], // KEYCODE_BUTTON_14
  [201, "Button 15"], // KEYCODE_BUTTON_15
  [202, "Button 16"], // KEYCODE_BUTTON_16
  [203, "Language Switch"], // KEYCODE_LANGUAGE_SWITCH
  [204, "Manner Mode"], // KEYCODE_MANNER_MODE
  [205, "3D Mode"], // KEYCODE_3D_MODE
  [206, "Contacts"], // KEYCODE_CONTACTS
  [207, "Calendar"], // KEYCODE_CALENDAR
  [208, "Music"], // KEYCODE_MUSIC
  [209, "Calculator"], // KEYCODE_CALCULATOR
  [210, "Zenkaku/Hankaku"], // KEYCODE_ZENKAKU_HANKAKU
  [211, "Eisu"], // KEYCODE_EISU
  [212, "Muhenkan"], // KEYCODE_MUHENKAN
  [213, "Henkan"], // KEYCODE_HENKAN
  [214, "Katakana/Hiragana"], // KEYCODE_KATAKANA_HIRAGANA
  [215, "Yen"], // KEYCODE_YEN
  [216, "Ro"], // KEYCODE_RO
  [217, "Kana"], // KEYCODE_KANA
  [218, "Assist"], // KEYCODE_ASSIST
  [219, "Brightness Down"], // KEYCODE_BRIGHTNESS_DOWN
  [220, "Brightness Up"], // KEYCODE_BRIGHTNESS_UP
  [221, "Media Audio Track"], // KEYCODE_MEDIA_AUDIO_TRACK
  [222, "Sleep"], // KEYCODE_SLEEP
  [223, "Wakeup"], // KEYCODE_WAKEUP
  [224, "Pairing"], // KEYCODE_PAIRING
  [225, "Media Top Menu"], // KEYCODE_MEDIA_TOP_MENU
  [226, "11"], // KEYCODE_11
  [227, "12"], // KEYCODE_12
  [228, "Last Channel"], // KEYCODE_LAST_CHANNEL
  [229, "TV Data Service"], // KEYCODE_TV_DATA_SERVICE
  [230, "Voice Assist"], // KEYCODE_VOICE_ASSIST
  [231, "TV Radio Service"], // KEYCODE_TV_RADIO_SERVICE
  [232, "TV Teletext"], // KEYCODE_TV_TELETEXT
  [233, "TV Number Entry"], // KEYCODE_TV_NUMBER_ENTRY
  [234, "TV Terrestrial Analog"], // KEYCODE_TV_TERRESTRIAL_ANALOG
  [235, "TV Terrestrial Digital"], // KEYCODE_TV_TERRESTRIAL_DIGITAL
  [236, "TV Satellite"], // KEYCODE_TV_SATELLITE
  [237, "TV Satellite BS"], // KEYCODE_TV_SATELLITE_BS
  [238, "TV Satellite CS"], // KEYCODE_TV_SATELLITE_CS
  [239, "TV Satellite Service"], // KEYCODE_TV_SATELLITE_SERVICE
  [240, "TV Network"], // KEYCODE_TV_NETWORK
  [241, "TV Antenna/Cable"], // KEYCODE_TV_ANTENNA_CABLE
  [242, "TV HDMI 1"], // KEYCODE_TV_INPUT_HDMI_1
  [243, "TV HDMI 2"], // KEYCODE_TV_INPUT_HDMI_2
  [244, "TV HDMI 3"], // KEYCODE_TV_INPUT_HDMI_3
  [245, "TV HDMI 4"], // KEYCODE_TV_INPUT_HDMI_4
  [246, "TV Composite 1"], // KEYCODE_TV_INPUT_COMPOSITE_1
  [247, "TV Composite 2"], // KEYCODE_TV_INPUT_COMPOSITE_2
  [248, "TV Component 1"], // KEYCODE_TV_INPUT_COMPONENT_1
  [249, "TV Component 2"], // KEYCODE_TV_INPUT_COMPONENT_2
  [250, "TV VGA 1"], // KEYCODE_TV_INPUT_VGA_1
  [251, "Audio Description"], // KEYCODE_TV_AUDIO_DESCRIPTION
  [252, "Audio Description Mix Up"], // KEYCODE_TV_AUDIO_DESCRIPTION_MIX_UP
  [253, "Audio Description Mix Down"], // KEYCODE_TV_AUDIO_DESCRIPTION_MIX_DOWN
  [254, "Zoom Mode"], // KEYCODE_TV_ZOOM_MODE
  [255, "Contents Menu"], // KEYCODE_TV_CONTENTS_MENU
  [256, "Media Context Menu"], // KEYCODE_TV_MEDIA_CONTEXT_MENU
  [257, "Timer Programming"], // KEYCODE_TV_TIMER_PROGRAMMING
  [258, "Help"], // KEYCODE_HELP
  [259, "Navigate Previous"], // KEYCODE_NAVIGATE_PREVIOUS
  [260, "Navigate Next"], // KEYCODE_NAVIGATE_NEXT
  [261, "Navigate In"], // KEYCODE_NAVIGATE_IN
  [262, "Navigate Out"], // KEYCODE_NAVIGATE_OUT
  [263, "Stem Primary"], // KEYCODE_STEM_PRIMARY
  [264, "Stem 1"], // KEYCODE_STEM_1
  [265, "Stem 2"], // KEYCODE_STEM_2
  [266, "Stem 3"], // KEYCODE_STEM_3
  [267, "DPad Up Left"], // KEYCODE_DPAD_UP_LEFT
  [268, "DPad Down Left"], // KEYCODE_DPAD_DOWN_LEFT
  [269, "DPad Up Right"], // KEYCODE_DPAD_UP_RIGHT
  [270, "DPad Down Right"], // KEYCODE_DPAD_DOWN_RIGHT
  [271, "Media Skip Forward"], // KEYCODE_MEDIA_SKIP_FORWARD
  [272, "Media Skip Backward"], // KEYCODE_MEDIA_SKIP_BACKWARD
  [273, "Media Step Forward"], // KEYCODE_MEDIA_STEP_FORWARD
  [274, "Media Step Backward"], // KEYCODE_MEDIA_STEP_BACKWARD
  [275, "Soft Sleep"], // KEYCODE_SOFT_SLEEP
  [276, "Cut"], // KEYCODE_CUT
  [277, "Copy"], // KEYCODE_COPY
  [278, "Paste"], // KEYCODE_PASTE
  [279, "System Navigation Up"], // KEYCODE_SYSTEM_NAVIGATION_UP
  [280, "System Navigation Down"], // KEYCODE_SYSTEM_NAVIGATION_DOWN
  [281, "System Navigation Left"], // KEYCODE_SYSTEM_NAVIGATION_LEFT
  [282, "System Navigation Right"], // KEYCODE_SYSTEM_NAVIGATION_RIGHT
  [283, "All Apps"], // KEYCODE_ALL_APPS
  [284, "Refresh"], // KEYCODE_REFRESH
  [285, "Thumbs Up"], // KEYCODE_THUMBS_UP
  [286, "Thumbs Down"], // KEYCODE_THUMBS_DOWN
  [287, "Profile Switch"], // KEYCODE_PROFILE_SWITCH
]);

const gamepad: Map<number, string> = new Map([
  // Xbox 360 buttons (XUSB)
  [0x1000, "A / Cross"],
  [0x2000, "B / Circle"],
  [0x4000, "X / Square"],
  [0x8000, "Y / Triangle"],
  [0x0100, "LB / L1"],
  [0x0200, "RB / R1"],
  [0x0040, "Left Thumb / L3"],
  [0x0080, "Right Thumb / R3"],
  [0x0010, "Start / Options"],
  [0x0020, "Back / Share"],
  [0x0400, "Guide"],
  [0x0001, "DPad Up"],
  [0x0002, "DPad Down"],
  [0x0004, "DPad Left"],
  [0x0008, "DPad Right"],

  // DualShock 4 special buttons
  [0x10000, "PS"],
  [0x20000, "Touchpad"],
]);
