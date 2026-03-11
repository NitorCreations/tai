package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/NitorCreations/tai/internal/config"
	tui "github.com/NitorCreations/tai/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Execute is the main CLI entry point. Returns the exit code.
func Execute() int {
	var exitCode int

	rootCmd := &cobra.Command{
		Use:   "tai",
		Short: "Terminal AI: generate shell commands from natural language",
		Long: `tai - Terminal AI: generate shell commands from natural language

Use "tai ask [query]" to open the interactive overlay.`,
		SilenceUsage: true,
	}

	askCmd := &cobra.Command{
		Use:          "ask [query]",
		Short:        "Open the interactive TUI overlay with an optional initial query",
		Long:         `Open the interactive TUI overlay. Optionally provide an initial query.`,
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode = runTUI(strings.Join(args, " "))
			return nil
		},
	}

	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return exitCode
}

func runTUI(initialQuery string) int {
	cfg := config.LoadConfig()

	ttyFile, err := OpenTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening /dev/tty: %v\n", err)
		return 1
	}
	defer ttyFile.Close()

	restoreStderr := func() {}
	if restore, err := RedirectStderrToNull(); err == nil {
		restoreStderr = restore
	}

	var prog *tea.Program
	m := tui.NewModel(cfg, initialQuery, &prog)
	prog = tea.NewProgram(m, tea.WithInput(ttyFile), tea.WithOutput(ttyFile))

	finalModel, err := prog.Run()
	restoreStderr()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	FlushTTYInput(ttyFile)

	result := finalModel.(tui.Model)
	if result.Accepted != "" {
		if os.Getenv("TAI_SHIM") != "" {
			fmt.Fprint(os.Stdout, result.Accepted)
			return 0
		}
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "sh"
		}
		c := exec.Command(shell, "-c", result.Accepted)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return exitErr.ExitCode()
			}
			return 1
		}
		return 0
	}
	return 1
}
