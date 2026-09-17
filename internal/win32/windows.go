//go:build windows

package win32

import (
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetWindowText          = user32.NewProc("GetWindowTextW")
	procGetWindowTextLength    = user32.NewProc("GetWindowTextLengthW")
	procGetWindowThreadProcess = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible        = user32.NewProc("IsWindowVisible")
	procIsWindow               = user32.NewProc("IsWindow")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procMoveWindow             = user32.NewProc("MoveWindow")
	procShowWindow             = user32.NewProc("ShowWindow")
	procGetAncestor            = user32.NewProc("GetAncestor")
	procWindowFromPoint        = user32.NewProc("WindowFromPoint")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procBringWindowToTop       = user32.NewProc("BringWindowToTop")
	procEnumWindows            = user32.NewProc("EnumWindows")
	procMonitorFromWindow      = user32.NewProc("MonitorFromWindow")
	procMonitorFromPoint       = user32.NewProc("MonitorFromPoint")
	procGetMonitorInfo         = user32.NewProc("GetMonitorInfoW")
	procGetWindowLongS         = user32.NewProc("GetWindowLongW")
	procSetWindowLong          = user32.NewProc("SetWindowLongW")
	procGetWindowPlacement     = user32.NewProc("GetWindowPlacement")
	procSetWindowPlacement     = user32.NewProc("SetWindowPlacement")
	procOpenProcess            = kernel32.NewProc("OpenProcess")
	procQueryFullProcessImage  = kernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle            = kernel32.NewProc("CloseHandle")
)

const (
	gwlStyle    = 0xFFFFFFF0 // -16
	gaRoot      = 2
	gaRootOwner = 3

	swHide        = 0
	swShow        = 5
	swRestore     = 9
	swShowNA      = 8

	monitorDefaultToNearest = 2

	wsCaption     = 0x00C00000
	wsThickFrame  = 0x00040000
	wsMinimizeBox = 0x00020000
	wsMaximizeBox = 0x00010000
	wsSysMenu     = 0x00080000
	wsOverlapped  = wsCaption | wsThickFrame | wsMinimizeBox | wsMaximizeBox | wsSysMenu

	swpNoZOrder    = 0x0004
	swpFrameChanged = 0x0020
	swpShowWindow  = 0x0040

	processQueryLimitedInfo = 0x1000
)

type Rect struct {
	Left, Top, Right, Bottom int32
}

func (r Rect) Width() int32  { return r.Right - r.Left }
func (r Rect) Height() int32 { return r.Bottom - r.Top }

// WindowTitle returns the title bar text of hwnd.
func WindowTitle(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLength.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return windows.UTF16ToString(buf)
}

// WindowExe returns the process image name (e.g. "notepad.exe") owning hwnd.
func WindowExe(hwnd uintptr) string {
	var pid uint32
	procGetWindowThreadProcess.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}
	h, _, _ := procOpenProcess.Call(processQueryLimitedInfo, 0, uintptr(pid))
	if h == 0 {
		return ""
	}
	defer procCloseHandle.Call(h)
	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	r, _, _ := procQueryFullProcessImage.Call(
		h, 0, uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r == 0 {
		return ""
	}
	return strings.ToLower(filepath.Base(windows.UTF16ToString(buf[:size])))
}

// WindowRect returns the window's bounding rectangle.
func WindowRect(hwnd uintptr) (Rect, bool) {
	var r Rect
	ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	return r, ret != 0
}

// MoveWindowTo repositions and resizes a window.
func MoveWindowTo(hwnd uintptr, r Rect) {
	procMoveWindow.Call(hwnd,
		uintptr(r.Left), uintptr(r.Top),
		uintptr(r.Width()), uintptr(r.Height()), 1)
}

// SetWindowPosTo repositions/resizes without changing Z-order.
func SetWindowPosTo(hwnd uintptr, r Rect) {
	procSetWindowPos.Call(hwnd, 0,
		uintptr(r.Left), uintptr(r.Top),
		uintptr(r.Width()), uintptr(r.Height()),
		swpNoZOrder|swpNoActivate|swpFrameChanged)
}

