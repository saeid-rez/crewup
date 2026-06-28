package ui

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ErrUserCancelled indicates the user aborted an interactive selection.
var ErrUserCancelled = errors.New("user cancelled setup")

type item struct {
	id       string
	title    string
	desc     string
	selected bool
}

type multiSelectModel struct {
	title    string
	items    []item
	cursor   int
	done     bool
	quitting bool
}

var (
	multiSelectTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	multiSelectCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	multiSelectSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	multiSelectDescStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	multiSelectHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func newMultiSelectModel(title string, items []item) *multiSelectModel {
	cloned := make([]item, len(items))
	copy(cloned, items)
	return &multiSelectModel{title: title, items: cloned}
}

func (m *multiSelectModel) Init() tea.Cmd { return nil }

func (m *multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "space":
			m.toggleCursor()
		case "a":
			m.toggleAll()
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *multiSelectModel) View() tea.View {
	var sb strings.Builder
	sb.WriteString(multiSelectTitleStyle.Render(m.title))
	sb.WriteString("\n\n")

	for i, it := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = multiSelectCursorStyle.Render("> ")
		}

		marker := "○"
		lineStyle := lipgloss.NewStyle()
		if it.selected {
			marker = multiSelectSelectedStyle.Render("●")
			lineStyle = multiSelectSelectedStyle
		}

		sb.WriteString(cursor)
		sb.WriteString(marker)
		sb.WriteString(" ")
		sb.WriteString(lineStyle.Render(it.title))
		sb.WriteString("\n")

		if it.desc != "" {
			sb.WriteString("    ")
			sb.WriteString(multiSelectDescStyle.Render(it.desc))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(multiSelectHelpStyle.Render("up/down or j/k: move • space: toggle • a: all • enter: confirm • q: cancel"))
	sb.WriteString("\n")
	return tea.NewView(sb.String())
}

func (m *multiSelectModel) toggleCursor() {
	if len(m.items) == 0 {
		return
	}
	m.items[m.cursor].selected = !m.items[m.cursor].selected
}

func (m *multiSelectModel) toggleAll() {
	if len(m.items) == 0 {
		return
	}

	selectAll := !m.allSelected()
	for i := range m.items {
		m.items[i].selected = selectAll
	}
}

func (m *multiSelectModel) allSelected() bool {
	if len(m.items) == 0 {
		return false
	}
	for _, it := range m.items {
		if !it.selected {
			return false
		}
	}
	return true
}

func (m *multiSelectModel) selectedIDs() []string {
	ids := make([]string, 0, len(m.items))
	for _, it := range m.items {
		if it.selected {
			ids = append(ids, it.id)
		}
	}
	return ids
}

func (m *multiSelectModel) result() ([]string, error) {
	if m.quitting && !m.done {
		return nil, ErrUserCancelled
	}
	return m.selectedIDs(), nil
}

// RunMultiSelect runs a Bubble Tea multi-select prompt and returns selected IDs.
func RunMultiSelect(title string, items []item) ([]string, error) {
	if len(items) == 0 {
		return []string{}, nil
	}

	model := newMultiSelectModel(title, items)
	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, err
	}

	resultModel, ok := finalModel.(*multiSelectModel)
	if !ok {
		return nil, fmt.Errorf("unexpected Bubble Tea model type %T", finalModel)
	}

	return resultModel.result()
}
