//go:build windows

package winapi

import (
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// ExplorerFolders maps each open File Explorer window (HWND) to the folder it shows.
// Virtual locations without a file system path (This PC, Recycle Bin) are skipped.
func ExplorerFolders() map[uintptr]string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	out := map[uintptr]string{}
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		// S_FALSE (already initialised) is reported as an error by go-ole; carry on.
		if oe, ok := err.(*ole.OleError); !ok || oe.Code() != 1 {
			return out
		}
	}
	defer ole.CoUninitialize()

	unk, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return out
	}
	defer unk.Release()
	shell, err := unk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return out
	}
	defer shell.Release()

	winsV, err := oleutil.CallMethod(shell, "Windows")
	if err != nil {
		return out
	}
	wins := winsV.ToIDispatch()
	defer wins.Release()

	countV, err := oleutil.GetProperty(wins, "Count")
	if err != nil {
		return out
	}
	count := int(countV.Val)
	for i := 0; i < count; i++ {
		itemV, err := oleutil.CallMethod(wins, "Item", i)
		if err != nil {
			continue
		}
		item := itemV.ToIDispatch()
		if item == nil {
			continue
		}
		hwnd, path := explorerItem(item)
		item.Release()
		if hwnd != 0 && path != "" {
			out[hwnd] = path
		}
	}
	return out
}

func explorerItem(item *ole.IDispatch) (uintptr, string) {
	hv, err := oleutil.GetProperty(item, "HWND")
	if err != nil {
		return 0, ""
	}
	hwnd := uintptr(hv.Val)

	docV, err := oleutil.GetProperty(item, "Document")
	if err != nil || docV.VT != ole.VT_DISPATCH {
		return hwnd, ""
	}
	doc := docV.ToIDispatch()
	if doc == nil {
		return hwnd, ""
	}
	defer doc.Release()
	folderV, err := oleutil.GetProperty(doc, "Folder")
	if err != nil || folderV.VT != ole.VT_DISPATCH {
		return hwnd, ""
	}
	folder := folderV.ToIDispatch()
	defer folder.Release()
	selfV, err := oleutil.GetProperty(folder, "Self")
	if err != nil || selfV.VT != ole.VT_DISPATCH {
		return hwnd, ""
	}
	self := selfV.ToIDispatch()
	defer self.Release()
	isFS, err := oleutil.GetProperty(self, "IsFileSystem")
	if err != nil {
		return hwnd, ""
	}
	if fs, ok := isFS.Value().(bool); !ok || !fs {
		return hwnd, ""
	}
	pathV, err := oleutil.GetProperty(self, "Path")
	if err != nil {
		return hwnd, ""
	}
	return hwnd, pathV.ToString()
}
