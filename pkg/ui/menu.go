package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type MenuOption struct {
	ID          string
	Title       string
	Description string
}

type menuModel struct {
	title       string
	description string
	options     []MenuOption
	cursor      int
	selected    string
	quitting    bool
}

var (
	menuTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	menuDescriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	menuCursorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	menuOptionStyle      = lipgloss.NewStyle().Bold(true)
	menuHelpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func newMenuModel(title, description string, options []MenuOption) *menuModel {
	cloned := make([]MenuOption, len(options))
	copy(cloned, options)
	return &menuModel{title: title, description: description, options: cloned}
}

func (m *menuModel) Init() tea.Cmd { return nil }

func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.options) > 0 {
				m.selected = m.options[m.cursor].ID
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *menuModel) View() tea.View {
	var sb strings.Builder
	sb.WriteString(menuTitleStyle.Render(m.title))
	sb.WriteString("\n")
	if m.description != "" {
		sb.WriteString(menuDescriptionStyle.Render(m.description))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for i, option := range m.options {
		cursor := "  "
		if m.cursor == i {
			cursor = menuCursorStyle.Render("> ")
		}
		sb.WriteString(cursor)
		sb.WriteString(menuOptionStyle.Render(option.Title))
		sb.WriteString("\n")
		if option.Description != "" {
			sb.WriteString("    ")
			sb.WriteString(multiSelectDescStyle.Render(option.Description))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(menuHelpStyle.Render("up/down or j/k: move • enter: select • q: cancel"))
	sb.WriteString("\n")
	return tea.NewView(sb.String())
}

func (m *menuModel) result() (string, error) {
	if m.quitting {
		return "", ErrUserCancelled
	}
	if m.selected == "" {
		return "", fmt.Errorf("no menu option selected")
	}
	return m.selected, nil
}

func RunMenu(title, description string, options []MenuOption) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no menu options provided")
	}

	model := newMenuModel(title, description, options)
	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return "", err
	}

	resultModel, ok := finalModel.(*menuModel)
	if !ok {
		return "", fmt.Errorf("unexpected Bubble Tea model type %T", finalModel)
	}

	return resultModel.result()
}
