// Package ui provides terminal UI components for Claudex session management.
// It uses the Bubble Tea framework to provide interactive session selection,
// profile selection, and other UI workflows.
package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF87")).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	dimmedItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			PaddingLeft(4)
)

type SessionItem struct {
	Title       string
	Description string
	Created     time.Time
	ItemType    string // "new", "ephemeral", "session"
}

func (i SessionItem) FilterValue() string { return i.Title }

type Model struct {
	List        list.Model
	SessionName string
	SessionPath string
	ProjectDir  string
	SessionsDir string
	Stage       string
	Quitting    bool
	Choice      string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case SessionChoiceMsg:
		m.SessionName = msg.SessionName
		m.SessionPath = msg.SessionPath
		m.Choice = msg.ItemType
		return m, tea.Quit

	case ProfileChoiceMsg:
		m.Choice = msg.ProfileName
		return m, tea.Quit

	case ResumeOrForkChoiceMsg:
		m.Choice = msg.Choice
		return m, tea.Quit

	case ResumeSubmenuChoiceMsg:
		m.Choice = msg.Choice
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.List.SelectedItem().(SessionItem)
			if ok {
				m.Choice = i.Title
				if m.Stage == "session" {
					return m, m.handleSessionChoice(i)
				} else if m.Stage == "profile" {
					return m, m.handleProfileChoice(i)
				} else if m.Stage == "resume_or_fork" {
					return m, m.handleResumeOrForkChoice(i)
				} else if m.Stage == "resume_submenu" {
					return m, m.handleResumeSubmenuChoice(i)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

type SessionChoiceMsg struct {
	SessionName string
	SessionPath string
	ItemType    string
}

func (m Model) handleSessionChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		if item.ItemType == "new" {
			// Return message to quit and handle outside TUI
			return SessionChoiceMsg{ItemType: "new"}
		}

		var sessionName, sessionPath string

		switch item.ItemType {
		case "ephemeral":
			sessionName = "ephemeral"
			sessionPath = ""

		case "session":
			sessionName = item.Title
			sessionPath = filepath.Join(m.SessionsDir, item.Title)
		}

		return SessionChoiceMsg{
			SessionName: sessionName,
			SessionPath: sessionPath,
			ItemType:    item.ItemType,
		}
	}
}

type ProfileChoiceMsg struct {
	ProfileName string
}

func (m Model) handleProfileChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ProfileChoiceMsg{ProfileName: item.Title}
	}
}

type ResumeOrForkChoiceMsg struct {
	Choice string // "resume" or "fork"
}

type ResumeSubmenuChoiceMsg struct {
	Choice string // "continue" or "fresh"
}

func (m Model) handleResumeOrForkChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ResumeOrForkChoiceMsg{Choice: item.ItemType}
	}
}

func (m Model) handleResumeSubmenuChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ResumeSubmenuChoiceMsg{Choice: item.ItemType}
	}
}

func (m Model) View() string {
	if m.Quitting {
		return "\n  👋 Goodbye!\n\n"
	}

	return docStyle.Render(m.List.View())
}

// Custom delegate for better item rendering
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 2 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SessionItem)
	if !ok {
		return
	}

	var icon string
	switch i.ItemType {
	case "new":
		icon = "➕"
	case "ephemeral":
		icon = "⚡"
	case "session":
		icon = "📁"
	case "profile":
		icon = "🎭"
	case "continue":
		icon = "▶"
	case "fresh":
		icon = "🔄"
	}

	str := fmt.Sprintf("%s %s", icon, i.Title)
	if i.Description != "" {
		str = fmt.Sprintf("%s\n   %s", str, dimmedItemStyle.Render(i.Description))
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("▶ "+str))
	} else {
		fmt.Fprint(w, normalItemStyle.Render(str))
	}
}

// Exported styles for external use
func TitleStyle() lipgloss.Style {
	return titleStyle
}
