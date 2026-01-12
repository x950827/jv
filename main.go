package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current view mode
type Mode int

const (
	InputMode Mode = iota
	ViewMode
)

// Selection represents what part of the node is selected
type Selection int

const (
	SelectNone Selection = iota
	SelectKey
	SelectValue
)

// Styles using Lip Gloss
var (
	keyStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Bold(true)
	stringStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A"))
	numberStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9E64"))
	boolStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#BB9AF7"))
	nullStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Italic(true)
	bracketStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#A9B1D6"))
	collapsedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89"))
	cursorStyle       = lipgloss.NewStyle().Background(lipgloss.Color("#3D59A1")).Foreground(lipgloss.Color("#C0CAF5"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")).Padding(0, 1)
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Bold(true).Padding(0, 1)
	borderStyle       = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#3D59A1"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")).Bold(true)
	hintStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Italic(true)
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Bold(true).Padding(0, 1)
	selectedKeyStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#7AA2F7")).Foreground(lipgloss.Color("#1A1B26")).Bold(true)
	selectedValStyle  = lipgloss.NewStyle().Background(lipgloss.Color("#9ECE6A")).Foreground(lipgloss.Color("#1A1B26"))
)

// Node represents a JSON node in the tree
type Node struct {
	Key       string
	Value     interface{}
	Type      string // "object", "array", "string", "number", "bool", "null"
	Children  []*Node
	Collapsed bool
	Depth     int
	Parent    *Node
	Index     int // For array elements
}

// clearStatusMsg is a message to clear the status
type clearStatusMsg struct{}

// Model is the Bubble Tea model
type Model struct {
	mode          Mode
	root          *Node
	flatNodes     []*Node // Flattened visible nodes
	cursor        int
	selection     Selection
	viewport      viewport.Model
	textarea      textarea.Model
	ready         bool
	width         int
	height        int
	jsonInput     string
	parseError    error
	statusMessage string
}

func main() {
	var m Model

	// Check if there's input from file or stdin
	if len(os.Args) > 1 {
		// Read from file
		input, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		m = NewModelWithJSON(string(input))
	} else {
		// Check if there's data on stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			m = NewModelWithJSON(string(input))
		} else {
			// No input - start in input mode
			m = NewModelForInput()
		}
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// NewModelForInput creates a new model in input mode
func NewModelForInput() Model {
	ta := textarea.New()
	ta.Placeholder = `Paste your JSON here...

Example:
{
  "name": "John",
  "age": 30
}`
	ta.Focus()
	ta.CharLimit = 0 // No limit
	ta.ShowLineNumbers = true

	return Model{
		mode:     InputMode,
		textarea: ta,
	}
}

// NewModelWithJSON creates a new model with the given JSON input
func NewModelWithJSON(jsonInput string) Model {
	m := Model{
		mode:      ViewMode,
		jsonInput: jsonInput,
	}
	m.parseJSON()
	return m
}

func (m *Model) parseJSON() {
	m.parseError = nil
	var data interface{}
	if err := json.Unmarshal([]byte(m.jsonInput), &data); err != nil {
		m.parseError = err
		return
	}

	m.root = buildTree(data, "", 0, nil, -1)
	m.flattenNodes()
}

func buildTree(data interface{}, key string, depth int, parent *Node, index int) *Node {
	node := &Node{
		Key:    key,
		Value:  data,
		Depth:  depth,
		Parent: parent,
		Index:  index,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		node.Type = "object"
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := buildTree(v[k], k, depth+1, node, -1)
			node.Children = append(node.Children, child)
		}
	case []interface{}:
		node.Type = "array"
		for i, item := range v {
			child := buildTree(item, "", depth+1, node, i)
			node.Children = append(node.Children, child)
		}
	case string:
		node.Type = "string"
	case float64:
		node.Type = "number"
	case bool:
		node.Type = "bool"
	case nil:
		node.Type = "null"
	}

	return node
}

func (m *Model) flattenNodes() {
	m.flatNodes = nil
	if m.root != nil {
		m.flattenNode(m.root)
	}
}

func (m *Model) flattenNode(node *Node) {
	m.flatNodes = append(m.flatNodes, node)
	if !node.Collapsed {
		for _, child := range node.Children {
			m.flattenNode(child)
		}
	}
}

func (m Model) Init() tea.Cmd {
	if m.mode == InputMode {
		return textarea.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMessage = ""
		m.selection = SelectNone
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.mode == InputMode {
			m.textarea.SetWidth(m.width - 4)
			m.textarea.SetHeight(m.height - 6)
		}

		headerHeight := 2
		footerHeight := 2
		contentHeight := m.height - headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(m.width-2, contentHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = m.width - 2
			m.viewport.Height = contentHeight
		}
	}

	if m.mode == InputMode {
		return m.updateInputMode(msg)
	}
	return m.updateViewMode(msg)
}

func (m Model) updateInputMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+d", "f5":
			// Parse and switch to view mode
			m.jsonInput = m.textarea.Value()
			if strings.TrimSpace(m.jsonInput) == "" {
				return m, nil
			}
			m.parseJSON()
			if m.parseError == nil {
				m.mode = ViewMode
				m.cursor = 0
			}
			return m, nil
		case "esc":
			// Clear error and continue editing
			m.parseError = nil
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) copyToClipboard(text, what string) tea.Cmd {
	if err := clipboard.WriteAll(text); err != nil {
		m.statusMessage = fmt.Sprintf("Failed to copy: %v", err)
	} else {
		// Truncate display if too long
		display := text
		if len(display) > 50 {
			display = display[:47] + "..."
		}
		m.statusMessage = fmt.Sprintf("Copied %s: %s", what, display)
	}
	return clearStatusAfter(2 * time.Second)
}

func (m *Model) getNodeValueAsString(node *Node) string {
	switch node.Type {
	case "string":
		return node.Value.(string)
	case "number":
		return fmt.Sprintf("%v", node.Value)
	case "bool":
		return fmt.Sprintf("%v", node.Value)
	case "null":
		return "null"
	case "object", "array":
		// Serialize the whole subtree as JSON
		b, err := json.MarshalIndent(node.Value, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", node.Value)
		}
		return string(b)
	}
	return ""
}

func (m Model) updateViewMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "i":
			// Go back to input mode
			m.mode = InputMode
			m.textarea.SetValue(m.jsonInput)
			m.textarea.Focus()
			return m, textarea.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.selection = SelectNone
				m.ensureCursorVisible()
			}
		case "down", "j":
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
				m.selection = SelectNone
				m.ensureCursorVisible()
			}
		case "enter", " ", "right":
			// Toggle collapse
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if len(node.Children) > 0 {
					node.Collapsed = !node.Collapsed
					m.flattenNodes()
					// Ensure cursor doesn't go out of bounds
					if m.cursor >= len(m.flatNodes) {
						m.cursor = len(m.flatNodes) - 1
					}
				}
			}
		case "h", "left":
			// Collapse current or go to parent
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if len(node.Children) > 0 && !node.Collapsed {
					node.Collapsed = true
					m.flattenNodes()
				} else if node.Parent != nil {
					// Find parent in flat nodes
					for i, n := range m.flatNodes {
						if n == node.Parent {
							m.cursor = i
							break
						}
					}
				}
			}
		case "tab":
			// Toggle between key and value selection
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if node.Key != "" || node.Index >= 0 {
					// Has a key/index, can toggle
					if m.selection == SelectKey {
						m.selection = SelectValue
					} else if m.selection == SelectValue {
						m.selection = SelectNone
					} else {
						m.selection = SelectKey
					}
				} else {
					// No key, toggle between value and none
					if m.selection == SelectValue {
						m.selection = SelectNone
					} else {
						m.selection = SelectValue
					}
				}
			}
		case "y":
			// Copy value
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				m.selection = SelectValue
				value := m.getNodeValueAsString(node)
				cmd = m.copyToClipboard(value, "value")
				return m, cmd
			}
		case "Y":
			// Copy key
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				if node.Key != "" {
					m.selection = SelectKey
					cmd = m.copyToClipboard(node.Key, "key")
					return m, cmd
				} else if node.Index >= 0 {
					m.selection = SelectKey
					cmd = m.copyToClipboard(fmt.Sprintf("%d", node.Index), "index")
					return m, cmd
				}
			}
		case "p":
			// Copy JSON path
			if m.cursor < len(m.flatNodes) {
				node := m.flatNodes[m.cursor]
				path := m.getJSONPath(node)
				cmd = m.copyToClipboard(path, "path")
				return m, cmd
			}
		case "e":
			// Expand all
			m.expandAll(m.root)
			m.flattenNodes()
		case "c":
			// Collapse all
			m.collapseAll(m.root)
			m.flattenNodes()
		case "g":
			// Go to top
			m.cursor = 0
			m.viewport.GotoTop()
		case "G":
			// Go to bottom
			m.cursor = len(m.flatNodes) - 1
			m.viewport.GotoBottom()
		case "pgup":
			m.cursor -= m.viewport.Height
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.ensureCursorVisible()
		case "pgdown":
			m.cursor += m.viewport.Height
			if m.cursor >= len(m.flatNodes) {
				m.cursor = len(m.flatNodes) - 1
			}
			m.ensureCursorVisible()
		}
	}

	// Update viewport content
	if m.ready {
		m.viewport.SetContent(m.renderContent())
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m *Model) getJSONPath(node *Node) string {
	var parts []string
	current := node
	for current != nil {
		if current.Key != "" {
			parts = append([]string{"." + current.Key}, parts...)
		} else if current.Index >= 0 {
			parts = append([]string{fmt.Sprintf("[%d]", current.Index)}, parts...)
		}
		current = current.Parent
	}
	if len(parts) == 0 {
		return "$"
	}
	path := "$" + strings.Join(parts, "")
	// Clean up leading dot after $
	path = strings.Replace(path, "$.", "$.", 1)
	return path
}

