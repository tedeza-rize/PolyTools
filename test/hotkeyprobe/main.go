//go:build windows

// Standalone probe: register Ctrl+Alt+T via raw RegisterHotKey and print when it fires.
// Run: go run ./test/hotkey_test_main.go  (from a console, not windowsgui)
package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var user32 = windows.NewLazySystemDLL("user32.dll")

var (
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessageW      = user32.NewProc("GetMessageW")
)

const (
	modAlt     = 0x0001
	modControl = 0x0002
	modNoRepeat = 0x4000
	wmHotkey   = 0x0312
	vkT        = 0x54
)

func main() {
	r, _, err := procRegisterHotKey.Call(0, 1, modAlt|modControl, vkT)
	if r == 0 {
		fmt.Println("RegisterHotKey FAILED:", err)
		return
	}
	fmt.Println("Ctrl+Alt+T registered, waiting...")

	var msg [48]byte // MSG struct is 48 bytes on amd64
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0)
		if r == 0 {
			break
		}
		wmsg := *(*uint32)(unsafe.Pointer(&msg[4])) // MSG.message
		if wmsg == wmHotkey {
			fmt.Println("HOTKEY FIRED")
		}
	}
}
