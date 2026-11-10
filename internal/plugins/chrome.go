package plugins

import (
	"strings"

	"Rewinder/internal/state"
)

// ChromePlugin handles Chromium browsers. A browser keeps its own session, so the
// useful thing to remember is which profile directory it ran with.
type ChromePlugin struct{}

func NewChromePlugin() *ChromePlugin { return &ChromePlugin{} }

func (p *ChromePlugin) ID() string { return "chrome" }

func (p *ChromePlugin) CanHandle(exePath string) bool {
	switch exeBase(exePath) {
	case "chrome.exe", "msedge.exe", "brave.exe", "vivaldi.exe", "opera.exe":
		return true
	}
	return false
}

func (p *ChromePlugin) NeedsOpenFiles() bool { return false }

func (p *ChromePlugin) Capture(_ Env, app *state.AppState) {
	vals := map[string]string{}
	if dir := extractFlagValue(app.CommandLine, "--user-data-dir"); dir != "" {
		vals["userDataDir"] = dir
	}
	if prof := extractFlagValue(app.CommandLine, "--profile-directory"); prof != "" {
		vals["profileDirectory"] = prof
	}
	app.PluginData[p.ID()] = state.PluginState{Values: vals}
}

func (p *ChromePlugin) Relaunch(env Env, saved *state.AppState) (bool, error) {
	if env.Launch == nil {
		return false, nil
	}
	return true, env.Launch(saved.ExecutablePath, p.launchCommand(saved), saved.WorkingDir)
}

// launchCommand rebuilds a clean command line. The captured one is not reused,
// because it may contain a URL or a one-off flag the browser was started with.
func (p *ChromePlugin) launchCommand(saved *state.AppState) string {
	parts := []string{quote(saved.ExecutablePath)}
	vals := saved.PluginData[p.ID()].Values
	if dir := vals["userDataDir"]; dir != "" {
		parts = append(parts, "--user-data-dir="+quote(dir))
	}
	if prof := vals["profileDirectory"]; prof != "" {
		parts = append(parts, "--profile-directory="+quote(prof))
	}
	parts = append(parts, "--restore-last-session")
	return strings.Join(parts, " ")
}

func (p *ChromePlugin) Reopen(Env, *state.AppState, *state.AppState) int { return 0 }

// extractFlagValue returns the value of --flag=value, --flag="value" or --flag value.
func extractFlagValue(cmd string, flag string) string {
	i := strings.Index(cmd, flag)
	for i != -1 {
		end := i + len(flag)
		if end == len(cmd) || cmd[end] == '=' || cmd[end] == ' ' || cmd[end] == '\t' {
			break
		}
		// matched a longer flag with the same prefix, keep looking
		next := strings.Index(cmd[end:], flag)
		if next == -1 {
			return ""
		}
		i = end + next
	}
	if i == -1 {
		return ""
	}
	rest := strings.TrimLeft(cmd[i+len(flag):], " \t")
	rest = strings.TrimPrefix(rest, "=")
	if strings.HasPrefix(rest, `"`) {
		rest = rest[1:]
		j := strings.Index(rest, `"`)
		if j == -1 {
			return rest
		}
		return rest[:j]
	}
	if k := strings.IndexAny(rest, " \t"); k != -1 {
		return rest[:k]
	}
	return rest
}

func quote(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}
