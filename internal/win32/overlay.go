//go:build windows

package win32

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Overlay is a click-through topmost layered window whose pixels are drawn
// by Go code into a 32bpp BGRA buffer and pushed via UpdateLayeredWindow.
type Overlay struct {
	hwnd   uintptr
	w, h   int32
	pix    []byte // BGRA premultiplied
	bmp    uintptr
	memDC  uintptr
	bits   unsafe.Pointer
	mu     sync.Mutex
	closed bool
}

var (
	procRegisterClassEx      = user32.NewProc("RegisterClassExW")
	procCreateWindowEx       = user32.NewProc("CreateWindowExW")
	procDefWindowProc        = user32.NewProc("DefWindowProcW")
	procUpdateLayered        = user32.NewProc("UpdateLayeredWindow")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procCreateDIBSection     = gdi32.NewProc("CreateDIBSection")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procSelectObject         = gdi32.NewProc("SelectObject")
	procDeleteDC             = gdi32.NewProc("DeleteDC")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
)

const (
	wsExNoActivate  = 0x08000000
	wsExToolWindow  = 0x00000080
	wsExTransparent = 0x00000020
	wsPopup         = 0x80000000
	wsVisible       = 0x10000000

	ulwAlpha = 0x00000002

	hwndTopMostV = ^uintptr(0) // -1

	dibRGBColors = 0
)

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

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

var overlayClassOnce sync.Once
var overlayClassName *uint16

func registerOverlayClass() {
	overlayClassOnce.Do(func() {
		overlayClassName, _ = windows.UTF16PtrFromString("PolyToolsOverlay")
		wndProc := windows.NewCallback(func(hwnd, msg, w, l uintptr) uintptr {
			r, _, _ := procDefWindowProc.Call(hwnd, msg, w, l)
			return r
		})
		var wc wndClassEx
		wc.CbSize = uint32(unsafe.Sizeof(wc))
		wc.LpfnWndProc = wndProc
		wc.LpszClassName = overlayClassName
		procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	})
}

type pointl struct{ X, Y int32 }
type sizel struct{ Cx, Cy int32 }

type blendFunction struct {
	BlendOp             uint8
	BlendFlags          uint8
	SourceConstantAlpha uint8
	AlphaFormat         uint8
}

// NewOverlay creates a hidden click-through layered window at (x,y,w,h).
// draw gets a BGRA pixel buffer to fill; call Present() to show changes.
func NewOverlay(x, y, w, h int32) (*Overlay, error) {
	registerOverlayClass()

	hwnd, _, err := procCreateWindowEx.Call(
		wsExLayered|wsExTransparent|wsExNoActivate|wsExToolWindow|0x00080000, // + WS_EX_TOPMOST
		uintptr(unsafe.Pointer(overlayClassName)),
		0,
		wsPopup,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		0, 0, 0, 0,
	)
	if hwnd == 0 {
		return nil, err
	}

	screenDC, _, _ := procGetDC.Call(0)
	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	procReleaseDC.Call(0, screenDC)

	bi := bitmapInfoHeader{
		Size:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:    w,
		Height:   -h,
		Planes:   1,
		BitCount: 32,
	}
	var bits unsafe.Pointer
	bmp, _, _ := procCreateDIBSection.Call(
		memDC, uintptr(unsafe.Pointer(&bi)), dibRGBColors,
		uintptr(unsafe.Pointer(&bits)), 0, 0,
	)
	procSelectObject.Call(memDC, bmp)

	return &Overlay{
		hwnd: hwnd, w: w, h: h,
		pix:  unsafe.Slice((*byte)(bits), int(w)*int(h)*4),
		bmp:  bmp, memDC: memDC, bits: bits,
	}, nil
}

// Present pushes the buffer to the screen and shows the window.
func (o *Overlay) Present() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return
	}
	src := pointl{0, 0}
	dst := pointl{}
	sz := sizel{o.w, o.h}
	blend := blendFunction{BlendOp: 0, SourceConstantAlpha: 255, AlphaFormat: 1} // AC_SRC_ALPHA
	procUpdateLayered.Call(
		o.hwnd, 0, uintptr(unsafe.Pointer(&dst)),
		uintptr(unsafe.Pointer(&sz)), o.memDC,
		uintptr(unsafe.Pointer(&src)), 0,
		uintptr(unsafe.Pointer(&blend)), ulwAlpha,
	)
	procShowWindow.Call(o.hwnd, swShowNA)
}

