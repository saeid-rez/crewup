package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestMultiSelectModelToggleCursor(t *testing.T) {
	m := newMultiSelectModel("Pick tools", []item{{id: "copilot", title: "Copilot"}})

	if m.items[0].selected {
		t.Fatal("expected item to start deselected")
	}

	m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	if !m.items[0].selected {
		t.Fatal("expected space to select the current item")
	}

	m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	if m.items[0].selected {
		t.Fatal("expected second space to deselect the current item")
	}
}

func TestMultiSelectModelToggleAll(t *testing.T) {
	m := newMultiSelectModel("Pick roles", []item{{id: "planner", title: "Planner"}, {id: "reviewer", title: "Reviewer"}})

	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if !m.items[0].selected || !m.items[1].selected {
		t.Fatal("expected a to select all items")
	}

	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if m.items[0].selected || m.items[1].selected {
		t.Fatal("expected second a to deselect all items")
	}
}

func TestMultiSelectModelCursorMovement(t *testing.T) {
	m := newMultiSelectModel("Pick MCPs", []item{{id: "one", title: "One"}, {id: "two", title: "Two"}})

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", m.cursor)
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 1 {
		t.Fatalf("expected cursor to stay at lower bound, got %d", m.cursor)
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if m.cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", m.cursor)
	}
}

func TestMultiSelectModelResultReturnsSelectedIDsInOrder(t *testing.T) {
	m := newMultiSelectModel("Pick tools", []item{{id: "copilot", title: "Copilot", selected: true}, {id: "opencode", title: "OpenCode"}, {id: "claude", title: "Claude", selected: true}})
	m.done = true

	ids, err := m.result()
	if err != nil {
		t.Fatalf("result error: %v", err)
	}

	if len(ids) != 2 || ids[0] != "copilot" || ids[1] != "claude" {
		t.Fatalf("unexpected selected IDs: %#v", ids)
	}
}

func TestMultiSelectModelResultReturnsCancellationError(t *testing.T) {
	m := newMultiSelectModel("Pick tools", []item{{id: "copilot", title: "Copilot"}})
	m.quitting = true

	_, err := m.result()
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("expected ErrUserCancelled, got %v", err)
	}
}
