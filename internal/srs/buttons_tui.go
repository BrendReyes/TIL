package srs

import (
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Button bar — reusable horizontal row of labeled action buttons.
//
// Usage:
//   bar := newButtonBar([]string{"Save", "Cancel"})
//   bar.focused = true          // whether the bar has keyboard focus
//
// Rendering:
//   bar.View()                  // returns the rendered row
//
// Input handling (call from parent Update):
//   handled, cmd := bar.Update(msg)
//   if handled { ... check bar.Activated() ... }
//
// After each Update call, check bar.activated to see if the user pressed
// enter on a button. Read bar.cursor to know which one.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	btnFocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 2).
			Bold(true).
			MarginRight(1)

	btnSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("205")).
				Padding(0, 2).
				Bold(true).
				MarginRight(1)

	btnNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")).
			Padding(0, 2).
			MarginRight(1)

	btnBarStyle = lipgloss.NewStyle().MarginTop(1)

	// Danger variant — used for destructive actions (delete, reset)
	btnDangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("9")).
			Padding(0, 2).
			Bold(true).
			MarginRight(1)
)

// ---------------------------------------------------------------------------
// buttonBar
// ---------------------------------------------------------------------------

type buttonBar struct {
	labels    []string // button labels in order
	danger    []bool   // parallel slice; true = render with danger style when focused
	cursor    int      // which button is highlighted
	focused   bool     // whether this bar has keyboard focus
	activated bool     // true for exactly one frame after enter is pressed
}

func newButtonBar(labels []string) buttonBar {
	return buttonBar{
		labels:  labels,
		danger:  make([]bool, len(labels)),
		cursor:  0,
		focused: false,
	}
}

// SetDanger marks button at index i as a destructive action.
func (b *buttonBar) SetDanger(i int) {
	if i >= 0 && i < len(b.danger) {
		b.danger[i] = true
	}
}

// Activated returns true if the user pressed enter this frame, then resets.
func (b *buttonBar) Activated() bool {
	if b.activated {
		b.activated = false
		return true
	}
	return false
}

// ActiveIndex returns which button was just activated.
func (b *buttonBar) ActiveIndex() int {
	return b.cursor
}

// ---------------------------------------------------------------------------
// Update — call from parent screen's Update when bar.focused == true.
// Returns true if the message was consumed by the bar.
// ---------------------------------------------------------------------------

func (b *buttonBar) HandleKey(key string) (consumed bool) {
	b.activated = false

	switch key {
	case "left", "h":
		if b.cursor > 0 {
			b.cursor--
		}
		return true

	case "right", "l":
		if b.cursor < len(b.labels)-1 {
			b.cursor++
		}
		return true

	case "enter", " ":
		b.activated = true
		return true
	}

	return false
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (b buttonBar) View() string {
	var btns []string
	for i, label := range b.labels {
		var s string
		if b.focused && i == b.cursor {
			if b.danger[i] {
				s = btnDangerStyle.Render(" " + label + " ")
			} else {
				s = btnSelectedStyle.Render(" " + label + " ")
			}
		} else if b.focused {
			s = btnNormalStyle.Render(" " + label + " ")
		} else {
			// bar not focused — render all as muted
			s = btnNormalStyle.Copy().
				Foreground(lipgloss.Color("240")).
				Render(" " + label + " ")
		}
		btns = append(btns, s)
	}

	return btnBarStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, btns...),
	)
}

// ---------------------------------------------------------------------------
// focusZone — tracks which interactive zone on a screen has focus.
// Screens with multiple zones (e.g. rows + button bar) use this.
// ---------------------------------------------------------------------------

type focusZone int

const (
	zoneRows    focusZone = iota // list rows / form fields
	zoneButtons                  // button bar
)