//go:build windows

package win32

import (
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procSetWinEventHook    = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent     = user32.NewProc("UnhookWinEvent")
	procGetMessage         = user32.NewProc("GetMessageW")
	procPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

const wmQuit = 0x0012

const (
	eventObjectCreate    = 0x8000
	eventObjectDestroy   = 0x8001
	eventObjectShow      = 0x8002
	eventSystemMoveSizeEnd = 0x000B

	wineventOutOfContext = 0x0000 // delivered to our message loop
	wineventSkipOwnProcess = 0x0002
)

// WinEvent is one accessibility/system event notification.
type WinEvent struct {
	Event uint32
	Hwnd  uintptr
	Pid   uint32
}

// StartWinEventHook subscribes to window events in [minEvent, maxEvent].
// Runs a message loop on a dedicated thread (required for out-of-context
// hooks). The callback runs on that thread — keep it cheap.
func StartWinEventHook(minEvent, maxEvent uint32, cb func(WinEvent)) (stop func()) {
	ready := make(chan struct{})
	var once sync.Once
	var threadID uintptr

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid, _, _ := procGetCurrentThreadId.Call()
		threadID = tid

		hookProc := windows.NewCallback(func(
			_ uintptr, event uint32, hwnd uintptr,
			_, _ int32, pid, _ uint32,
		) {
			cb(WinEvent{Event: event, Hwnd: hwnd, Pid: pid})
		})

		hook, _, _ := procSetWinEventHook.Call(
			uintptr(minEvent), uintptr(maxEvent), 0, hookProc, 0, 0,
			wineventOutOfContext|wineventSkipOwnProcess,
		)
		if hook == 0 {
			close(ready)
			return
		}
		defer procUnhookWinEvent.Call(hook)
		close(ready)

		buf := make([]byte, 48)
		for {
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&buf[0])), 0, 0, 0,
			)
			if ret == 0 || int32(ret) == -1 {
				break
			}
		}
	}()

	<-ready
	return func() {
		once.Do(func() {
			if threadID != 0 {
				procPostThreadMessage.Call(threadID, wmQuit, 0, 0)
			}
		})
	}
}