func (m *Model) ensureCursorVisible() {
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
}

func (m *Model) expandAll(node *Node) {
	if node == nil {
		return
	}
	node.Collapsed = false
	for _, child := range node.Children {
		m.expandAll(child)
	}
}

func (m *Model) collapseAll(node *Node) {
	if node == nil {
		return
	}
	if len(node.Children) > 0 {
		node.Collapsed = true
	}
	for _, child := range node.Children {
		m.collapseAll(child)
	}
}

func (m Model) View() string {
	if m.mode == InputMode {
		return m.viewInputMode()
	}
	return m.viewViewMode()
}

func (m Model) viewInputMode() string {
	// Header
	title := titleStyle.Render("JSON Viewer - Input Mode")
	header := lipgloss.JoinHorizontal(lipgloss.Top, title)

	// Error display
	var errorMsg string
	if m.parseError != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Parse Error: %v", m.parseError))
	}

	// Textarea with border
	textareaView := borderStyle.Width(m.width - 2).Render(m.textarea.View())

	// Help footer
	help := helpStyle.Render("Ctrl+D or F5: parse & view • Ctrl+C: quit")
	hint := hintStyle.Render("Paste your JSON and press Ctrl+D to view")

	var parts []string
	parts = append(parts, header)
	if errorMsg != "" {
		parts = append(parts, errorMsg)
	}
	parts = append(parts, textareaView, hint, help)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) viewViewMode() string {
	if m.parseError != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E")).
			Bold(true).
			Padding(2).
			Render(fmt.Sprintf("JSON Parse Error: %v\n\nPress 'i' to edit", m.parseError))
	}

	if !m.ready {
		return "Loading..."
	}

	// Header
	title := titleStyle.Render("JSON Viewer")
	info := helpStyle.Render(fmt.Sprintf("%d nodes", len(m.flatNodes)))

	// Status message or info
	var headerRight string
	if m.statusMessage != "" {
		headerRight = statusStyle.Render(m.statusMessage)
	} else {
		headerRight = info
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, title, strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(headerRight)-4)), headerRight)

	// Footer with help
	help := helpStyle.Render("↑↓/jk: nav • ←→: collapse • tab: select • y: copy value • Y: copy key • p: copy path • i: edit • q: quit")

	// Main content with border
	content := borderStyle.Width(m.width - 2).Render(m.viewport.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, content, help)
}

