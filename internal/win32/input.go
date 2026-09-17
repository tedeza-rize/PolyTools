//go:build windows

package win32

import (
	"unsafe"
)

var (
	procSendInput         = user32.NewProc("SendInput")
	procGetAsyncKeyState  = user32.NewProc("GetAsyncKeyState")
	procGetKeyState       = user32.NewProc("GetKeyState")
	procGetLastInputInfo  = user32.NewProc("GetLastInputInfo")
	procToUnicode         = user32.NewProc("ToUnicode")
	procGetKeyboardLayout = user32.NewProc("GetKeyboardLayout")
	procMapVirtualKey     = user32.NewProc("MapVirtualKeyW")
)

const (
	inputKeyboard = 1

	keyEventFExtendedKey = 0x0001
	keyEventFKeyUp       = 0x0002
	keyEventFUnicode     = 0x0004
	keyEventFScancode    = 0x0008

	vkBack    = 0x08
	vkTab     = 0x09
	vkReturn  = 0x0D
	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12 // Alt
	vkCapital = 0x14
	vkEscape  = 0x1B
	vkSpace   = 0x20
	vkLWin    = 0x5B
	vkRWin    = 0x5C
)

type keybdInput struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// input mirrors the Win32 INPUT struct (keyboard variant — the union is
// sized for MOUSEINPUT, the largest member).
type input struct {
	Type uint32
	_    uint32 // padding for 64-bit alignment of the union
	Ki   keybdInput
	_    [16]byte // union tail padding (mouse/hw members)
}

func sendInputs(in []input) {
	procSendInput.Call(
		uintptr(len(in)),
		uintptr(unsafe.Pointer(&in[0])),
		unsafe.Sizeof(in[0]),
	)
}

// KeyPress sends a virtual-key press+release.
func KeyPress(vk uint16) {
	down := input{Type: inputKeyboard, Ki: keybdInput{Vk: vk}}
	up := input{Type: inputKeyboard, Ki: keybdInput{Vk: vk, Flags: keyEventFKeyUp}}
	sendInputs([]input{down, up})
}

// KeyDown / KeyUp send only one half of a key event (for chord injection).
func KeyDown(vk uint16) {
	sendInputs([]input{{Type: inputKeyboard, Ki: keybdInput{Vk: vk}}})
}
func KeyUp(vk uint16) {
	sendInputs([]input{{Type: inputKeyboard, Ki: keybdInput{Vk: vk, Flags: keyEventFKeyUp}}})
}

// SendKeys presses the given virtual keys together (e.g. Ctrl+V).
func SendKeys(vks ...uint16) {
	in := make([]input, 0, len(vks)*2)
	for _, vk := range vks {
		in = append(in, input{Type: inputKeyboard, Ki: keybdInput{Vk: vk}})
	}
	for i := len(vks) - 1; i >= 0; i-- {
		in = append(in, input{Type: inputKeyboard, Ki: keybdInput{Vk: vks[i], Flags: keyEventFKeyUp}})
	}
	sendInputs(in)
}

// SendBackspaces erases n characters before the caret.
func SendBackspaces(n int) {
	in := make([]input, 0, n*2)
	for i := 0; i < n; i++ {
		in = append(in,
			input{Type: inputKeyboard, Ki: keybdInput{Vk: vkBack}},
			input{Type: inputKeyboard, Ki: keybdInput{Vk: vkBack, Flags: keyEventFKeyUp}},
		)
	}
	sendInputs(in)
}

// SendText types arbitrary Unicode text (KEYEVENTF_UNICODE, works for any
// script without touching the clipboard).
func SendText(s string) {
	runes := []rune(s)
	in := make([]input, 0, len(runes)*2)
	for _, r := range runes {
		// Encode as UTF-16; surrogate pairs send as two inputs each.
		for _, u := range utf16Of(r) {
			in = append(in,
				input{Type: inputKeyboard, Ki: keybdInput{Scan: u, Flags: keyEventFUnicode}},
				input{Type: inputKeyboard, Ki: keybdInput{Scan: u, Flags: keyEventFUnicode | keyEventFKeyUp}},
			)
		}
	}
	sendInputs(in)
}

func utf16Of(r rune) []uint16 {
	if r <= 0xFFFF {
		return []uint16{uint16(r)}
	}
	r -= 0x10000
	return []uint16{uint16(0xD800 + r>>10), uint16(0xDC00 + r&0x3FF)}
}

// KeyDownState reports whether the virtual key is currently held.
func KeyDownState(vk uint16) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// ModifierHeld reports whether the named modifier group is held.
func ModifierHeld(name string) bool {
	switch name {
	case "alt":
		return KeyDownState(vkMenu)
	case "ctrl":
		return KeyDownState(vkControl)
	case "shift":
		return KeyDownState(vkShift)
	case "win":
		return KeyDownState(vkLWin) || KeyDownState(vkRWin)
	}
	return false
}

type lastInputInfo struct {
	CbSize uint32
	DwTime uint32
}

// IdleSeconds returns how long since the last keyboard/mouse input.
func IdleSeconds() uint32 {
	var lii lastInputInfo
	lii.CbSize = uint32(unsafe.Sizeof(lii))
	procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	tick, _, _ := procGetTickCount64.Call()
	return uint32((uint64(tick) - uint64(lii.DwTime)) / 1000)
}

// VKFromName maps friendly key names to virtual-key codes (for remapping
// tables like "CapsLock = Escape").
var VKByName = map[string]uint16{
	"backspace": vkBack, "tab": vkTab, "enter": vkReturn, "return": vkReturn,
	"shift": vkShift, "ctrl": vkControl, "control": vkControl, "alt": vkMenu,
	"capslock": vkCapital, "escape": vkEscape, "esc": vkEscape, "space": vkSpace,
	"lwin": vkLWin, "rwin": vkRWin, "win": vkLWin,
	"up": 0x26, "down": 0x28, "left": 0x25, "right": 0x27,
	"insert": 0x2D, "delete": 0x2E, "home": 0x24, "end": 0x23,
	"pageup": 0x21, "pagedown": 0x22,
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74, "f6": 0x75,
	"f7": 0x76, "f8": 0x77, "f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
	"numlock": 0x90, "scrolllock": 0x91, "printscreen": 0x2C, "pause": 0x13,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34, "5": 0x35,
	"6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45, "f": 0x46,
	"g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A, "k": 0x4B, "l": 0x4C,
	"m": 0x4D, "n": 0x4E, "o": 0x4F, "p": 0x50, "q": 0x51, "r": 0x52,
	"s": 0x53, "t": 0x54, "u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58,
	"y": 0x59, "z": 0x5A,
	";": 0xBA, "=": 0xBB, ",": 0xBC, "-": 0xBD, ".": 0xBE, "/": 0xBF,
	"`": 0xC0, "[": 0xDB, "\\": 0xDC, "]": 0xDD, "'": 0xDE,
}

// VKName is the inverse of VKByName for display.
func VKName(vk uint16) string {
	for name, v := range VKByName {
		if v == vk && len(name) > 1 {
			return name
		}
	}
	if vk >= 0x30 && vk <= 0x39 || vk >= 0x41 && vk <= 0x5A {
		return string(rune(vk))
	}
	return ""
}
