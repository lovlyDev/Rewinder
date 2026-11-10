package plugins

import (
	"path/filepath"
	"strings"

	"Rewinder/internal/state"
)

// OfficePlugin remembers which documents Word, Excel and PowerPoint have open.
// Office keeps a handle on each open document, so open file handles reveal them.
type OfficePlugin struct{}

func NewOfficePlugin() *OfficePlugin { return &OfficePlugin{} }

func (p *OfficePlugin) ID() string { return "office" }

func (p *OfficePlugin) CanHandle(exePath string) bool {
	switch exeBase(exePath) {
	case "winword.exe", "excel.exe", "powerpnt.exe":
		return true
	}
	return false
}

func (p *OfficePlugin) NeedsOpenFiles() bool { return true }

var officeExts = map[string]bool{
	".doc": true, ".docx": true, ".docm": true, ".rtf": true,
	".xls": true, ".xlsx": true, ".xlsm": true, ".xlsb": true, ".csv": true,
	".ppt": true, ".pptx": true, ".pptm": true,
}

const maxOfficeDocs = 10

// OfficeDocs filters open files down to user documents.
func OfficeDocs(files []state.FileRef) []string {
	seen := map[string]bool{}
	var docs []string
	for _, f := range files {
		name := filepath.Base(f.Path)
		if strings.HasPrefix(name, "~$") || !officeExts[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		low := strings.ToLower(f.Path)
		if strings.Contains(low, `\appdata\local\temp\`) || strings.Contains(low, `\microsoft\office\`) || seen[low] {
			continue
		}
		seen[low] = true
		docs = append(docs, f.Path)
		if len(docs) >= maxOfficeDocs {
			break
		}
	}
	return docs
}

func (p *OfficePlugin) Capture(_ Env, app *state.AppState) {
	docs := OfficeDocs(app.OpenFiles)
	app.PluginData[p.ID()] = state.PluginState{Paths: docs}
	// Office window titles start with the document name: "Report.docx - Word".
	for i := range app.Windows {
		for _, d := range docs {
			if strings.HasPrefix(strings.ToLower(app.Windows[i].Title), strings.ToLower(filepath.Base(d))) {
				app.Windows[i].Path = d
				break
			}
		}
	}
}

func (p *OfficePlugin) Relaunch(env Env, saved *state.AppState) (bool, error) {
	docs := pathsOf(saved, p.ID())
	if len(docs) == 0 {
		return false, nil
	}
	openAll(env, docs)
	return true, nil
}

func (p *OfficePlugin) Reopen(env Env, saved, current *state.AppState) int {
	return openAll(env, missing(pathsOf(saved, p.ID()), pathsOf(current, p.ID())))
}
