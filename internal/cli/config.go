package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/NitorCreations/tai/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Manage configuration",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.LoadConfig()
			fmt.Fprintf(os.Stderr, "Config file: %s\n", config.ConfigPath())
			printConfig(cfg)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:          "get <key>",
		Short:        "Get a config value",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			handleConfigGet(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:          "set <key> <value>",
		Short:        "Set a config value",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			handleConfigSet(args[0], args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:          "edit",
		Short:        "Open config in $EDITOR",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			handleConfigEdit()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:          "reset",
		Short:        "Reset config to defaults",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			handleConfigReset()
		},
	})

	return cmd
}

func handleConfigGet(key string) {
	cfg := config.LoadConfig()
	switch key {
	case "model":
		fmt.Println(cfg.Model)
	default:
		fmt.Fprintf(os.Stderr, "Unknown config key: %s\nValid keys: model\n", key)
		os.Exit(1)
	}
}

func handleConfigSet(key, value string) {
	cfg := config.LoadConfig()
	switch key {
	case "model":
		cfg.Model = value
	default:
		fmt.Fprintf(os.Stderr, "Unknown config key: %s\nValid keys: model\n", key)
		os.Exit(1)
	}
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Set %s=%s\n", key, value)
	printConfig(cfg)
}

func handleConfigEdit() {
	cfg := config.LoadConfig()
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, config.ConfigPath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open editor: %v\n", err)
		os.Exit(1)
	}
}

func handleConfigReset() {
	fmt.Fprint(os.Stderr, "Reset config to defaults? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) == "y" {
		if err := config.SaveConfig(config.DefaultConfig()); err != nil {
			fmt.Fprintf(os.Stderr, "Error resetting config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "Config reset to defaults.")
	} else {
		fmt.Fprintln(os.Stderr, "Cancelled.")
	}
}

func printConfig(cfg config.TaiConfig) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", data)
}
