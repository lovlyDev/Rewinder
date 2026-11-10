//go:build windows

package winapi

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	gdi32 = windows.NewLazySystemDLL("gdi32.dll")

	procPrivateExtractIconsW = user32.NewProc("PrivateExtractIconsW")
	procGetIconInfo          = user32.NewProc("GetIconInfo")
	procDestroyIcon          = user32.NewProc("DestroyIcon")
	procGetDC                = user32.NewProc("GetDC")
	procReleaseDC            = user32.NewProc("ReleaseDC")
	procGetDIBits            = gdi32.NewProc("GetDIBits")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
)

var ErrNoIcon = errors.New("no icon")

type iconInfo struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type bitmapInfo struct {
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
	Colors        [4]uint32 // room for a colour table GetDIBits may write
}

// AppIconPNG returns the executable's main icon at size×size pixels as PNG.
func AppIconPNG(exePath string, size int) ([]byte, error) {
	path, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return nil, err
	}
	var hicon uintptr
	n, _, _ := procPrivateExtractIconsW.Call(uintptr(unsafe.Pointer(path)), 0, uintptr(size), uintptr(size),
		uintptr(unsafe.Pointer(&hicon)), 0, 1, 0)
	if n == 0 || n == 0xFFFFFFFF || hicon == 0 {
		return nil, ErrNoIcon
	}
	defer procDestroyIcon.Call(hicon)

	var ii iconInfo
	if r, _, _ := procGetIconInfo.Call(hicon, uintptr(unsafe.Pointer(&ii))); r == 0 {
		return nil, ErrNoIcon
	}
	defer procDeleteObject.Call(ii.HbmMask)
	if ii.HbmColor == 0 {
		return nil, ErrNoIcon // monochrome icon: not worth showing
	}
	defer procDeleteObject.Call(ii.HbmColor)

	bi := bitmapInfo{Size: 40, Width: int32(size), Height: -int32(size), Planes: 1, BitCount: 32} // top-down BGRA
	pixels := make([]byte, size*size*4)
	dc, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, dc)
	if r, _, _ := procGetDIBits.Call(dc, ii.HbmColor, 0, uintptr(size),
		uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&bi)), 0); r == 0 {
		return nil, ErrNoIcon
	}

	hasAlpha := false
	for i := 3; i < len(pixels); i += 4 {
		if pixels[i] != 0 {
			hasAlpha = true
			break
		}
	}
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for i := 0; i < len(pixels); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2] = pixels[i+2], pixels[i+1], pixels[i]
		if hasAlpha {
			img.Pix[i+3] = pixels[i+3]
		} else {
			img.Pix[i+3] = 255 // old icons keep transparency in the mask only
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
