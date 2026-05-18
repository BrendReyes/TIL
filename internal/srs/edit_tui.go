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

type model struct {
	bodyInput textarea.Model
	tagInput  textinput.Model
	focus     int
	err       error
	saved     bool
	quitting  bool
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
		focus:     0,
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
		case "esc", "ctrl+c":
			m.quitting = true
			return m, bubbletea.Quit

		case "ctrl+s", "ctrl+enter":
			if err := m.validate(); err != nil {
				m.err = err
				return m, nil
			}
			m.saved = true
			return m, bubbletea.Quit

		case "enter":
			// Enter on the tag field attempts to save
			if m.focus == 1 {
				if err := m.validate(); err != nil {
					m.err = err
					return m, nil
				}
				m.saved = true
				return m, bubbletea.Quit
			}

		case "tab", "shift+tab":
			if m.focus == 0 {
				m.focus = 1
				m.bodyInput.Blur()
				cmds = append(cmds, m.tagInput.Focus())
			} else {
				m.focus = 0
				m.tagInput.Blur()
				cmds = append(cmds, m.bodyInput.Focus())
			}
		}
	}

	var cmd bubbletea.Cmd
	if m.focus == 0 {
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	} else {
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

	if m.focus == 0 {
		bodyLabel = focusedStyle.Render("● " + bodyLabel)
		tagLabel = blurredStyle.Render("  " + tagLabel)
		bodyView = focusedBorderStyle.Render(m.bodyInput.View())
		tagView = blurredBorderStyle.Render(m.tagInput.View())
	} else {
		bodyLabel = blurredStyle.Render("  " + bodyLabel)
		tagLabel = focusedStyle.Render("● " + tagLabel)
		bodyView = blurredBorderStyle.Render(m.bodyInput.View())
		tagView = focusedBorderStyle.Render(m.tagInput.View())
	}

	errDisplay := ""
	if m.err != nil {
		errDisplay = errorStyle.Render(m.err.Error())
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
		helpStyle.Render("tab: switch • enter: save (on tag) • ctrl+s: save • esc: cancel"),
	)

	return docStyle.Render(s)
}