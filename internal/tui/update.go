package tui

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/NitorCreations/tai/internal/copilot"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// ─── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ModelResolvedMsg:
		m.resolvedModel = msg.Model

	case CommandsMsg:
		m.state = stateResults
		m.commands = msg.Commands
		m.selectedIdx = 0
		m.inputVal = ""
		m.inputCursor = 0

	case ErrMsg:
		if !m.Cancelled {
			m.state = stateError
			m.errMsg = msg.Msg
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {

	case stateInput:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if strings.TrimSpace(m.inputVal) == "" {
				return m, nil
			}
			m.query = strings.TrimSpace(m.inputVal)
			m.state = stateLoading
			m.resolvedModel = m.cfg.Model
			return m, m.fetchCmd(m.query, m.history)
		default:
			m.inputVal, m.inputCursor = inputKey(m.inputVal, m.inputCursor, msg)
		}

	case stateLoading:
		if msg.Type == tea.KeyCtrlC {
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			m.Cancelled = true
			return m, tea.Quit
		}

	case stateResults:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Cancelled = true
			return m, tea.Quit
		case tea.KeyUp:
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case tea.KeyDown:
			if m.selectedIdx < len(m.commands)-1 {
				m.selectedIdx++
			}
		case tea.KeyEnter:
			if m.selectedIdx < len(m.commands) {
				m.Accepted = m.commands[m.selectedIdx].Command
				return m, tea.Quit
			}
		case tea.KeyTab:
			if m.selectedIdx < len(m.commands) {
				m.inputVal = m.commands[m.selectedIdx].Command
				m.inputCursor = utf8.RuneCountInString(m.inputVal)
				m.state = stateEditing
			}
		case tea.KeyRunes:
			if msg.String() == "/" {
				m.inputVal = m.query
				m.inputCursor = utf8.RuneCountInString(m.inputVal)
				m.state = stateFollowup
			}
		}

	case stateEditing:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.Cancelled = true
			return m, tea.Quit
		case tea.KeyEscape:
			m.state = stateResults
		case tea.KeyEnter:
			m.Accepted = m.inputVal
			return m, tea.Quit
		default:
			m.inputVal, m.inputCursor = inputKey(m.inputVal, m.inputCursor, msg)
		}

	case stateFollowup:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// back to results
			m.state = stateResults
		case tea.KeyEnter:
			if strings.TrimSpace(m.inputVal) == "" {
				return m, nil
			}
			followupQuery := strings.TrimSpace(m.inputVal)
			newHistory := append(m.history, copilot.ConversationTurn{
				Query:    m.query,
				Commands: m.commands,
			})
			m.history = newHistory
			m.query = followupQuery
			m.inputVal = ""
			m.inputCursor = 0
			m.state = stateLoading
			m.resolvedModel = m.cfg.Model
			return m, m.fetchCmd(followupQuery, newHistory)
		default:
			m.inputVal, m.inputCursor = inputKey(m.inputVal, m.inputCursor, msg)
		}

	case stateError:
		if msg.Type == tea.KeyCtrlC {
			m.state = stateInput
			m.inputVal = ""
		}
	}

	return m, nil
}

// runeBytePos returns the byte offset of the cursor-th rune boundary in s.
func runeBytePos(s string, cursor int) int {
	pos := 0
	for i := 0; i < cursor; i++ {
		_, size := utf8.DecodeRuneInString(s[pos:])
		pos += size
	}
	return pos
}

// inputKey handles text-editing keys for any input field, returning the updated
// (value, cursorRuneIndex). Keys not handled here are silently ignored.
func inputKey(val string, cursor int, msg tea.KeyMsg) (string, int) {
	runeLen := utf8.RuneCountInString(val)
	switch msg.Type {
	case tea.KeyLeft:
		if cursor > 0 {
			cursor--
		}
	case tea.KeyRight:
		if cursor < runeLen {
			cursor++
		}
	case tea.KeyHome:
		cursor = 0
	case tea.KeyEnd:
		cursor = runeLen
	case tea.KeyBackspace:
		if cursor > 0 {
			bytePos := runeBytePos(val, cursor)
			_, prevSize := utf8.DecodeLastRuneInString(val[:bytePos])
			val = val[:bytePos-prevSize] + val[bytePos:]
			cursor--
		}
	case tea.KeyDelete:
		if cursor < runeLen {
			bytePos := runeBytePos(val, cursor)
			_, nextSize := utf8.DecodeRuneInString(val[bytePos:])
			val = val[:bytePos] + val[bytePos+nextSize:]
		}
	case tea.KeySpace:
		bytePos := runeBytePos(val, cursor)
		val = val[:bytePos] + " " + val[bytePos:]
		cursor++
	case tea.KeyRunes:
		ch := msg.String()
		bytePos := runeBytePos(val, cursor)
		val = val[:bytePos] + ch + val[bytePos:]
		cursor += utf8.RuneCountInString(ch)
	}
	return val, cursor
}

func (m Model) fetchCmd(query string, history []copilot.ConversationTurn) tea.Cmd {
	prog := m.program
	cfg := m.cfg
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		_ = cancel // will be called via model.cancelFunc (set below)

		// We can't easily set the cancelFunc from inside a tea.Cmd because
		// tea.Cmd runs in a goroutine and the model is immutable. Instead we
		// use a fresh context and live with the limitation that Ctrl+C during
		// loading quits immediately (the goroutine will be orphaned briefly).
		cmds, err := copilot.GetCommands(ctx, query, cfg, copilot.Options{
			History: history,
			OnModelResolved: func(model string) {
				if prog != nil && *prog != nil {
					(*prog).Send(ModelResolvedMsg{Model: model})
				}
			},
		})
		if err != nil {
			return ErrMsg{Msg: err.Error()}
		}
		return CommandsMsg{Commands: cmds}
	}
}
