<img width="500" height="705" alt="image" src="https://github.com/user-attachments/assets/fdc0e019-897f-4bde-a4be-889fa0dceccc" />



# JPG Shamsi Renamer

A tiny portable Windows desktop app that renames .jpg files by appending their Shamsi (Jalali) capture date, read from the image's EXIF metadata — and optionally prepending the camera model.

```
IMG_0003.jpg → IMG_0003-بهمن-1398.jpg
IMG_0003.jpg → NIKON D7200_IMG_0003-1398-11-06.jpg
```

[![Go](https://github.com/imanghafoori1/shamsi-image-renamer/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/imanghafoori1/shamsi-image-renamer/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/imanghafoori1/shamsi-image-renamer/branch/main/graph/badge.svg)](https://codecov.io/gh/imanghafoori1/shamsi-image-renamer)

Written in Go with a native Win32 GUI (lxn/walk). No installer, no runtime, no network access. One .exe, a few MB.

### Why
Camera files come off the SD card as DSC_9340.JPG, IMG_0003.JPG, P1000123.JPG — meaningless once they leave the camera. Sorting by modification time is unreliable (copying, editing, and cloud sync all rewrite it), and the capture date is the only date that matters when browsing a photo library.

This tool reads DateTimeOriginal directly from EXIF, converts it to the Persian calendar, and encodes the result in the filename so the date survives any future move, copy, or backup.

## Features

- **Shamsi date suffix** from EXIF `DateTimeOriginal` — the true capture time, not the file's mtime.
- **Five date formats** (see below).
- **Optional camera-model prefix**, read from EXIF `Model`.
- **Idempotent.** Re-running on an already-processed folder is a no-op — the tool recognises its own previous output and strips it before rebuilding, so you never get `IMG_0003-بهمن-1398-بهمن-1398.jpg`.
- **Repairs old broken names.** Files left behind by earlier builds (`NIKOND7200_NIKOND7200_DSC_9340-بهمن-1398-1398-bahman-6`) are cleaned back to a single canonical form.
- **Skips safely.** Files with no EXIF, unreadable EXIF, or a name collision are left untouched and reported.
- **Inline report.** After each run, the window shows the count and two scrollable lists: renamed files and skipped files (with the reason for each skip).
- **No network, no telemetry, no writes outside the chosen folder.**

## Date formats
| Dropdown label | Example suffix |
|---|---|
| `1398-11-28  (year-month-day)` | `-1398-11-28` |
| `1398-bahman-28  (year-monthname-day)` | `-1398-bahman-28` |
| `1398-bahman  (year-monthname)` | `-1398-bahman` |
| `بهمن-1398  (monthname-year, Persian)` | `-بهمن-1398` |
| `بهمن  (monthname only, Persian)` | `-بهمن` |
The first option is the default. Format 1 uses zero-padded month/day, so filenames sort correctly.

## Camera model handling
When Prepend camera model is checked:

| EXIF `Model` | Prefix |
|---|---|
| `NIKON D7200` | `NIKON D7200_` (kept verbatim, space preserved) |
| `Canon EOS 5D` | `CanonEOS5D_` |
| `Canon EOS 5D Mark IV` | `CanonEOS5DMarkIV_` |

Rule in one sentence: exact NIKON D<digits> models keep their original spelling; everything else is sanitized (spaces and illegal characters removed) and then NIKONDâ€¦ is rewritten as NIKON_Dâ€¦.

## Getting it
### Pre-built binary
Download the latest renamer.exe from the Releases page and double-click it. Nothing to install.

> The binary is unsigned, so Windows SmartScreen will show a "Windows protected your PC" warning on first run. Click More info â†’ Run anyway. If you'd rather not trust a random binary, build it yourself (below) — it's a small, readable project.

### From source
Requires:

- Go 1.26 or newer (download)

- Windows (the GUI library is Windows-only)

```powershell
git clone https://github.com/<your-username>/renamer.git
cd renamer
go mod download
go build -ldflags="-H windowsgui" -o renamer.exe
```

The -H windowsgui flag suppresses the console window; without it, a black command prompt appears behind the GUI.

One extra step on a fresh clone: the Walk library requires a Windows application manifest to use Common Controls 6; otherwise, the window renders with the old, ugly widget style — or fails to appear at all.

```powershell
go install github.com/akavel/rsrc@latest
rsrc -manifest test.manifest -o rsrc.syso
go build -ldflags="-H windowsgui" -o renamer.exe
```

rsrc.syso is generated from test.manifest and is gitignored. Regenerate it whenever you clean the build directory.

---

### Usage
1. Launch renamer.exe.

2. Click Browseâ€¦ and pick a folder of JPGs, or paste a path (D:\Photos\Trip2025, \\NAS\Photos\2025, â€¦).

3. (Optional) Tick Prepend camera model.

4. Pick a date format.

5. Click Rename Files.

6. Read the report at the bottom of the window:
```
Renamed: 42 Skipped: 3
```
The two list boxes show every change and every skip with its reason.


### What "skipped" means

| Reason |	Meaning |
|---|---|
| no EXIF data |	The file is a JPG but has no EXIF block at all (e.g. downloaded, screenshotted, or stripped by a messaging app). |
| no DateTimeOriginal |	Has EXIF but not the capture-time tag specifically. |
| already renamed |	Filename already matches the format you selected. Nothing to do. |
| target already exists |	Renaming would overwrite a different file. Left alone. |
| rename failed |	OS-level failure (permission, locked file, etc.). |

### Safety notes
Test on a copy first. Copy 3-4 photos into a scratch folder, run the tool, and confirm the output names look right. Then point it at your real library.

Renames are not recorded to a log. The report on screen is the only record. If you want a permanent trace, run the tool once per folder and screenshot the report, or use a version-control tool like git on the photo folder (with large-file storage configured), or keep a checksum manifest with sha256sum before and after.

Only .jpg / .jpeg files are touched. Other extensions are ignored silently.

The tool never deletes or moves files between folders. Everything stays in the folder you selected.

No network access. The binary makes no outbound connections.

---

### Building and testing

```powershell

Run all tests
go test ./...

Verbose test output
go test ./... -v

Build for release
go build -ldflags="-H windowsgui" -o renamer.exe
```

Tests generate their own JPEG+EXIF fixtures at runtime (no binary fixtures committed) and clean up via Go's t.TempDir(). The test suite covers:

- All five date formats, byte-for-byte

- Month-name tables (Latin and Persian)

- Model sanitization and normalization

- The decoration-stripping logic, including the exact DSC_9340-بهمن-1398-1398-bahman-6-… accumulated-suffix bug

- End-to-end rename, idempotence, format switching, checkbox toggling, collision handling, and no-EXIF skip paths

---


The ui_* files each build one visual row and hold no state of their own — shared widget references live in the uiState struct in ui.go. Everything outside the ui_* files is plain Go with no GUI dependency, which is why the test suite can exercise the entire rename pipeline without a display.

---

### Limitations
Windows only. The GUI library (lxn/walk) targets Win32. The non-GUI code (EXIF reading, date conversion, filename logic) is portable, and the CLI-friendly processDir function could be reused in a cross-platform build.

Persian digits are not produced. Suffixes use ASCII digits (1398), not Persian-Indic numerals (۱۳۹۸). This keeps filenames portable across filesystems and searchable from any locale.

EXIF only. If the capture date has been stripped (many social and messaging apps do this), the tool cannot recover it.

No undo. See "Safety notes" above.

---

### Contributing
Issues and pull requests are welcome. Before sending a PR:

```powershell
go test ./... -v
```

should be green, and any behavior change should come with a test that fails on the old code and passes on the new. The test suite is the contract — it pins down every filename the tool produces for a given input, so a regression in formatting or in the decoration-stripping logic will be caught immediately.

---

### License
MIT — see LICENSE.

---

### Acknowledgments
- lxn/walk — native Win32 GUI for Go

- rwcarlsen/goexif — EXIF parsing

- jalaali/go-jalaali — Gregorian â†” Jalali conversion
