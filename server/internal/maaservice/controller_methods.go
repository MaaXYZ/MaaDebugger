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

func toUint64String(v uint64) string {
	return strconv.FormatUint(v, 10)
}

var adbScreencapMethods = []ControllerMethod{
	{Label: adb.ScreencapDefault.String(), Value: toUint64String(uint64(adb.ScreencapDefault))},
	{Label: adb.ScreencapEncodeToFileAndPull.String(), Value: toUint64String(uint64(adb.ScreencapEncodeToFileAndPull))},
	{Label: adb.ScreencapEncode.String(), Value: toUint64String(uint64(adb.ScreencapEncode))},
	{Label: adb.ScreencapRawWithGzip.String(), Value: toUint64String(uint64(adb.ScreencapRawWithGzip))},
	{Label: adb.ScreencapRawByNetcat.String(), Value: toUint64String(uint64(adb.ScreencapRawByNetcat))},
	{Label: adb.ScreencapMinicapDirect.String(), Value: toUint64String(uint64(adb.ScreencapMinicapDirect))},
	{Label: adb.ScreencapMinicapStream.String(), Value: toUint64String(uint64(adb.ScreencapMinicapStream))},
	{Label: adb.ScreencapEmulatorExtras.String(), Value: toUint64String(uint64(adb.ScreencapEmulatorExtras))},
	{Label: adb.ScreencapAll.String(), Value: toUint64String(uint64(adb.ScreencapAll))},
}

var adbInputMethods = []ControllerMethod{
	{Label: adb.InputDefault.String(), Value: toUint64String(uint64(adb.InputDefault))},
	{Label: adb.InputAdbShell.String(), Value: toUint64String(uint64(adb.InputAdbShell))},
	{Label: adb.InputMinitouchAndAdbKey.String(), Value: toUint64String(uint64(adb.InputMinitouchAndAdbKey))},
	{Label: adb.InputMaatouch.String(), Value: toUint64String(uint64(adb.InputMaatouch))},
	{Label: adb.InputEmulatorExtras.String(), Value: toUint64String(uint64(adb.InputEmulatorExtras))},
	{Label: adb.InputAll.String(), Value: toUint64String(uint64(adb.InputAll))},
}

var windowScreencapMethods = []ControllerMethod{
	{Label: win32.ScreencapGDI.String(), Value: toUint64String(uint64(win32.ScreencapGDI))},
	{Label: win32.ScreencapFramePool.String(), Value: toUint64String(uint64(win32.ScreencapFramePool))},
	{Label: win32.ScreencapDXGIDesktopDup.String(), Value: toUint64String(uint64(win32.ScreencapDXGIDesktopDup))},
	{Label: win32.ScreencapDXGIDesktopDupWindow.String(), Value: toUint64String(uint64(win32.ScreencapDXGIDesktopDupWindow))},
	{Label: win32.ScreencapPrintWindow.String(), Value: toUint64String(uint64(win32.ScreencapPrintWindow))},
	{Label: win32.ScreencapScreenDC.String(), Value: toUint64String(uint64(win32.ScreencapScreenDC))},
}

var win32InputMethods = []ControllerMethod{
	{Label: win32.InputSeize.String(), Value: toUint64String(uint64(win32.InputSeize))},
	{Label: win32.InputSendMessage.String(), Value: toUint64String(uint64(win32.InputSendMessage))},
	{Label: win32.InputPostMessage.String(), Value: toUint64String(uint64(win32.InputPostMessage))},
	{Label: win32.InputLegacyEvent.String(), Value: toUint64String(uint64(win32.InputLegacyEvent))},
	{Label: win32.InputPostThreadMessage.String(), Value: toUint64String(uint64(win32.InputPostThreadMessage))},
	{Label: win32.InputSendMessageWithCursorPos.String(), Value: toUint64String(uint64(win32.InputSendMessageWithCursorPos))},
	{Label: win32.InputPostMessageWithCursorPos.String(), Value: toUint64String(uint64(win32.InputPostMessageWithCursorPos))},
	{Label: win32.InputSendMessageWithCursorPos.String(), Value: toUint64String(uint64(win32.InputSendMessageWithCursorPos))},
	{Label: win32.InputPostMessageWithCursorPos.String(), Value: toUint64String(uint64(win32.InputPostMessageWithCursorPos))},
}

var gamepadInputMethods = []ControllerMethod{
	{Label: "Xbox 360", Value: toUint64String(uint64(maa.GamepadTypeXbox360)), Icon: "i-simple-icons:xbox"},
	{Label: "DualShock 4", Value: toUint64String(uint64(maa.GamepadTypeDualShock4)), Icon: "i-simple-icons:playstation"},
}
