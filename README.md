# <img src="build/appicon.png" width="32"> Rewinder — Time Machine for Windows windows

**Remembers where and how your windows were, and puts them back with one key.**

[Русская версия](README-RU.md)

## What it does

- **Undo for your window layout.** Every move, maximize or minimize goes into a history. `Ctrl + Alt + Z` opens an overlay with one step back already selected for the app you were in. `Enter` puts the windows back.
- **Relaunches closed apps.** If an app is closed, Rewinder starts it with the same arguments and puts its windows where they were. Store/MSIX apps are started by their AppUserModelID.
- **Knows a few apps better:**
  - **File Explorer**: remembers which folders were open and reopens the missing ones.
  - **Word, Excel, PowerPoint**: remember open documents and reopen them.
  - **Chrome, Edge, Brave, Vivaldi, Opera**: relaunch with the same profile and `--restore-last-session`.
- **History survives a reboot**: 7 days, up to 500 snapshots per app.
- **No administrator rights needed.**

## What it does not do

Rewinder restores **windows**, not their **content**: unsaved text, the exact browser tabs at snapshot time, or scroll positions cannot be brought back. Browser tabs are restored by the browser itself.

## Install

1. Download `Rewinder.exe` from [releases](../../releases).
2. Run it. Rewinder lives in the system tray.
3. Turn on autostart in Settings if you like (it then starts hidden).

## Keys

| Key | Action |
|---|---|
| `Ctrl + Alt + Z` | open the overlay for the current app |
| `←` / `→` or mouse wheel | step through history |
| `Enter` | restore the selected snapshot |
| `Esc` | close the overlay |

The main window (Timeline) lists every app in the history, including closed ones, and can restore any of them.

## Privacy

- Only window layout and titles, open folder and Office document paths, and the launch command line are stored.
- The clipboard is never read.
- Password managers (KeePass, KeePassXC, 1Password, Bitwarden) are never tracked.
- Everything stays in `%LOCALAPPDATA%\Rewinder\history` and is never sent anywhere.

## Build

Requires Go 1.23+, Node.js and the [Wails CLI](https://wails.io) v2.11.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.11.0
GOTOOLCHAIN=go1.24.0 wails build
```

The Wails 2.11 bindings generator does not work with Go 1.25 or newer. With a recent Go, build with `GOTOOLCHAIN=go1.24.0 wails build`.

## Tests

```bash
go test ./...
```

Integration tests open real windows on the desktop, move, close and restore them:

```bash
go test -tags integration ./...
```

## Project layout

```
main.go, app.go        Wails entry point and the methods the UI calls
internal/
  winapi/              thin Win32 wrappers: windows, placement, processes, launch, open files
  state/               captures an app's windows into an AppState
  snapshot/            per-app timeline: ingest, change detection, persistence
  restore/             relaunch / reopen, match windows, put them back
  plugins/             app-specific capture and relaunch (browsers, Explorer, Office)
  events/              WinEvent hook: window moved / shown / focused
  hotkey/, tray/       configurable global shortcut and the tray icon
  settings/            validated preferences and atomic settings.json persistence
  policy/              what is tracked and for how long
  services/            wires everything together: capture loop, restore, tracking
frontend/src/
  App.svelte           picks the view and subscribes to backend events
  ui/                  small accessible controls, one folder per element
  features/            titlebar, apps, timeline, overlay and modal settings sections
  lib/                 event stores, backend actions, visibility and motion
  i18n/                English/Russian dictionaries and language store
  theme/               CSS tokens, reset and instant theme application
```

## Preferences and verification

Settings live in `%LOCALAPPDATA%\Rewinder\settings.json`; background images are copied to `background.png` in the same directory. The global shortcut, language, accent, corners, density, background and motion can be changed in the settings modal. The initial language follows the Windows UI language (Russian or English; other locales use English).

The quick overlay is a compact always-on-top window on the cursor monitor. **Open Rewinder** restores the main window geometry. Closing the main window hides it in the tray; Exit stops the app.

Run `./scripts/verify.ps1` for all required frontend and Go checks. On Windows this also runs the integration tests and `wails build` under Go 1.24.0. After a production build, `node scripts/ui-smoke.mjs` checks the UI in headless Edge with a mock Wails bridge and saves screenshots to `build/ui-review/`. This does not replace native hotkey/focus checks; see [manual checklist and stage reports](WORK_LOG.md).

Hidden views are unmounted and snapshot events accumulate until the window is shown. There are no frontend polling intervals; toast timeouts stop while hidden. Background blur is rasterized once per image/parameter change. Wails v2 cannot release the WebView2 process memory while hidden.

Icons are vendored SVG paths from [Phosphor Icons](https://github.com/phosphor-icons/core), under the included MIT license.

## Troubleshooting

The log is at `%LOCALAPPDATA%\Rewinder\rewinder.log`. Please attach it to an [issue](../../issues).

## License

MIT

---

**Author**: lovly · **Version**: 0.2.0
