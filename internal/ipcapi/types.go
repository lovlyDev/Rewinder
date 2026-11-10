package ipcapi

import "time"

type AppSummary struct {
	AppID           string `json:"appID"`
	Name            string `json:"name"`
	ExecutablePath  string `json:"executablePath"`
	LastActivityUTC int64  `json:"lastActivityUTC"`
	SnapshotCount   int    `json:"snapshotCount"`
	TrackingState   string `json:"trackingState"`
	RetentionStatus string `json:"retentionStatus"`
	Excluded        bool   `json:"excluded"` // the user stopped tracking this executable
}

type SnapshotMeta struct {
	SnapshotID   string `json:"snapshotID"`
	AppID        string `json:"appID"`
	Timestamp    int64  `json:"timestampUTC"`
	WindowsCount int    `json:"windowsCount"`
	FilesAdded   int    `json:"filesAdded"`
	FilesRemoved int    `json:"filesRemoved"`
	RestoredFrom string `json:"restoredFrom,omitempty"`
}

type SnapshotCreatedEvent struct {
	AppID      string       `json:"appID"`
	Snapshot   SnapshotMeta `json:"snapshot"`
	OccurredAt int64        `json:"occurredAtUTC"`
}

type RestoreProgressEvent struct {
	AppID      string         `json:"appID"`
	SnapshotID string         `json:"snapshotID"`
	Stage      string         `json:"stage"`
	Percent    int            `json:"percent"`
	Message    string         `json:"message"`
	Code       string         `json:"code"`
	Params     map[string]any `json:"params,omitempty"`
}

type RestoreErrorEvent struct {
	AppID      string         `json:"appID"`
	SnapshotID string         `json:"snapshotID"`
	Error      string         `json:"error"`
	Code       string         `json:"code"`
	Params     map[string]any `json:"params,omitempty"`
}

type TrackingStateChangedEvent struct {
	AppID  *string `json:"appID,omitempty"`
	State  string  `json:"state"`
	Reason string  `json:"reason,omitempty"`
	AtUTC  int64   `json:"atUTC"`
}

func NowUTC() int64 { return time.Now().UTC().UnixMilli() }

// LayoutRect is a rectangle in screen pixels.
type LayoutRect struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	W int32 `json:"w"`
	H int32 `json:"h"`
}

type LayoutWindow struct {
	Rect       LayoutRect `json:"rect"`
	Title      string     `json:"title"`
	Minimized  bool       `json:"minimized"`
	Maximized  bool       `json:"maximized"`
	Foreground bool       `json:"foreground"`
}

type LayoutMonitor struct {
	Rect    LayoutRect `json:"rect"`
	Primary bool       `json:"primary"`
}

// SnapshotLayout is what a snapshot looks like on screen: the monitors connected
// now and the saved windows, for the preview minimap.
type SnapshotLayout struct {
	Monitors []LayoutMonitor `json:"monitors"`
	Windows  []LayoutWindow  `json:"windows"`
}
