# jv

A beautiful terminal JSON viewer with syntax highlighting, collapsible nodes, and copy support. Built with [Charm](https://charm.sh) libraries for a delightful TUI experience.

## Features

- **Syntax highlighting** - Keys, strings, numbers, booleans, and null values in distinct colors (Tokyo Night theme)
- **Collapsible tree view** - Navigate and collapse/expand JSON nodes
- **Interactive input mode** - Paste JSON directly without piping
- **Copy to clipboard** - Copy keys, values, or JSON paths
- **Keyboard navigation** - Vim-style keybindings
- **Mouse support** - Scroll with mouse wheel
- **Cross-platform** - Works on macOS, Linux, and Windows

## Installation

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/x950827/jv/main/install.sh | sh
```

### Go Install

```bash
go install github.com/x950827/jv@latest
```

### Build from Source

```bash
git clone https://github.com/x950827/jv.git
cd jv
go build -o jv
sudo mv jv /usr/local/bin/
```

## Usage

### Interactive Mode

Run without arguments to open the input editor where you can paste JSON:

```bash
jv
```

### From File

```bash
jv data.json
```

### From Pipe

```bash
cat data.json | jv

# From curl
curl -s https://api.github.com/users/octocat | jv

# From other commands
docker inspect container_id | jv
kubectl get pods -o json | jv
```

## Keyboard Shortcuts

### Input Mode

| Key | Action |
|-----|--------|
| `Ctrl+D` / `F5` | Parse JSON and switch to view mode |
| `Ctrl+C` | Quit |

### View Mode

#### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `g` | Go to top |
| `G` | Go to bottom |
| `PgUp` / `PgDown` | Page up/down |

#### Expand/Collapse

| Key | Action |
|-----|--------|
| `→` / `l` / `Enter` / `Space` | Toggle expand/collapse |
| `←` / `h` | Collapse current or go to parent |
| `e` | Expand all nodes |
| `c` | Collapse all nodes |

#### Copy

| Key | Action |
|-----|--------|
| `Tab` | Toggle selection (key → value → none) |
| `y` | Copy value to clipboard |
| `Y` | Copy key to clipboard |
| `p` | Copy JSON path (e.g., `$.data.items[0].name`) |

#### Other

| Key | Action |
|-----|--------|
| `i` | Switch to input mode (edit JSON) |
| `q` / `Ctrl+C` | Quit |

## Libraries Used

This project is built with the excellent [Charm](https://charm.sh) ecosystem:

| Library | Purpose |
|---------|---------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Model-View-Update architecture) |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Style definitions and terminal layouts |
| [Bubbles](https://github.com/charmbracelet/bubbles) | TUI components (viewport, textarea) |

Other dependencies:

| Library | Purpose |
|---------|---------|
| [clipboard](https://github.com/atotto/clipboard) | Cross-platform clipboard support |

## Color Theme

jv uses a Tokyo Night inspired color palette:

| Element | Color |
|---------|-------|
| Keys | `#7AA2F7` (Blue) |
| Strings | `#9ECE6A` (Green) |
| Numbers | `#FF9E64` (Orange) |
| Booleans | `#BB9AF7` (Purple) |
| Null | `#565F89` (Gray, italic) |
| Brackets | `#A9B1D6` (Light gray) |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Charm](https://charm.sh) for the amazing TUI libraries
- [Tokyo Night](https://github.com/enkia/tokyo-night-vscode-theme) for color inspiration
