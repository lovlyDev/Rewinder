package state

import "time"

type Rect struct {
	Left   int32 `json:"left"`
	Top    int32 `json:"top"`
	Right  int32 `json:"right"`
	Bottom int32 `json:"bottom"`
}

type Point struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

// WindowState is one application window. Geometry is stored as a WINDOWPLACEMENT
// (normal rect + show command), which is what SetWindowPlacement needs to put a
// minimized or maximized window back correctly.
type WindowState struct {
	HWND       uintptr `json:"hwnd"`
	ClassName  string  `json:"className,omitempty"`
	Title      string  `json:"title,omitempty"`
	ZOrder     int     `json:"zOrder"` // 0 = topmost among this app's windows
	MonitorID  string  `json:"monitorID,omitempty"`
	Rect       Rect    `json:"rect"`       // actual on-screen bounds at capture time
	NormalRect Rect    `json:"normalRect"` // restored (non-minimized, non-maximized) bounds
	ShowCmd    uint32  `json:"showCmd"`
	Flags      uint32  `json:"flags,omitempty"`
	MinPos     Point   `json:"minPos"`
	MaxPos     Point   `json:"maxPos"`

	IsForeground bool `json:"isForeground"`
	IsMinimized  bool `json:"isMinimized"`
	IsMaximized  bool `json:"isMaximized"`

	// Path is the document or folder the window shows, when a plugin knows it.
	Path string `json:"path,omitempty"`
}

type FileRef struct {
	Path string `json:"path"`
}

// PluginState is plugin-owned data. It uses concrete types so it survives a JSON round trip.
type PluginState struct {
	Paths  []string          `json:"paths,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

type AppState struct {
	AppID          string `json:"appID"`
	PID            int    `json:"pid"`
	ExecutablePath string `json:"executablePath"`
	CommandLine    string `json:"commandLine,omitempty"`
	WorkingDir     string `json:"workingDir,omitempty"`
	AppUserModelID string `json:"appUserModelID,omitempty"` // set for packaged (Store/MSIX) apps

	ForegroundWindowClass string                 `json:"foregroundWindowClass,omitempty"`
	Windows               []WindowState          `json:"windows"`
	OpenFiles             []FileRef              `json:"openFiles,omitempty"`
	PluginData            map[string]PluginState `json:"pluginData,omitempty"`
	Timestamp             time.Time              `json:"timestamp"`
}

// Clone returns a deep copy, so stored snapshots never alias live data.
func (a AppState) Clone() AppState {
	out := a
	out.Windows = append([]WindowState(nil), a.Windows...)
	out.OpenFiles = append([]FileRef(nil), a.OpenFiles...)
	if a.PluginData != nil {
		out.PluginData = make(map[string]PluginState, len(a.PluginData))
		for k, v := range a.PluginData {
			ps := PluginState{Paths: append([]string(nil), v.Paths...)}
			if v.Values != nil {
				ps.Values = make(map[string]string, len(v.Values))
				for vk, vv := range v.Values {
					ps.Values[vk] = vv
				}
			}
			out.PluginData[k] = ps
		}
	}
	return out
}
