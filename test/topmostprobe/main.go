//go:build windows

// Probe: toggle WS_EX_TOPMOST on a window by handle.
// Usage: go run ./test/topmost_test_main.go <hwnd-decimal>
package main

import (
	"fmt"
	"os"
	"strconv"

	"polytools/internal/win32"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: topmost_test <hwnd>")
		return
	}
	hwnd, _ := strconv.ParseUint(os.Args[1], 10, 64)
	cur := win32.IsTopMost(uintptr(hwnd))
	fmt.Printf("hwnd=%d topmost(before)=%v\n", hwnd, cur)
	win32.SetTopMost(uintptr(hwnd), !cur)
	fmt.Printf("topmost(after)=%v\n", win32.IsTopMost(uintptr(hwnd)))
}
