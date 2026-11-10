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
- **Undo the undo, the right way.** A state that came from a restore is marked, so the next `Ctrl + Alt + Z` goes further back instead of flipping back and forth.
- **Preview before you restore**: a minimap of the monitors and windows of every snapshot.
- **Named layouts**: save where all open apps are ("Work", "Stream") and put them all back with one click, from the tray, or with `Ctrl + Alt + 1…9`.
- **Unplugged monitor? No lost windows.** A window saved on a monitor that is gone comes back onto the primary one.
- **Exclusions**: stop tracking any app and delete its history.
- **History survives a reboot**: 7 days, up to 500 snapshots per app.
- **No administrator rights needed.** To restore apps that run as administrator, turn on *Run as administrator* (UAC asks once; autostart then uses a Task Scheduler task).
- **Updates itself** from GitHub Releases: checks after start and every 6 hours, verifies the SHA-256 checksum (and the signature when releases are signed), swaps the exe and restarts quietly. Can be set to *ask first* or *off*.

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

## CI and releases

- **CI** (`.github/workflows/ci.yml`, every push and pull request, Windows runner): gofmt, `go vet`, unit tests, `svelte-check`, frontend build, a check that `icons.ts` matches the pinned Phosphor version, integration tests on the runner's desktop (reported, not blocking) and a `Rewinder.exe` artifact.
- **Release** (`.github/workflows/release.yml`): push a tag `vMAJOR.MINOR.PATCH` (a `-beta.1` suffix makes a pre-release). It stamps the version into the exe, builds `Rewinder.exe` and the NSIS installer, signs both when a certificate is configured, writes `SHA256SUMS.txt` and publishes the GitHub Release. The built-in updater reads exactly that release.

  ```bash
  git tag v0.3.0
  git push origin v0.3.0
  ```
- **Code signing** is optional. Add two repository secrets: `WINDOWS_CERT_PFX` (the `.pfx` certificate, base64) and `WINDOWS_CERT_PASSWORD`. Once releases are signed, the updater refuses any later build that is not signed. Without a certificate Windows SmartScreen warns on first start; free signing for open-source projects is available from [SignPath Foundation](https://signpath.org).
- **Dependabot** keeps Go modules, npm packages and the actions up to date (Wails stays pinned at 2.11).

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
  ui/                  one folder per control: button, select, switch, slider, segmented,
                       dialog, tooltip, toast, kbd, avatar, search, swatch, empty, icon
  features/            titlebar, apps, timeline, overlay, settings (one folder per section)
  lib/                 stores and backend calls: apps, restore, preferences, hotkey, format
  i18n/                English/Russian dictionaries and language store
  theme/               tokens, fonts, reset, and applying appearance settings live
scripts/icons.mjs      extracts outline + fill Phosphor icons into ui/icon/icons.ts
```

The UI is Svelte 5 with [bits-ui](https://bits-ui.com) for dropdowns, sliders, switches,
tooltips and the settings dialog. Animations use only `transform` and `opacity`, because
WebView2 renders on the CPU here (its GPU process is turned off to save ~60 MB).

## Preferences and verification

Settings live in `%LOCALAPPDATA%\Rewinder\settings.json`; background images are copied to `background.png` in the same directory. The global shortcut, language, accent, corners, density, background and motion can be changed in the settings modal. The initial language follows the Windows UI language (Russian or English; other locales use English).

The quick overlay is a compact always-on-top window on the cursor monitor. **Open Rewinder** restores the main window geometry. Closing the main window hides it in the tray; Exit stops the app.

Run `./scripts/verify.ps1` for all required frontend and Go checks. On Windows this also runs the integration tests and `wails build` under Go 1.24.0. Native hotkey and focus behaviour still needs a manual check; see [manual checklist and stage reports](WORK_LOG.md).

Hidden views are unmounted and snapshot events accumulate until the window is shown. There are no frontend polling intervals; toast timeouts stop while hidden. Background blur is rasterized once per image/parameter change. Wails v2 cannot release the WebView2 process memory while hidden.

Icons are vendored SVG paths from [Phosphor Icons](https://github.com/phosphor-icons/core), under the included MIT license.

## Troubleshooting

The log is at `%LOCALAPPDATA%\Rewinder\rewinder.log`. Please attach it to an [issue](../../issues).

## License

MIT

---

**Author**: lovly · **Version**: 0.2.0
