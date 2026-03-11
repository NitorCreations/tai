package tui

import (
	"strings"

	"github.com/NitorCreations/tai/internal/copilot"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Styles ────────────────────────────────────────────────────────────────────

var (
	cyanBold = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	gray     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dimGray  = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Faint(true)
	red      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	cyan     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	selected = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	normal   = lipgloss.NewStyle()
)

// ─── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	var lines []string

	// Header
	header := cyanBold.Render("tai")
	if m.query != "" {
		header += gray.Render("  " + m.query)
	}
	lines = append(lines, header)

	// Separator
	sep := separatorLine(m.width)
	lines = append(lines, sep)

	// Body
	switch m.state {
	case stateInput:
		lines = append(lines, renderInput(m.inputVal, m.inputCursor, "Describe the command you need..."))

	case stateLoading:
		loading := cyan.Render(m.spinner.View()+" ") + gray.Render("Asking Copilot")
		if m.resolvedModel != "" {
			loading += gray.Render(" (" + m.resolvedModel + ")")
		}
		loading += gray.Render("...")
		lines = append(lines, loading)

	case stateResults:
		lines = append(lines, renderCommands(m.commands, m.selectedIdx))

	case stateEditing:
		lines = append(lines, renderInput(m.inputVal, m.inputCursor, "Edit command..."))

	case stateFollowup:
		lines = append(lines, renderInput(m.inputVal, m.inputCursor, "Ask a follow-up to refine..."))

	case stateError:
		lines = append(lines, red.Render("✖ ")+red.Render(m.errMsg))
	}

	// Footer
	lines = append(lines, sep)
	lines = append(lines, renderFooter(m.state))

	body := strings.Join(lines, "\n")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("14")).
		PaddingLeft(1).
		PaddingRight(1)

	return borderStyle.Render(body) + "\n"
}

func separatorLine(width int) string {
	w := width - 6 // account for border (2) + padding (2) + some margin
	if w < 10 {
		w = 40
	}
	return gray.Render(strings.Repeat("─", w))
}

func renderInput(value string, cursor int, placeholder string) string {
	if value == "" {
		return "> " + gray.Render(placeholder) + cyan.Render("█")
	}
	bytePos := runeBytePos(value, cursor)
	before := value[:bytePos]
	after := value[bytePos:]
	return "> " + before + cyan.Render("█") + after
}

func renderCommands(commands []copilot.Command, selectedIdx int) string {
	var lines []string
	for i, cmd := range commands {
		var line string
		if i == selectedIdx {
			line = selected.Render("❯ " + cmd.Command)
		} else {
			line = normal.Render("  " + cmd.Command)
		}
		if cmd.Destructive {
			line += " " + red.Render("⚠ destructive")
		}
		lines = append(lines, line)
		lines = append(lines, dimGray.Render("    "+cmd.Description))
	}
	return strings.Join(lines, "\n")
}

func renderFooter(state appState) string {
	switch state {
	case stateResults:
		return dimGray.Render("[↑↓] Navigate  [Enter] Run  [Tab] Edit  [/] Refine  [Ctrl+C] Cancel")
	case stateInput:
		return dimGray.Render("[Enter] Submit  [Ctrl+C] Cancel")
	case stateEditing:
		return dimGray.Render("[Enter] Submit  [Esc] Back  [Ctrl+C] Cancel")
	case stateFollowup:
		return dimGray.Render("[Enter] Refine  [Esc] Back  [Ctrl+C] Cancel")
	case stateLoading:
		return dimGray.Render("[Ctrl+C] Cancel")
	case stateError:
		return dimGray.Render("[Ctrl+C] Close")
	}
	return ""
}

// Ensure Model implements tea.Model at compile time.
var _ tea.Model = Model{}