// Show / Hide / Restore wrap ShowWindow.
func HideWindow(hwnd uintptr)    { procShowWindow.Call(hwnd, swHide) }
func ShowWindowS(hwnd uintptr)   { procShowWindow.Call(hwnd, swShow) }
func RestoreWindow(hwnd uintptr) { procShowWindow.Call(hwnd, swRestore) }

func IsWindowAlive(hwnd uintptr) bool {
	r, _, _ := procIsWindow.Call(hwnd)
	return r != 0
}

// RootWindow returns the top-level ancestor of hwnd.
func RootWindow(hwnd uintptr) uintptr {
	r, _, _ := procGetAncestor.Call(hwnd, gaRoot)
	if r == 0 {
		return hwnd
	}
	return r
}

// WindowAtPoint returns the top-level window under a screen point.
func WindowAtPoint(x, y int32) uintptr {
	pt := uintptr(uint64(uint32(y))<<32 | uint64(uint32(x)))
	w, _, _ := procWindowFromPoint.Call(pt)
	if w == 0 {
		return 0
	}
	return RootWindow(w)
}

// SetForeground brings hwnd to the front (best effort).
func SetForeground(hwnd uintptr) {
	procSetForegroundWindow.Call(hwnd)
	procBringWindowToTop.Call(hwnd)
}

// WindowStyle returns GWL_STYLE bits.
func WindowStyle(hwnd uintptr) uintptr {
	s, _, _ := procGetWindowLongS.Call(hwnd, gwlStyle)
	return s
}

// SetWindowStyle replaces GWL_STYLE and redraws the frame.
func SetWindowStyle(hwnd uintptr, style uintptr) {
	procSetWindowLong.Call(hwnd, gwlStyle, style)
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged)
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor Rect
	RcWork    Rect
	DwFlags   uint32
}

// MonitorRect returns the full monitor rect containing hwnd.
func MonitorRect(hwnd uintptr) Rect {
	m, _, _ := procMonitorFromWindow.Call(hwnd, monitorDefaultToNearest)
	var mi monitorInfo
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfo.Call(m, uintptr(unsafe.Pointer(&mi)))
	return mi.RcMonitor
}

// MonitorRectAt returns the monitor rect containing a screen point.
func MonitorRectAt(x, y int32) Rect {
	pt := uintptr(uint64(uint32(y))<<32 | uint64(uint32(x)))
	m, _, _ := procMonitorFromPoint.Call(pt, monitorDefaultToNearest)
	var mi monitorInfo
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfo.Call(m, uintptr(unsafe.Pointer(&mi)))
	return mi.RcMonitor
}

// EnumTopWindows calls cb for each visible top-level window.
// Returning false stops enumeration.
func EnumTopWindows(cb func(hwnd uintptr) bool) {
	wndProc := windows.NewCallback(func(hwnd, _ uintptr) uintptr {
		v, _, _ := procIsWindowVisible.Call(hwnd)
		if v == 0 || WindowTitle(hwnd) == "" {
			return 1
		}
		if cb(hwnd) {
			return 1
		}
		return 0
	})
	procEnumWindows.Call(wndProc, 0)
}

// WindowPlacement captures show state + restore rect for later use.
type windowPlacement struct {
	Length           uint32
	Flags            uint32
	ShowCmd          uint32
	PtMinPosition    struct{ X, Y int32 }
	PtMaxPosition    struct{ X, Y int32 }
	RcNormalPosition Rect
}

// PlacementRect returns the window's normal (restored) rect — unlike
// WindowRect this is meaningful for minimized/maximized windows.
func PlacementRect(hwnd uintptr) (Rect, bool) {
	var wp windowPlacement
	wp.Length = uint32(unsafe.Sizeof(wp))
	r, _, _ := procGetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&wp)))
	return wp.RcNormalPosition, r != 0
}
