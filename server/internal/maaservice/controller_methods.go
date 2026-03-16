package maaservice

import (
	"strconv"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/win32"
)

type MethodType string

const (
	ADBScreencap    MethodType = "adb_screencap"
	ADBInput        MethodType = "adb_input"
	WindowScreencap MethodType = "window_screencap"
	Win32Input      MethodType = "win32_input"
	GamepadInput    MethodType = "gamepad_type"
)

type ControllerMethod struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Icon  string `json:"icon"`
}

func toUint64String[T ~uint64](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}

func method[T interface {
	~uint64
	String() string
}](v T) ControllerMethod {
	return ControllerMethod{
		Label: v.String(),
		Value: toUint64String(v),
	}
}

func methodWithIcon[T interface {
	~uint64
}](label string, v T, icon string) ControllerMethod {
	return ControllerMethod{
		Label: label,
		Value: toUint64String(v),
		Icon:  icon,
	}
}

var (
	ADBScreencapDefault                 = toUint64String(adb.ScreencapDefault)
	ADBScreenCapEncodeToFileAndPull     = toUint64String(adb.ScreencapEncodeToFileAndPull)
	ADBScreencapEncode                  = toUint64String(adb.ScreencapEncode)
	ADBScreencapRawWithGzip             = toUint64String(adb.ScreencapRawWithGzip)
	ADBScreencapRawByNetcat             = toUint64String(adb.ScreencapRawByNetcat)
	ADBScreencapMinicapDirect           = toUint64String(adb.ScreencapMinicapDirect)
	ADBScreencapMinicapStream           = toUint64String(adb.ScreencapMinicapStream)
	ADBScreencapEmulatorExtras          = toUint64String(adb.ScreencapEmulatorExtras)
	ADBScreencapAll                     = toUint64String(adb.ScreencapAll)
	ADBInputDefault                     = toUint64String(adb.InputDefault)
	ADBInputAdbShell                    = toUint64String(adb.InputAdbShell)
	ADBInputMinitouchAndAdbKey          = toUint64String(adb.InputMinitouchAndAdbKey)
	ADBInputMaatouch                    = toUint64String(adb.InputMaatouch)
	ADBInputEmulatorExtras              = toUint64String(adb.InputEmulatorExtras)
	ADBInputAll                         = toUint64String(adb.InputAll)
	WindowScreencapGDI                  = toUint64String(win32.ScreencapGDI)
	WindowScreencapFramePool            = toUint64String(win32.ScreencapFramePool)
	WindowScreencapDXGIDesktopDup       = toUint64String(win32.ScreencapDXGIDesktopDup)
	WindowScreencapDXGIDesktopDupWindow = toUint64String(win32.ScreencapDXGIDesktopDupWindow)
	WindowScreencapPrintWindow          = toUint64String(win32.ScreencapPrintWindow)
	WindowScreencapScreenDC             = toUint64String(win32.ScreencapScreenDC)
	Win32InputSeize                     = toUint64String(win32.InputSeize)
	Win32InputSendMessage               = toUint64String(win32.InputSendMessage)
	Win32InputPostMessage               = toUint64String(win32.InputPostMessage)
	Win32InputLegacyEvent               = toUint64String(win32.InputLegacyEvent)
	Win32InputPostThreadMessage         = toUint64String(win32.InputPostThreadMessage)
	Win32InputSendMessageWithCursorPos  = toUint64String(win32.InputSendMessageWithCursorPos)
	Win32InputPostMessageWithCursorPos  = toUint64String(win32.InputPostMessageWithCursorPos)
	GamepadTypeXbox360                  = toUint64String(maa.GamepadTypeXbox360)
	GamepadTypeDualShock4               = toUint64String(maa.GamepadTypeDualShock4)
)

var adbScreencapMethods = []ControllerMethod{
	method(adb.ScreencapDefault),
	method(adb.ScreencapEncodeToFileAndPull),
	method(adb.ScreencapEncode),
	method(adb.ScreencapRawWithGzip),
	method(adb.ScreencapRawByNetcat),
	method(adb.ScreencapMinicapDirect),
	method(adb.ScreencapMinicapStream),
	method(adb.ScreencapEmulatorExtras),
	method(adb.ScreencapAll),
}

var adbInputMethods = []ControllerMethod{
	method(adb.InputDefault),
	method(adb.InputAdbShell),
	method(adb.InputMinitouchAndAdbKey),
	method(adb.InputMaatouch),
	method(adb.InputEmulatorExtras),
	method(adb.InputAll),
}

var windowScreencapMethods = []ControllerMethod{
	method(win32.ScreencapGDI),
	method(win32.ScreencapFramePool),
	method(win32.ScreencapDXGIDesktopDup),
	method(win32.ScreencapDXGIDesktopDupWindow),
	method(win32.ScreencapPrintWindow),
	method(win32.ScreencapScreenDC),
}

var win32InputMethods = []ControllerMethod{
	method(win32.InputSeize),
	method(win32.InputSendMessage),
	method(win32.InputPostMessage),
	method(win32.InputLegacyEvent),
	method(win32.InputPostThreadMessage),
	method(win32.InputSendMessageWithCursorPos),
	method(win32.InputPostMessageWithCursorPos),
}

var gamepadInputMethods = []ControllerMethod{
	methodWithIcon("Xbox 360", maa.GamepadTypeXbox360, "i-simple-icons:xbox"),
	methodWithIcon("DualShock 4", maa.GamepadTypeDualShock4, "i-simple-icons:playstation"),
}
