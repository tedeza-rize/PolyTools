//go:build windows

package win32

import (
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procSetWindowsHookEx   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHook  = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHook       = user32.NewProc("CallNextHookEx")
	procGetMessage         = user32.NewProc("GetMessageW")
	procPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

const (
	whKeyboardLL = 13
	whMouseLL    = 14
	wmQuit       = 0x0012

	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmRButtonDown = 0x0204
	wmRButtonUp   = 0x0205
	wmMouseMove   = 0x0200
	wmMouseWheel  = 0x020A
	wmKeyDown     = 0x0100
	wmKeyUp       = 0x0101
	wmSysKeyDown  = 0x0104
	wmSysKeyUp    = 0x0105
)

// KeyEvent is one low-level keyboard event delivered to the callback.
type KeyEvent struct {
	VK       uint16
	ScanCode uint16
	Down     bool
	Time     uint32
}

// MouseEvent is one low-level mouse event.
type MouseEvent struct {
	X, Y  int32
	Msg   uint32 // wmLButtonDown etc.
	Wheel int32  // wheel delta when Msg == wmMouseWheel
}

// hookRunner owns a dedicated OS thread that installs the hook and pumps
// messages — required for WH_*_LL hooks.
type hookRunner struct {
	threadID uintptr
	ready    chan struct{}
	stopOnce sync.Once
}

// StartKeyboardHook installs a global low-level keyboard hook on a
// dedicated thread. The callback must return quickly; return true to
// swallow the event. It runs on the hook thread — offload heavy work.
func StartKeyboardHook(cb func(KeyEvent) (swallow bool)) (stop func()) {
	return startHook(whKeyboardLL, cb)
}

// StartMouseHook installs a global low-level mouse hook.
func StartMouseHook(cb func(MouseEvent) (swallow bool)) (stop func()) {
	return startHook(whMouseLL, cb)
}

type kbdLLHookStruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	ExtraInfo   uintptr
}

type msLLHookStruct struct {
	Pt        struct{ X, Y int32 }
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

func startHook(which int, cb any) (stop func()) {
	r := &hookRunner{ready: make(chan struct{})}
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		r.threadID, _, _ = procGetCurrentThreadId.Call()
		var hookProc uintptr
		if which == whKeyboardLL {
			kcb := cb.(func(KeyEvent) bool)
			hookProc = windows.NewCallback(func(code int, wParam uintptr, lParam uintptr) uintptr {
				if code < 0 {
					ret, _, _ := procCallNextHook.Call(0, uintptr(code), wParam, lParam)
					return ret
				}
				k := (*kbdLLHookStruct)(unsafe.Pointer(lParam))
				ev := KeyEvent{
					VK:       uint16(k.VkCode),
					ScanCode: uint16(k.ScanCode),
					Down:     wParam == wmKeyDown || wParam == wmSysKeyDown,
					Time:     k.Time,
				}
				if kcb(ev) {
					return 1
				}
				ret, _, _ := procCallNextHook.Call(0, uintptr(code), wParam, lParam)
				return ret
			})
		} else {
			mcb := cb.(func(MouseEvent) bool)
			hookProc = windows.NewCallback(func(code int, wParam uintptr, lParam uintptr) uintptr {
				if code < 0 {
					ret, _, _ := procCallNextHook.Call(0, uintptr(code), wParam, lParam)
					return ret
				}
				m := (*msLLHookStruct)(unsafe.Pointer(lParam))
				ev := MouseEvent{X: m.Pt.X, Y: m.Pt.Y, Msg: uint32(wParam)}
				if wParam == wmMouseWheel {
					ev.Wheel = int32(int16(m.MouseData >> 16))
				}
				if mcb(ev) {
					return 1
				}
				ret, _, _ := procCallNextHook.Call(0, uintptr(code), wParam, lParam)
				return ret
			})
		}

		hook, _, err := procSetWindowsHookEx.Call(uintptr(which), hookProc, 0, 0)
		if hook == 0 {
			close(r.ready)
			return
		}
		_ = err
		close(r.ready)

		// Message pump — LL hooks are dispatched through this thread's queue.
		buf := make([]byte, 48)
		for {
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&buf[0])), 0, 0, 0,
			)
			if ret == 0 || int32(ret) == -1 {
				break
			}
			// DispatchMessage not needed for hooks; the pump itself is enough.
		}
		procUnhookWindowsHook.Call(hook)
	}()

	<-r.ready
	return func() {
		r.stopOnce.Do(func() {
			if r.threadID != 0 {
				procPostThreadMessage.Call(r.threadID, wmQuit, 0, 0)
			}
		})
	}
}
