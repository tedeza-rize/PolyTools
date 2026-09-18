//go:build windows

package win32

import (
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// SystemParametersInfo actions for the built-in Windows screensaver.
const (
	spiGetScreenSaveActive  = 0x0010
	spiSetScreenSaveActive  = 0x0011
	spiGetScreenSaveTimeout = 0x000E
)

const (
	spifUpdateIniFile = 0x01
	spifSendChange    = 0x02
)

var (
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
	procLockWorkStation      = user32.NewProc("LockWorkStation")
)

// ScreensaverActive reports whether the Windows screensaver is enabled.
func ScreensaverActive() bool {
	var active uint32 // BOOL
	procSystemParametersInfo.Call(spiGetScreenSaveActive, 0,
		uintptr(unsafe.Pointer(&active)), 0)
	return active != 0
}

// SetScreensaverActive enables or disables the Windows screensaver. The
// change is written to the user profile and broadcast to other windows so
// it takes effect without a sign-out.
func SetScreensaverActive(active bool) {
	var v uintptr
	if active {
		v = 1
	}
	procSystemParametersInfo.Call(spiSetScreenSaveActive, v, 0,
		spifUpdateIniFile|spifSendChange)
}

// ScreensaverTimeout returns the Windows screensaver idle timeout in
// seconds, or 0 when it cannot be determined.
func ScreensaverTimeout() uint32 {
	var timeout uint32
	procSystemParametersInfo.Call(spiGetScreenSaveTimeout, 0,
		uintptr(unsafe.Pointer(&timeout)), 0)
	return timeout
}

// ScreensaverPath returns the configured .scr executable, "" when none is
// configured (Windows then just shows a blank screen on timeout).
func ScreensaverPath() string {
	return desktopValue("SCRNSAVE.EXE")
}

// ScreensaverSecure reports whether Windows locks the workstation when the
// screensaver engages ("On resume, display logon screen").
func ScreensaverSecure() bool {
	return desktopValue("ScreenSaverIsSecure") == "1"
}

func desktopValue(name string) string {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Control Panel\Desktop`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, err := k.GetStringValue(name)
	if err != nil {
		return ""
	}
	return v
}

// LockWorkStation locks the session, as if the user pressed Win+L.
func LockWorkStation() {
	procLockWorkStation.Call()
}