func (m Model) renderContent() string {
	var sb strings.Builder

	for i, node := range m.flatNodes {
		line := m.renderNode(node, i == m.cursor)
		if i == m.cursor && m.selection == SelectNone {
			// Apply cursor highlight only when nothing specific is selected
			line = cursorStyle.Render(line)
		}
		sb.WriteString(line)
		if i < len(m.flatNodes)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) renderNode(node *Node, isCursor bool) string {
	indent := strings.Repeat("  ", node.Depth)
	var line strings.Builder
	line.WriteString(indent)

	// Add collapse/expand indicator for containers
	if len(node.Children) > 0 {
		if node.Collapsed {
			line.WriteString(collapsedStyle.Render("▸ "))
		} else {
			line.WriteString(collapsedStyle.Render("▾ "))
		}
	} else {
		line.WriteString("  ")
	}

	// Render key (if present)
	if node.Key != "" {
		keyText := fmt.Sprintf(`"%s"`, node.Key)
		if isCursor && m.selection == SelectKey {
			line.WriteString(selectedKeyStyle.Render(keyText))
		} else {
			line.WriteString(keyStyle.Render(keyText))
		}
		line.WriteString(bracketStyle.Render(": "))
	} else if node.Index >= 0 {
		indexText := fmt.Sprintf("[%d] ", node.Index)
		if isCursor && m.selection == SelectKey {
			line.WriteString(selectedKeyStyle.Render(indexText))
		} else {
			line.WriteString(collapsedStyle.Render(indexText))
		}
	}

	// Render value
	valueStr := m.renderValue(node, isCursor && m.selection == SelectValue)
	line.WriteString(valueStr)

	// Pad line to fill width (for cursor highlighting)
	lineStr := line.String()
	visibleWidth := lipgloss.Width(lineStr)
	if visibleWidth < m.width-4 {
		lineStr += strings.Repeat(" ", m.width-4-visibleWidth)
	}

	return lineStr
}

func (m Model) renderValue(node *Node, highlight bool) string {
	style := selectedValStyle
	if !highlight {
		style = lipgloss.NewStyle()
	}

	switch node.Type {
	case "object":
		if node.Collapsed {
			text := fmt.Sprintf("{...%d items}", len(node.Children))
			if highlight {
				return style.Render(text)
			}
			return bracketStyle.Render("{") + collapsedStyle.Render(fmt.Sprintf("...%d items", len(node.Children))) + bracketStyle.Render("}")
		} else if len(node.Children) == 0 {
			if highlight {
				return style.Render("{}")
			}
			return bracketStyle.Render("{}")
		} else {
			if highlight {
				return style.Render("{")
			}
			return bracketStyle.Render("{")
		}
	case "array":
		if node.Collapsed {
			text := fmt.Sprintf("[...%d items]", len(node.Children))
			if highlight {
				return style.Render(text)
			}
			return bracketStyle.Render("[") + collapsedStyle.Render(fmt.Sprintf("...%d items", len(node.Children))) + bracketStyle.Render("]")
		} else if len(node.Children) == 0 {
			if highlight {
				return style.Render("[]")
			}
			return bracketStyle.Render("[]")
		} else {
			if highlight {
				return style.Render("[")
			}
			return bracketStyle.Render("[")
		}
	case "string":
		text := fmt.Sprintf(`"%s"`, node.Value.(string))
		if highlight {
			return style.Render(text)
		}
		return stringStyle.Render(text)
	case "number":
		text := fmt.Sprintf("%v", node.Value)
		if highlight {
			return style.Render(text)
		}
		return numberStyle.Render(text)
	case "bool":
		text := fmt.Sprintf("%v", node.Value)
		if highlight {
			return style.Render(text)
		}
		return boolStyle.Render(text)
	case "null":
		if highlight {
			return style.Render("null")
		}
		return nullStyle.Render("null")
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
