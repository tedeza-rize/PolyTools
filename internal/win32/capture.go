//go:build windows

package win32

import (
	"bytes"
	"image"
	"image/png"
	"unsafe"
)

var (
	procCreateCompatibleDC  = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBmp = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject        = gdi32.NewProc("SelectObject")
	procBitBlt              = gdi32.NewProc("BitBlt")
	procGetDIBits           = gdi32.NewProc("GetDIBits")
	procDeleteDC            = gdi32.NewProc("DeleteDC")
	procDeleteObject        = gdi32.NewProc("DeleteObject")
)

const srccopy = 0x00CC0020
const dibRGBColors = 0

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// CaptureRegion grabs a screen rectangle and returns it PNG-encoded.
func CaptureRegion(x, y, w, h int32) ([]byte, error) {
	srcDC, _, _ := procGetDC.Call(0)
	if srcDC == 0 {
		return nil, errNoDC
	}
	defer procReleaseDC.Call(0, srcDC)

	memDC, _, _ := procCreateCompatibleDC.Call(srcDC)
	if memDC == 0 {
		return nil, errNoDC
	}
	defer procDeleteDC.Call(memDC)

	bmp, _, _ := procCreateCompatibleBmp.Call(srcDC, uintptr(w), uintptr(h))
	if bmp == 0 {
		return nil, errNoDC
	}
	defer procDeleteObject.Call(bmp)
	procSelectObject.Call(memDC, bmp)

	procBitBlt.Call(memDC, 0, 0, uintptr(w), uintptr(h), srcDC,
		uintptr(x), uintptr(y), srccopy)

	// Read back as 32bpp BGRA (top-down: negative height).
	bi := bitmapInfoHeader{
		Size:      uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:     w,
		Height:    -h,
		Planes:    1,
		BitCount:  32,
	}
	pix := make([]byte, w*h*4)
	procGetDIBits.Call(memDC, bmp, 0, uintptr(h),
		uintptr(unsafe.Pointer(&pix[0])),
		uintptr(unsafe.Pointer(&bi)), dibRGBColors)

	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for i := 0; i < int(w)*int(h); i++ {
		// BGRA -> RGBA
		img.Pix[i*4+0] = pix[i*4+2]
		img.Pix[i*4+1] = pix[i*4+1]
		img.Pix[i*4+2] = pix[i*4+0]
		img.Pix[i*4+3] = 255
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var errNoDC = &win32Error{"device context unavailable"}

type win32Error struct{ msg string }

func (e *win32Error) Error() string { return e.msg }
