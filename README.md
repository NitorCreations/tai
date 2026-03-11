# tai — Terminal AI

Generate shell commands from natural language, right in your terminal.

tai uses your GitHub Copilot SDK to suggest commands. Select one, run it, or edit it before executing.

## Prerequisites

- [GitHub Copilot CLI](https://github.com/features/copilot/cli) — installed and logged in (subscription required)

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/NitorCreations/tai/main/install.sh | sh
```

The binary is placed at `~/.local/share/tai/tai`.

### Add to PATH

Add the install directory to your PATH so you can run `tai` directly:

```sh
export PATH="$HOME/.local/share/tai:$PATH"
```

Add this to `~/.bashrc` or `~/.zshrc` to make it permanent.

### Shell keybinding widget (optional)

`tai-widget` is a helper script that lets tai place commands directly into your shell prompt line instead of executing them in a subshell. Install it with:

```sh
tai install          # auto-detects bash or zsh
tai install bash     # explicit
tai install zsh      # explicit
```

This writes `tai-widget` to `~/.local/bin/`. Bind it to a key (e.g. `Alt+t`) for quick access from anywhere in the terminal.

### Tab completions (optional)

```sh
# bash
tai completion bash >> ~/.bashrc

# zsh
tai completion zsh >> ~/.zshrc
```

## Usage

```sh
tai ask                       # open interactive prompt
tai ask "list open ports"     # start with a query pre-filled
```

### TUI key bindings

| State | Key | Action |
|---|---|---|
| Input | `Enter` | Submit query |
| Results | `↑` / `↓` | Navigate suggestions |
| Results | `Enter` | Accept and run command |
| Results | `Tab` | Edit command before running |
| Results | `/` | Follow-up query (refine results) |
| Editing | `Enter` | Run edited command |
| Editing | `Esc` | Cancel edit, back to results |
| Follow-up | `Enter` | Submit follow-up |
| Follow-up | `Esc` / `Ctrl+C` | Back to results |
| Any | `Ctrl+C` | Quit |

Commands marked with `⚠` are flagged as potentially destructive.

## Configuration

Config file: `~/.config/tai/config.json`

```sh
tai config get model           # print current model
tai config set model gpt-4o    # change model
tai config edit                # open in $EDITOR
tai config reset               # restore defaults
```

Default model is `auto` (Copilot selects the model).

## Building from source

```sh
git clone https://github.com/NitorCreations/tai.git
cd tai
make
```

The binary is placed in `dist/`.

## License

MIT — see [LICENSE](LICENSE).
