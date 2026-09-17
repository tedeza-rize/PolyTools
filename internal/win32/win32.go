//go:build windows

package win32

import "golang.org/x/sys/windows"

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")

	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
	procGetForegroundWindow     = user32.NewProc("GetForegroundWindow")
	procGetWindowLong           = user32.NewProc("GetWindowLongW")
	procSetWindowPos            = user32.NewProc("SetWindowPos")
)

// SetThreadExecutionState flags
const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

// SetWindowPos / styles
const (
	wsExTopMost   = 0x00000008
	swpNoMove     = 0x0002
	swpNoSize     = 0x0001
	swpNoActivate = 0x0010
)

// GWL_EXSTYLE is -20; expressed as unsigned 32-bit so it can be passed as a
// uintptr argument (the callee reads the low 32 bits as int).
const gwlExStyle = 0xFFFFFFEC

var (
	hwndTopMost   = ^uintptr(0)     // (HWND)-1
	hwndNoTopMost = ^uintptr(0) - 1 // (HWND)-2
)

// KeepAwake prevents sleep (and optionally display sleep) until ClearAwake.
func KeepAwake(keepDisplayOn bool) {
	flags := uintptr(esContinuous | esSystemRequired)
	if keepDisplayOn {
		flags |= esDisplayRequired
	}
	procSetThreadExecutionState.Call(flags)
}

// ClearAwake lets the system sleep normally again.
func ClearAwake() {
	procSetThreadExecutionState.Call(esContinuous)
}

// ForegroundWindow returns the HWND of the window the user is working with.
func ForegroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
}

// IsTopMost reports whether hwnd has WS_EX_TOPMOST.
func IsTopMost(hwnd uintptr) bool {
	style, _, _ := procGetWindowLong.Call(hwnd, uintptr(gwlExStyle))
	return style&wsExTopMost != 0
}

// SetTopMost toggles WS_EX_TOPMOST on hwnd without moving/resizing it.
func SetTopMost(hwnd uintptr, top bool) {
	insert := hwndNoTopMost
	if top {
		insert = hwndTopMost
	}
	procSetWindowPos.Call(hwnd, insert, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpNoActivate)
}
