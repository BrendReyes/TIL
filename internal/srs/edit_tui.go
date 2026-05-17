package srs

import (
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    bubbletea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// Styles for a polished look
var (
    focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    docStyle     = lipgloss.NewStyle().Margin(1, 2)
    helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
    titleStyle   = lipgloss.NewStyle().
            Background(lipgloss.Color("62")).
            Foreground(lipgloss.Color("230")).
            Padding(0, 1).
            MarginBottom(1).
            Bold(true)
)

type model struct {
    bodyInput textarea.Model
    tagInput  textinput.Model
    focus     int // 0 for body, 1 for tag
    err       error
    saved     bool
    quitting  bool
}

// RunEditor launches the TUI and returns (body, tag, saved, error)
func RunEditor(initialBody, initialTag string) (string, string, bool, error) {
    ta := textarea.New()
    ta.Placeholder = "What did you learn?"
    ta.SetValue(initialBody)
    ta.Focus()
    ta.CharLimit = 500
    ta.SetWidth(50)
    ta.SetHeight(5)
    ti := textinput.New()
    ti.Placeholder = "Tags (optional)"
    ti.SetValue(initialTag)
    ti.CharLimit = 50
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

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
    var cmds []bubbletea.Cmd
    switch msg := msg.(type) {
    case bubbletea.KeyMsg:
        switch msg.String() {
        case "esc", "ctrl+c":
            m.quitting = true
            return m, bubbletea.Quit
        case "ctrl+s":
            m.saved = true
            return m, bubbletea.Quit
        case "tab", "shift+tab":
            // Toggle focus between fields
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
	// Update the focused component
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
    // Style the labels based on focus
    bodyLabel := "Learning Entry"
    tagLabel := "Tags"
    if m.focus == 0 {
        bodyLabel = focusedStyle.Render("● " + bodyLabel)
        tagLabel = blurredStyle.Render("  " + tagLabel)
    } else {
        bodyLabel = blurredStyle.Render("  " + bodyLabel)
        tagLabel = focusedStyle.Render("● " + tagLabel)
    }
    s := lipgloss.JoinVertical(
        lipgloss.Left,
        titleStyle.Render("Edit Entry"),
        bodyLabel,
        m.bodyInput.View(),
        "",
        tagLabel,
        m.tagInput.View(),
        helpStyle.Render("tab: switch • ctrl+s: save • esc: cancel"),
    )
    return docStyle.Render(s)
}