package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/NitorCreations/tai/internal/config"
	sdk "github.com/github/copilot-sdk/go"
)

type Command struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Destructive bool   `json:"destructive"`
}

type ConversationTurn struct {
	Query    string
	Commands []Command
}

type Options struct {
	OnModelResolved func(model string)
	History         []ConversationTurn
}

const systemPrompt = `You are a shell command assistant. The user will describe what they want to do in the terminal in natural language.

CRITICAL: Your response must be a single raw JSON object. Do NOT include any text before or after the JSON. Do NOT use markdown. Do NOT use code fences. Do NOT explain anything. Just output the JSON.

Required format (output EXACTLY this structure, nothing else):
{"commands":[{"command":"<shell command>","description":"<one sentence>","destructive":<true|false>}]}

Rules:
- If the request is unambiguous, provide exactly 1 command
- If there are 2–3 meaningfully different valid approaches, list them all
- Mark destructive:true for commands that delete, overwrite, kill, format, or are otherwise hard to undo (rm, kill, mkfs, dd, truncate, etc.)
- Keep descriptions short (one sentence)
- START your response with { and END with } — no other characters outside the JSON`

var jsonRe = regexp.MustCompile(`\{[\s\S]*\}`)

func buildPrompt(query string, history []ConversationTurn) string {
	if len(history) == 0 {
		return query
	}
	var sb strings.Builder
	sb.WriteString("[Previous conversation]\n")
	for i, t := range history {
		fmt.Fprintf(&sb, "Turn %d:\n  User: %s\n  Commands:\n", i+1, t.Query)
		for _, c := range t.Commands {
			fmt.Fprintf(&sb, "  - %s\n", c.Command)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("[Follow-up request]\n")
	sb.WriteString(query)
	return sb.String()
}

func GetCommands(ctx context.Context, query string, cfg config.TaiConfig, opts Options) ([]Command, error) {
	client := sdk.NewClient(&sdk.ClientOptions{})
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Copilot client: %w", err)
	}
	defer client.Stop() //nolint:errcheck

	sessionCfg := &sdk.SessionConfig{
		SystemMessage: &sdk.SystemMessageConfig{
			Mode:    "replace",
			Content: systemPrompt,
		},
		OnPermissionRequest: sdk.PermissionHandler.ApproveAll,
	}
	if cfg.Model != "" {
		sessionCfg.Model = cfg.Model
	}

	session, err := client.CreateSession(ctx, sessionCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Disconnect() //nolint:errcheck

	var resolvedModel string
	reportModel := func(model string) {
		if model == "" || model == resolvedModel {
			return
		}
		resolvedModel = model
		if opts.OnModelResolved != nil {
			display := model
			if cfg.Model == "auto" {
				display = "auto:" + model
			}
			opts.OnModelResolved(display)
		}
	}

	unsub := session.On(func(event sdk.SessionEvent) {
		switch event.Type {
		case sdk.SessionStart:
			if event.Data.SelectedModel != nil {
				reportModel(*event.Data.SelectedModel)
			}
		case sdk.SessionModelChange:
			if event.Data.NewModel != nil {
				reportModel(*event.Data.NewModel)
			}
		case sdk.AssistantUsage:
			if event.Data.Model != nil {
				reportModel(*event.Data.Model)
			}
		}
	})
	defer unsub()

	prompt := buildPrompt(query, opts.History)
	result, err := session.SendAndWait(ctx, sdk.MessageOptions{Prompt: prompt})
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("no response from Copilot")
	}

	// Last-resort model resolution from message history.
	if resolvedModel == "" {
		if messages, err2 := session.GetMessages(ctx); err2 == nil {
			for i := len(messages) - 1; i >= 0; i-- {
				e := messages[i]
				switch e.Type {
				case sdk.AssistantUsage:
					if e.Data.Model != nil {
						reportModel(*e.Data.Model)
					}
				case sdk.SessionModelChange:
					if e.Data.NewModel != nil {
						reportModel(*e.Data.NewModel)
					}
				case sdk.SessionStart:
					if e.Data.SelectedModel != nil {
						reportModel(*e.Data.SelectedModel)
					}
				}
				if resolvedModel != "" {
					break
				}
			}
		}
	}

	if result.Data.Content == nil {
		return nil, fmt.Errorf("empty response from Copilot")
	}

	raw := strings.TrimSpace(*result.Data.Content)
	match := jsonRe.FindString(raw)
	if match == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var parsed struct {
		Commands []Command `json:"commands"`
	}
	if err := json.Unmarshal([]byte(match), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(parsed.Commands) == 0 {
		return nil, fmt.Errorf("no commands in response")
	}
	return parsed.Commands, nil
}
