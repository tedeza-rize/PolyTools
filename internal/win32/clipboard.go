//go:build windows

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procIsClipboardFmt   = user32.NewProc("IsClipboardFormatAvailable")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procDragQueryFile    = windows.NewLazySystemDLL("shell32.dll").NewProc("DragQueryFileW")
)

const (
	cfUnicodeText = 13
	cfHDrop       = 15
	gmemMoveable  = 0x0002
)

func openClipboard() bool {
	for i := 0; i < 10; i++ {
		r, _, _ := procOpenClipboard.Call(0)
		if r != 0 {
			return true
		}
	}
	return false
}

// ClipboardText reads CF_UNICODETEXT.
func ClipboardText() string {
	if !openClipboard() {
		return ""
	}
	defer procCloseClipboard.Call()
	h, _, _ := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return ""
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return ""
	}
	defer procGlobalUnlock.Call(h)
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(p)))
}

// SetClipboardText writes CF_UNICODETEXT (clipboard is emptied first).
func SetClipboardText(s string) bool {
	if !openClipboard() {
		return false
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()

	u16, err := windows.UTF16FromString(s)
	if err != nil {
		return false
	}
	size := uintptr(len(u16) * 2)
	h, _, _ := procGlobalAlloc.Call(gmemMoveable, size)
	if h == 0 {
		return false
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return false
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u16))
	copy(dst, u16)
	procGlobalUnlock.Call(h)
	r, _, _ := procSetClipboardData.Call(cfUnicodeText, h)
	return r != 0
}

// ClipboardFiles returns the file list (CF_HDROP) if present.
func ClipboardFiles() []string {
	if !openClipboard() {
		return nil
	}
	defer procCloseClipboard.Call()
	ok, _, _ := procIsClipboardFmt.Call(cfHDrop)
	if ok == 0 {
		return nil
	}
	h, _, _ := procGetClipboardData.Call(cfHDrop)
	if h == 0 {
		return nil
	}
	var count uint32
	c, _, _ := procDragQueryFile.Call(h, 0xFFFFFFFF, 0, 0)
	count = uint32(c)
	var files []string
	for i := uint32(0); i < count; i++ {
		n, _, _ := procDragQueryFile.Call(h, uintptr(i), 0, 0)
		if n == 0 {
			continue
		}
		buf := make([]uint16, n+1)
		procDragQueryFile.Call(h, uintptr(i),
			uintptr(unsafe.Pointer(&buf[0])), n+1)
		files = append(files, windows.UTF16ToString(buf))
	}
	return files
}
