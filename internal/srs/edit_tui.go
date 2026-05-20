package srs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for a polished look
var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).MarginBottom(1)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			MarginBottom(1).
			Bold(true)
)

// focus zones for the edit form (shared by standalone RunEditor and list inline edit)
const (
	editFocusBody    = 0
	editFocusTag     = 1
	editFocusButtons = 2
)

// button indices
const (
	editBtnSave   = 0
	editBtnCancel = 1
)

type model struct {
	bodyInput textarea.Model
	tagInput  textinput.Model
	focus     int // editFocusBody | editFocusTag | editFocusButtons
	bar       buttonBar
	err       error
	saved     bool
	quitting  bool
}

func newEditBar() buttonBar {
	bar := newButtonBar([]string{"Save", "Cancel"})
	bar.SetDanger(editBtnCancel)
	return bar
}

func RunEditor(initialBody, initialTag string) (string, string, bool, error) {
	ta := textarea.New()
	ta.Placeholder = "What did you learn?"
	ta.SetValue(initialBody)
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(50)
	ta.SetHeight(5)

	ti := textinput.New()
	ti.Placeholder = "e.g. go, sql, algorithms"
	ti.SetValue(initialTag)
	ti.CharLimit = 100
	ti.Width = 30

	m := model{
		bodyInput: ta,
		tagInput:  ti,
		focus:     editFocusBody,
		bar:       newEditBar(),
	}

	p := bubbletea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", "", false, err
	}

	m = finalModel.(model)
	return m.bodyInput.Value(), m.tagInput.Value(), m.saved, nil
}

func (m model) Init() bubbletea.Cmd {
	return textarea.Blink
}

// validate checks both fields and returns the first error found, or nil.
func (m model) validate() error {
	if strings.TrimSpace(m.bodyInput.Value()) == "" {
		return fmt.Errorf("Error: Body cannot be empty")
	}
	if strings.TrimSpace(m.tagInput.Value()) == "" {
		return fmt.Errorf("Error: Tag cannot be empty")
	}
	return nil
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	var cmds []bubbletea.Cmd

	switch msg := msg.(type) {
	case bubbletea.KeyMsg:
		m.err = nil

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, bubbletea.Quit

		case "esc":
			// esc always cancels from any zone
			m.quitting = true
			return m, bubbletea.Quit

		case "ctrl+s", "ctrl+enter":
			if err := m.validate(); err != nil {
				m.err = err
				return m, nil
			}
			m.saved = true
			return m, bubbletea.Quit

		case "tab":
			switch m.focus {
			case editFocusBody:
				m.focus = editFocusTag
				m.bodyInput.Blur()
				m.bar.focused = false
				cmds = append(cmds, m.tagInput.Focus())
			case editFocusTag:
				m.focus = editFocusButtons
				m.tagInput.Blur()
				m.bar.focused = true
			case editFocusButtons:
				m.focus = editFocusBody
				m.bar.focused = false
				cmds = append(cmds, m.bodyInput.Focus())
			}
			return m, bubbletea.Batch(cmds...)

		case "shift+tab":
			switch m.focus {
			case editFocusBody:
				m.focus = editFocusButtons
				m.bodyInput.Blur()
				m.bar.focused = true
			case editFocusTag:
				m.focus = editFocusBody
				m.tagInput.Blur()
				m.bar.focused = false
				cmds = append(cmds, m.bodyInput.Focus())
			case editFocusButtons:
				m.focus = editFocusTag
				m.bar.focused = false
				cmds = append(cmds, m.tagInput.Focus())
			}
			return m, bubbletea.Batch(cmds...)
		}

		// ── Button bar active ────────────────────────────────────────────────
		if m.focus == editFocusButtons {
			m.bar.HandleKey(msg.String())
			if m.bar.Activated() {
				switch m.bar.ActiveIndex() {
				case editBtnSave:
					if err := m.validate(); err != nil {
						m.err = err
						m.focus = editFocusBody
						m.bar.focused = false
						cmds = append(cmds, m.bodyInput.Focus())
						return m, bubbletea.Batch(cmds...)
					}
					m.saved = true
					return m, bubbletea.Quit
				case editBtnCancel:
					m.quitting = true
					return m, bubbletea.Quit
				}
			}
			return m, nil
		}

		// ── enter on tag field saves ─────────────────────────────────────────
		if msg.String() == "enter" && m.focus == editFocusTag {
			if err := m.validate(); err != nil {
				m.err = err
				return m, nil
			}
			m.saved = true
			return m, bubbletea.Quit
		}
	}

	var cmd bubbletea.Cmd
	if m.focus == editFocusBody {
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	} else if m.focus == editFocusTag {
		m.tagInput, cmd = m.tagInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, bubbletea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	bodyLabel := "Learning Entry"
	tagLabel := "Tag (required)"
	var bodyView, tagView string

	switch m.focus {
	case editFocusBody:
		bodyLabel = focusedStyle.Render("● " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = focusedBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	case editFocusTag:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = focusedStyle.Render("● " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = focusedBorderStyle.Render(m.tagInput.View())
	case editFocusButtons:
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	}

	errDisplay := ""
	if m.err != nil {
		errDisplay = errorStyle.Render(m.err.Error())
	}

	var help string
	if m.focus == editFocusButtons {
		help = helpStyle.Render("←/→: choose action • enter: activate • tab: back to body")
	} else {
		help = helpStyle.Render("tab: next field • enter: save (on tag) • ctrl+s: save • esc: cancel")
	}

	s := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Edit Entry"),
		errDisplay,
		bodyLabel,
		bodyView,
		"",
		tagLabel,
		tagView,
		m.bar.View(),
		help,
	)

	return docStyle.Render(s)
}