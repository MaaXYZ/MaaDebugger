//go:build windows

package platform

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func UACEnabled() (*bool, error) {
	token := windows.GetCurrentProcessToken()

	var elevation uint32
	var outLen uint32
	if err := windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&outLen,
	); err != nil {
		return nil, err
	}

	enabled := elevation != 0
	return &enabled, nil
}
