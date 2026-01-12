# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
go build -o jv

# Run (three modes)
./jv                    # Interactive input mode
./jv file.json          # From file
cat file.json | ./jv    # From pipe

# Lint/vet
go vet ./...

# Tidy dependencies
go mod tidy

# Release build (requires goreleaser)
goreleaser release --clean
```

## Architecture

This is a single-file TUI application (`main.go`) built with the Bubble Tea framework using Model-View-Update (MVU) architecture.

### Core Components

**Model** - Central state container with two modes:
- `InputMode`: Textarea for pasting JSON (uses `bubbles/textarea`)
- `ViewMode`: Tree viewer with navigation (uses `bubbles/viewport`)

**Node** - Tree structure representing parsed JSON. Each node has:
- Type (object/array/string/number/bool/null)
- Children for nested structures
- Parent pointer for navigation
- Collapsed state for expand/collapse

**Key flows:**
1. JSON input → `parseJSON()` → `buildTree()` recursively creates Node tree
2. Node tree → `flattenNodes()` → flat slice of visible nodes for rendering
3. Collapse/expand → re-flattens tree → viewport updates

### Styling

All styles defined as package-level `lipgloss.Style` variables using Tokyo Night color palette. Key colors:
- Keys: `#7AA2F7` (blue)
- Strings: `#9ECE6A` (green)
- Numbers: `#FF9E64` (orange)
- Booleans: `#BB9AF7` (purple)

### Clipboard

Uses `atotto/clipboard` for cross-platform copy. Status messages auto-clear after 2 seconds via `tea.Tick`.
