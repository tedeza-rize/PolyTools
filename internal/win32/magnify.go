//go:build windows

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	magnification = windows.NewLazySystemDLL("magnification.dll")

	procMagInitialize          = magnification.NewProc("MagInitialize")
	procMagUninitialize        = magnification.NewProc("MagUninitialize")
	procMagSetFullscreenTransf = magnification.NewProc("MagSetFullscreenTransform")
)

// MagInit initializes the magnification runtime for fullscreen transforms.
func MagInit() bool {
	r, _, _ := procMagInitialize.Call()
	return r != 0
}

// MagShutdown releases it.
func MagShutdown() {
	procMagUninitialize.Call()
}

// SetFullscreenScale magnifies the whole desktop by factor (e.g. 2.0).
// xOffset/yOffset choose which screen point lands at the top-left.
// factor <= 1 restores the unmagnified desktop.
func SetFullscreenScale(factor float32, xOffset, yOffset int32) bool {
	if factor <= 1.0 {
		r, _, _ := procMagSetFullscreenTransf.Call(0x3F800000, 0, 0) // 1.0f
		return r != 0
	}
	r, _, _ := procMagSetFullscreenTransf.Call(
		uintptr(*(*uint32)(unsafeFloat(&factor))),
		uintptr(xOffset), uintptr(yOffset),
	)
	return r != 0
}

func unsafeFloat(f *float32) *uint32 {
	return (*uint32)(unsafe.Pointer(f))
}
