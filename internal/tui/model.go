package tui

import (
	"context"
	"time"

	"github.com/NitorCreations/tai/internal/config"
	"github.com/NitorCreations/tai/internal/copilot"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// ─── Message types ─────────────────────────────────────────────────────────────

type CommandsMsg struct{ Commands []copilot.Command }
type ModelResolvedMsg struct{ Model string }
type ErrMsg struct{ Msg string }

// ─── App state ─────────────────────────────────────────────────────────────────

type appState int

const (
	stateInput    appState = iota
	stateLoading  appState = iota
	stateResults  appState = iota
	stateEditing  appState = iota
	stateFollowup appState = iota
	stateError    appState = iota
)

// ─── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	state         appState
	inputVal      string
	inputCursor   int // rune index into inputVal
	commands      []copilot.Command
	selectedIdx   int
	history       []copilot.ConversationTurn
	resolvedModel string
	errMsg        string
	spinner       spinner.Model
	cancelFunc    context.CancelFunc
	cfg           config.TaiConfig
	query         string // query that produced current results/loading state
	Accepted      string // set when user accepts a command
	Cancelled     bool   // set when user cancels
	program       **tea.Program
	width         int
}

var brailleSpinner = spinner.Spinner{
	Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	FPS:    time.Second / 12,
}

func NewModel(cfg config.TaiConfig, initialQuery string, program **tea.Program) Model {
	s := spinner.New()
	s.Spinner = brailleSpinner

	state := stateInput
	query := ""
	if initialQuery != "" {
		state = stateLoading
		query = initialQuery
	}

	return Model{
		state:         state,
		inputVal:      initialQuery,
		query:         query,
		spinner:       s,
		cfg:           cfg,
		program:       program,
		resolvedModel: cfg.Model,
	}
}

// ─── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}
	if m.state == stateLoading {
		cmds = append(cmds, m.fetchCmd(m.query, m.history))
	}
	return tea.Batch(cmds...)
}
