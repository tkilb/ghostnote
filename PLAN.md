# Ghostnote: Architecture & Implementation Plan

Ghostnote is a general-purpose terminal music scraper and tagger built in Go. It pairs a high-efficiency `lazygit`-inspired terminal user interface (TUI) with a robust scraping and post-processing pipeline powered by `yt-dlp` and `ffmpeg`.

---

## 1. Project Goals & Architecture Overview

### Goals
- **Keyboard-Driven Terminal Experience**: High-density, responsive TUI designed for fast workflows using arrow-key navigation (leaving `hjkl` free for custom/feature mappings).
- **Multi-Source Support**: Seamless audio scraping from YouTube, YouTube Music, SoundCloud, Bandcamp, and direct media URLs.
- **Stitched OST / Mix Slicing**: First-class handling of long video mixes/OSTs with automatic chapter extraction and track-by-track queue expansion.
- **Clean Tagging & Naming**: Automatic heuristic stripping of video title noise, embedded album artwork, and flat file naming.
- **Visual Consistency**: Exact UI palette, borders, status bars, and modal conventions adapted from `~/Git/lazytask`.

### Technology Stack
- **Language**: Go (1.24+)
- **TUI Framework**:
  - [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea) (Application runtime & state loop)
  - [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss) (Layouts, borders, styling)
  - [`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles) (Text inputs, spinners, progress bars, viewports)
- **Scraper / Media Backend**:
  - `yt-dlp`: Stream extraction, metadata discovery, chapter detection.
  - `ffmpeg`: Audio transcoding, chapter slicing, metadata/ID3 tagging, thumbnail embedding.
- **System Integrations**:
  - `atotto/clipboard` or OSC 52 for clipboard paste auto-detection.
  - `gopkg.in/yaml.v3` for persistent user configuration.

---

## 2. TUI Design & Layout

Ghostnote replicates the visual language of `lazytask`: rounded borders, bold titles embedded into top border lines, status indicators, and modal popups.

### Palette & Styling (Matching `lazytask`)
- **Focused Panel Border / Title**: ANSI Green (`"2"`)
- **Unfocused Panel Border / Title**: Terminal Default (`""`)
- **Active Selection / Row Highlight**: ANSI Blue (`"4"`)
- **Active Tab Highlight**: Bold Blue (`"4"`)
- **Errors / High Priority**: ANSI Red (`"1"`)
- **Warnings / In-Progress**: ANSI Yellow (`"3"`)
- **Informational / Low Priority**: ANSI Cyan (`"6"`)

### Layout Grid

```
┌───────────────────────────┬──────────────────────────────────────────┐
│ [1]-Config                │ [0]-Details                              │
│ Dir: ~/Music              │ Title:  Corridors of Time                │
│ Format: MP3 (320k)        │ Artist: Yasunori Mitsuda                 │
│ Pattern: {artist} - {titl…│ Album:  Chrono Trigger OST               │
├───────────────────────────┤ Track:  14/25   Year: 1995               │
│ [2]-Queue (3 active)      │ Source: https://youtube.com/watch?v=...  │
│ > [14] Corridors of Time  │ Art:    [Embedded Thumbnail]             │
│   [15] Schala's Theme     │ Output: ~/Music/Yasunori Mitsuda - Co... │
│   [16] Zeals Kingdom      │ Status: Slicing (72%)                    │
├───────────────────────────┤                                          │
│ [3]-Completed (13)        │ [e] Edit Metadata  [d] Delete File       │
│   [11] Wind Scene         │                                          │
│   [12] Frog's Theme       │                                          │
└───────────────────────────┴──────────────────────────────────────────┘
 [1-3,0/Tab] panels  [↑/↓] nav  [a] add  [d] dest  [f] format  [?] help  [q] quit
```

### Panels & Responsibilities
- **`[1] Config`**: Quick summary of current destination directory, audio format/codec, and active naming template. Pressing `d` opens the folder picker modal; `f` toggles audio format.
- **`[2] Queue`**: Active and pending downloads. Each item shows title, duration/chapter index, and live progress bar (`%`, speed, ETA, and processing phase).
- **`[3] Completed`**: History of successfully processed tracks in the current session/destination.
- **`[0] Details`**: Deep inspector for the currently focused track. Displays full tag metadata, output path, source URL, and cover art status. Pressing `e` opens an inline metadata editor.

### Keybinding Architecture
*Note: Arrow keys are primary navigation keys, leaving `h`, `j`, `k`, and `l` unreserved.*

- **Global Navigation**:
  - `0`, `1`, `2`, `3`: Jump directly to panel.
  - `Tab` / `Shift+Tab`: Cycle through panels.
  - `↑` / `↓`: Navigate items in lists (Queue, Completed, Forms).
  - `Enter`: Select / confirm / open detail.
  - `Esc`: Close modal popup / cancel prompt.
  - `?`: Toggle help popup (displaying context-sensitive key reference).
  - `q` / `Ctrl+C`: Quit application.
- **Global Actions**:
  - `a`: Add URL (opens input modal, pre-populated with clipboard if valid URL).
  - `d`: Open interactive destination directory picker modal.
  - `f`: Toggle audio format (`MP3` -> `FLAC` -> `Opus` -> `Source-Copy`).
- **Queue Actions (when `[2] Queue` focused)**:
  - `x`: Cancel / remove download.
  - `r`: Retry failed item.
  - `c`: Clear completed items from queue.
  - `Ctrl+↑` / `Ctrl+↓`: Re-order queue priority.
- **Details Actions (when `[0] Details` focused)**:
  - `e`: Edit metadata (opens form to edit Title, Artist, Album, Track #).

---

## 3. Core Engine & Scraping Pipeline

### Source Extraction (`yt-dlp`)
1. **URL Inspection**: Fetch JSON metadata using `yt-dlp --dump-single-json --flat-playlist <URL>` without downloading audio streams yet.
2. **Type Detection**:
   - **Single Track**: Directly queues 1 item.
   - **Standard Playlist**: Expands into $N$ individual track queue items.
   - **Chaptered Video (Stitched OST / Mix)**: Detects embedded chapters in JSON metadata.
3. **Execution**: Spawns `yt-dlp` asynchronously via `os/exec` with streaming stdout/stderr parsing for progress bars.

### Stitched OST & Mix Handling
When a stitched YouTube video contains chapters:
1. **Queue Expansion**: Ghostnote creates a parent download job and expands $N$ individual child queue items representing each chapter.
2. **Audio Extraction**: The full high-quality audio stream is downloaded once to a temporary cache.
3. **Lossless Slicing**: `ffmpeg` slices the cached stream using chapter start/end timestamps (`-ss`, `-to`, `-c copy` or transcode target).
4. **Independent Tagging**: Each slice receives its specific chapter title, calculated track number, overall album name, and embedded cover art.

### Metadata Processing & Tagging
- **Heuristic Cleaner**: Strips common video title noise:
  - Tokens: `(Official Music Video)`, `(Official Audio)`, `[HD]`, `4K`, `(Lyric Video)`, `(Audio)`.
  - Pattern Match: Splits `"Artist - Title"` strings automatically.
- **Tag Embedding**: `ffmpeg` writes metadata tags directly to output files:
  - MP3: ID3v2.3 / ID3v2.4 (`TIT2`, `TPE1`, `TALB`, `TRCK`, `TYER`, `APIC`).
  - FLAC / Opus: Vorbis comments (`TITLE`, `ARTIST`, `ALBUM`, `TRACKNUMBER`, `DATE`, `METADATA_BLOCK_PICTURE`).
- **Cover Art**: Thumbnail image is retrieved during scraping and embedded into the file container.

### Destination & File Naming
- **Flat File Hierarchy**: Files are placed directly into the selected destination folder:
  - Default Pattern: `{artist} - {title}.{ext}`
  - Track-numbered Pattern (for OSTs): `{track:02d} - {title}.{ext}`
- **Sanitization**: Strips forbidden filesystem characters (`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`) and collapses redundant whitespace.
- **Duplicate Detection**: Inspects destination directory before downloading. If a file with the target filename already exists, the item is marked as `[Existing]` with an option to skip or overwrite.

---

## 4. Configuration & Persistence

Configuration is persisted to `~/.config/ghostnote/config.yaml`:

```yaml
version: 1
destination: "~/Music"
audio:
  format: "mp3"          # mp3, flac, opus, copy
  bitrate: "320k"        # for mp3/opus
naming:
  pattern: "{artist} - {title}"
  ost_pattern: "{track:02d} - {title}"
  strip_noise: true
queue:
  max_concurrent: 3
  clipboard_quick_add: true
```

---

## 5. Idiomatic Go Project Structure

```
ghostnote/
├── cmd/
│   └── ghostnote/
│       └── main.go                 # Entrypoint, Bubble Tea program initialization
├── internal/
│   ├── config/
│   │   ├── config.go               # YAML configuration loading, validation, saving
│   │   └── config_test.go
│   ├── engine/
│   │   ├── ytdlp.go                # yt-dlp process runner, JSON parser, progress scanner
│   │   ├── ffmpeg.go               # ffmpeg slicer, transcoder, tag embedder
│   │   ├── queue.go                # Concurrent worker pool and job coordinator
│   │   └── types.go                # Job, Track, Status, Chapter data structures
│   ├── metadata/
│   │   ├── clean.go                # Heuristic regex cleaners for YouTube titles
│   │   ├── tagger.go               # Tag specification and validation
│   │   └── clean_test.go           # Table-driven unit tests for title cleaning
│   └── ui/
│       ├── app.go                  # Root Bubble Tea model, grid layout, key router
│       ├── panel/
│       │   └── panel.go            # Border, frame, title, tabs, and footer styling (lazytask style)
│       ├── configpanel/
│       │   └── configpanel.go      # [1] Config panel view & model
│       ├── queuepanel/
│       │   └── queuepanel.go       # [2] Queue panel view & model with progress bars
│       ├── completedpanel/
│       │   └── completedpanel.go   # [3] Completed history panel view & model
│       ├── detailspanel/
│       │   └── detailspanel.go     # [0] Details & tag inspector/editor view & model
│       ├── dirpicker/
│       │   └── dirpicker.go        # Interactive filesystem directory navigation modal
│       ├── addmodal/
│       │   └── addmodal.go         # URL input modal with clipboard prefill
│       ├── helppopup/
│       │   └── helppopup.go        # Context-sensitive '?' keybinding modal
│       └── statusbar/
│           └── statusbar.go        # Bottom key hints status bar
├── go.mod
├── go.sum
└── PLAN.md
```

---

## 6. Implementation Milestones

### Phase 1: Engine & CLI Backend Foundation
- [x] Initialize Go module with `bubbletea`, `lipgloss`, `bubbles`, and `yaml.v3`.
- [x] Implement `internal/config` with defaults and disk persistence.
- [x] Implement `internal/metadata` heuristic title cleaning with comprehensive table-driven tests.
- [x] Implement `internal/engine`:
  - `yt-dlp` metadata extraction & chapter parsing.
  - Download runner with real-time progress parsing.
  - `ffmpeg` chapter slicing, transcoding, and cover-art embedding.
  - Queue coordinator with worker pool and live notification listener.

### Phase 2: TUI Components & Visual Shell
- [x] Port `internal/ui/panel` matching lazytask/lazygit (rounded borders, active/inactive palette).
- [x] Implement root `app.go` with responsive grid layout (`WindowSizeMsg` handling).
- [x] Build `[1] Config`, `[2] Queue`, `[3] Completed`, and `[0] Details` panel shells.
- [x] Implement `statusbar` and `helppopup` modals with arrow-key navigation.
- [x] Wire panel jumping (`0-3`, `Tab`, `Shift+Tab`) and item cursor movement (`↑`/`↓`).

### Phase 3: Interactive Modals & Workflow Wiring
- [x] Build `addmodal`: URL text input, Enter to submit, auto-populate from system clipboard.
- [x] Build `dirpicker`: In-TUI interactive directory browser (`↑`/`↓`, `Enter`/`Space`/`Esc`).
- [x] Connect `addmodal` to `engine.Queue`:
  - Detect single track vs chaptered OST.
  - Expand OST chapters into individual child queue items.
- [x] Wire live download progress, slicing status, and completion state into `[2] Queue`.
- [x] Implement `[0] Details` inline metadata editing modal (`e` key) to modify tags before/after slicing.

### Phase 4: Polish & Refinement
- [ ] Duplicate detection (scan target directory prior to download).
- [ ] Queue actions: retry (`r`), cancel (`x`), clear completed (`c`).
- [ ] Audio format quick-toggle (`f` key: MP3 / FLAC / Opus / Copy).
- [ ] Error handling popups for missing `yt-dlp` / `ffmpeg` dependencies.

---

## 7. Future Enhancements (Tabled)

- **Interactive In-App Search & Discovery**:
  - Direct search panel querying `ytsearch:` or SoundCloud API directly within Ghostnote.
  - Browse search results, preview duration/uploader, and press Space/Enter to queue.
- **External Metadata Enrichment**:
  - Asynchronous MusicBrainz / Discogs lookups for high-confidence album matching.
- **Batch Import**:
  - Import batches from local text files, M3U playlists, or browser bookmarks.
