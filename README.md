# <img src="internal/public/icon.ico" width="32"> Rewinder — Time Machine for Windows windows

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
wails build
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

## Troubleshooting

The log is at `%LOCALAPPDATA%\Rewinder\rewinder.log`. Please attach it to an [issue](../../issues).

## License

MIT

---

**Author**: lovly · **Version**: 0.2.0
