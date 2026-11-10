//go:build windows

package update

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Authenticode check through WinVerifyTrust. The rule an update follows: if the
// running Rewinder is signed, the new one must carry a valid signature too.
// (Unsigned development builds can only update to builds verified by checksum.)

var (
	wintrust           = windows.NewLazySystemDLL("wintrust.dll")
	procWinVerifyTrust = wintrust.NewProc("WinVerifyTrust")
)

// WINTRUST_ACTION_GENERIC_VERIFY_V2
var actionGenericVerifyV2 = windows.GUID{Data1: 0x00AAC56B, Data2: 0xCD44, Data3: 0x11D0, Data4: [8]byte{0x8C, 0xC2, 0x00, 0xC0, 0x4F, 0xC2, 0x95, 0xEE}}

type winTrustFileInfo struct {
	cbStruct       uint32
	pcwszFilePath  *uint16
	hFile          windows.Handle
	pgKnownSubject *windows.GUID
}

type winTrustData struct {
	cbStruct            uint32
	pPolicyCallbackData uintptr
	pSIPClientData      uintptr
	dwUIChoice          uint32
	fdwRevocationChecks uint32
	dwUnionChoice       uint32
	pFile               *winTrustFileInfo
	dwStateAction       uint32
	hWVTStateData       windows.Handle
	pwszURLReference    *uint16
	dwProvFlags         uint32
	dwUIContext         uint32
	pSignatureSettings  uintptr
}

const (
	wtdUINone            = 2
	wtdRevokeNone        = 0
	wtdChoiceFile        = 1
	wtdStateActionVerify = 1
	wtdStateActionClose  = 2
	wtdCacheOnlyURLRetr  = 0x1000
)

// Signed reports whether path has a valid, trusted Authenticode signature.
func Signed(path string) bool {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	file := winTrustFileInfo{pcwszFilePath: p}
	file.cbStruct = uint32(unsafe.Sizeof(file))
	data := winTrustData{
		dwUIChoice:          wtdUINone,
		fdwRevocationChecks: wtdRevokeNone,
		dwUnionChoice:       wtdChoiceFile,
		pFile:               &file,
		dwStateAction:       wtdStateActionVerify,
		dwProvFlags:         wtdCacheOnlyURLRetr,
	}
	data.cbStruct = uint32(unsafe.Sizeof(data))
	r, _, _ := procWinVerifyTrust.Call(uintptr(windows.InvalidHandle), uintptr(unsafe.Pointer(&actionGenericVerifyV2)), uintptr(unsafe.Pointer(&data)))
	data.dwStateAction = wtdStateActionClose
	_, _, _ = procWinVerifyTrust.Call(uintptr(windows.InvalidHandle), uintptr(unsafe.Pointer(&actionGenericVerifyV2)), uintptr(unsafe.Pointer(&data)))
	return r == 0
}

// CheckSignature enforces the rule above for a downloaded update.
func CheckSignature(current, downloaded string) error {
	if Signed(current) && !Signed(downloaded) {
		return errors.New("update.signature")
	}
	return nil
}
