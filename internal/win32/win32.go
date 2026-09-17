//go:build windows

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")

	procSetThreadExecutionState     = kernel32.NewProc("SetThreadExecutionState")
	procGetSystemTimes              = kernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx        = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64              = kernel32.NewProc("GetTickCount64")
	procGetSystemPowerStatus        = kernel32.NewProc("GetSystemPowerStatus")
	procGetForegroundWindow         = user32.NewProc("GetForegroundWindow")
	procGetWindowLong               = user32.NewProc("GetWindowLongW")
	procGetWindowLongPtr            = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtr            = user32.NewProc("SetWindowLongPtrW")
	procSetLayeredWindowAttributes  = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowPos                = user32.NewProc("SetWindowPos")
	procGetCursorPos                = user32.NewProc("GetCursorPos")
	procGetDC                       = user32.NewProc("GetDC")
	procReleaseDC                   = user32.NewProc("ReleaseDC")
	procEnumDisplayDevices          = user32.NewProc("EnumDisplayDevicesW")
	procMessageBeep                 = user32.NewProc("MessageBeep")
	procGetPixel                    = gdi32.NewProc("GetPixel")
)

// SetThreadExecutionState flags
const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

// SetWindowPos / styles
const (
	wsExTopMost = 0x00000008
	wsExLayered = 0x00080000
	swpNoMove   = 0x0002
	swpNoSize   = 0x0001
)

const swpNoActivate = 0x0010

// SetLayeredWindowAttributes flags
const lwaAlpha = 0x00000002

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

// ExStyle returns the extended window style bits of hwnd.
func ExStyle(hwnd uintptr) uintptr {
	style, _, _ := procGetWindowLongPtr.Call(hwnd, uintptr(gwlExStyle))
	return style
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

// IsLayered reports whether hwnd already has WS_EX_LAYERED.
func IsLayered(hwnd uintptr) bool {
	return ExStyle(hwnd)&wsExLayered != 0
}

// SetTransparency makes hwnd percent-opaque (0–100). Adds WS_EX_LAYERED if
// the window does not have it; the caller tracks whether it was added so
// ClearTransparency can remove it only then.
func SetTransparency(hwnd uintptr, percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	ex := ExStyle(hwnd)
	if ex&wsExLayered == 0 {
		procSetWindowLongPtr.Call(hwnd, uintptr(gwlExStyle), ex|wsExLayered)
	}
	alpha := uintptr(255 * percent / 100)
	procSetLayeredWindowAttributes.Call(hwnd, 0, alpha, lwaAlpha)
}

// ClearTransparency restores full opacity; removes WS_EX_LAYERED when
// removeLayered is true (i.e. we added it ourselves).
func ClearTransparency(hwnd uintptr, removeLayered bool) {
	procSetLayeredWindowAttributes.Call(hwnd, 0, 255, lwaAlpha)
	if removeLayered {
		ex := ExStyle(hwnd)
		procSetWindowLongPtr.Call(hwnd, uintptr(gwlExStyle), ex&^wsExLayered)
	}
}

// CursorPos returns the current cursor position in screen coordinates.
func CursorPos() (x, y int32) {
	var pt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return pt.X, pt.Y
}

// PixelColor samples the color of the pixel at (x, y) in screen coordinates.
func PixelColor(x, y int32) (r, g, b uint8, ok bool) {
	dc, _, _ := procGetDC.Call(0)
	if dc == 0 {
		return 0, 0, 0, false
	}
	defer procReleaseDC.Call(0, dc)
	c, _, _ := procGetPixel.Call(dc, uintptr(x), uintptr(y))
	if c == 0xFFFFFFFF {
		return 0, 0, 0, false
	}
	// COLORREF is 0x00BBGGRR.
	return uint8(c), uint8(c >> 8), uint8(c >> 16), true
}

// Beep plays the default system sound — cheap audible feedback.
func Beep() {
	procMessageBeep.Call(0)
}

// --- system information ---

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type systemPowerStatus struct {
	ACLineStatus        uint8
	BatteryFlag         uint8
	BatteryLifePercent  uint8
	SystemStatusFlag    uint8
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// SystemTimes samples idle/kernel/user times (in 100ns units).
func SystemTimes() (idle, kernel, user uint64) {
	var i, k, u filetime
	procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Pointer(&k)),
		uintptr(unsafe.Pointer(&u)),
	)
	ft := func(f filetime) uint64 { return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime) }
	return ft(i), ft(k), ft(u)
}

// MemoryStatus returns total/available physical memory in bytes plus the
// 0–100 memory load percentage.
func MemoryStatus() (total, avail uint64, loadPercent uint32) {
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m)))
	return m.TotalPhys, m.AvailPhys, m.MemoryLoad
}

// UptimeSeconds returns ms tick count converted to seconds.
func UptimeSeconds() uint64 {
	ms, _, _ := procGetTickCount64.Call()
	return uint64(ms) / 1000
}

// PowerStatus returns AC line state and battery info.
// hasBattery is false on desktops; percent is -1 when unknown.
func PowerStatus() (onAC bool, percent int, hasBattery bool) {
	var s systemPowerStatus
	procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&s)))
	if s.BatteryFlag&0x80 != 0 || s.BatteryLifePercent == 255 {
		return s.ACLineStatus == 1, -1, false
	}
	return s.ACLineStatus == 1, int(s.BatteryLifePercent), true
}

type displayDevice struct {
	Cb           uint32
	DeviceName   [32]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

// PrimaryGPUName returns the display name of the first display adapter.
func PrimaryGPUName() string {
	var d displayDevice
	d.Cb = uint32(unsafe.Sizeof(d))
	r, _, _ := procEnumDisplayDevices.Call(0, 0, uintptr(unsafe.Pointer(&d)), 0)
	if r == 0 {
		return ""
	}
	return windows.UTF16ToString(d.DeviceString[:])
}