// Fill sets every pixel to the given BGRA premultiplied color.
func (o *Overlay) Fill(b, g, r, a uint8) {
	for i := 0; i < len(o.pix); i += 4 {
		o.pix[i], o.pix[i+1], o.pix[i+2], o.pix[i+3] = b, g, r, a
	}
}

// FillRect paints a solid rect (premultiplied).
func (o *Overlay) FillRect(x, y, w, h int32, b, g, r, a uint8) {
	for yy := y; yy < y+h; yy++ {
		if yy < 0 || yy >= o.h {
			continue
		}
		for xx := x; xx < x+w; xx++ {
			if xx < 0 || xx >= o.w {
				continue
			}
			i := (yy*o.w + xx) * 4
			// simple source-over blend for partial alpha
			if a == 255 {
				o.pix[i], o.pix[i+1], o.pix[i+2], o.pix[i+3] = b, g, r, a
			} else {
				da := uint32(o.pix[i+3])
				fa := uint32(a)
				outA := fa + da*(255-fa)/255
				if outA == 0 {
					continue
				}
				for c := int32(0); c < 3; c++ {
					dc := uint32(o.pix[i+c])
					var sc uint32
					switch c {
					case 0:
						sc = uint32(b)
					case 1:
						sc = uint32(g)
					default:
						sc = uint32(r)
					}
					o.pix[i+c] = uint8((sc*fa + dc*da*(255-fa)/255) / outA)
				}
				o.pix[i+3] = uint8(outA)
			}
		}
	}
}

var (
	procCreateFont   = gdi32.NewProc("CreateFontW")
	procTextOut      = gdi32.NewProc("TextOutW")
	procSetTextColor = gdi32.NewProc("SetTextColor")
	procSetBkMode    = gdi32.NewProc("SetBkMode")
)

// Text draws white text at (x,y) with the given pixel height using GDI.
// GDI writes RGB but leaves the alpha byte at 0 on DIB sections, so after
// drawing we convert luminance to alpha for antialiased white text.
func (o *Overlay) Text(s string, x, y, height int32) {
	font, _, _ := procCreateFont.Call(
		uintptr(height), 0, 0, 0, 700, 0, 0, 0,
		1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(fontName())),
	)
	defer procDeleteObject.Call(font)
	procSelectObject.Call(o.memDC, font)
	procSetBkMode.Call(o.memDC, 1) // TRANSPARENT
	procSetTextColor.Call(o.memDC, 0x00FFFFFF)
	u16, _ := windows.UTF16FromString(s)
	procTextOut.Call(o.memDC, uintptr(x), uintptr(y),
		uintptr(unsafe.Pointer(&u16[0])), uintptr(len(u16)-1))

	// Alpha-fix pass over the whole buffer: pixels GDI touched have alpha 0
	// but RGB > 0; use luminance as coverage for white text.
	for i := 0; i+3 < len(o.pix); i += 4 {
		if o.pix[i+3] == 0 && (o.pix[i]|o.pix[i+1]|o.pix[i+2]) != 0 {
			lum := o.pix[i]
			if o.pix[i+1] > lum {
				lum = o.pix[i+1]
			}
			if o.pix[i+2] > lum {
				lum = o.pix[i+2]
			}
			o.pix[i], o.pix[i+1], o.pix[i+2], o.pix[i+3] = lum, lum, lum, lum
		}
	}
}

func fontName() *uint16 {
	name, _ := windows.UTF16PtrFromString("Segoe UI")
	return name
}

// Destroy frees the window and GDI resources.
func (o *Overlay) Destroy() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return
	}
	o.closed = true
	procDestroyWindow.Call(o.hwnd)
	procDeleteObject.Call(o.bmp)
	procDeleteDC.Call(o.memDC)
}
