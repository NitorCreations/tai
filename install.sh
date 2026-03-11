#!/usr/bin/env sh
# tai install/update script
# Downloads the latest tai binary from GitHub Releases and installs it to
# ~/.local/share/tai/tai.
set -eu

REPO="NitorCreations/tai"
INSTALL_DIR="${HOME}/.local/share/tai"
BINARY_NAME="tai"

# ── helpers ──────────────────────────────────────────────────────────────────

die() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not found"
}

# ── detect OS and arch ───────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux"  ;;
    Darwin) echo "darwin" ;;
    *)      die "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)         echo "amd64" ;;
    aarch64|arm64|armv8*) echo "arm64" ;;
    *)                    die "unsupported architecture: $(uname -m)" ;;
  esac
}

# ── fetch latest release tag from GitHub API ─────────────────────────────────

latest_version() {
  need curl
  url="https://api.github.com/repos/${REPO}/releases/latest"
  version=$(curl -fsSL "$url" | grep '"tag_name":' | sed 's/.*"tag_name":[ ]*"\([^"]*\)".*/\1/')
  [ -n "$version" ] || die "could not determine latest version from GitHub API"
  echo "$version"
}

# ── main ──────────────────────────────────────────────────────────────────────

need curl
need tar

OS=$(detect_os)
ARCH=$(detect_arch)
VERSION=$(latest_version)

# goreleaser archive name template: tai_VERSION_OS_ARCH.tar.gz (version without "v" prefix)
ARCHIVE="tai_${VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

printf 'Installing tai %s (%s/%s)...\n' "$VERSION" "$OS" "$ARCH"

# ── download and extract ──────────────────────────────────────────────────────

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

printf 'Downloading %s\n' "$DOWNLOAD_URL"
curl -fsSL "$DOWNLOAD_URL" -o "${TMP}/${ARCHIVE}"
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP" "$BINARY_NAME"

# ── install binary ────────────────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"
cp "${TMP}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod 755 "${INSTALL_DIR}/${BINARY_NAME}"

# ── done ──────────────────────────────────────────────────────────────────────

printf '\ntai %s installed to %s\n' "$VERSION" "${INSTALL_DIR}/${BINARY_NAME}"

printf '\n'
printf 'Next steps:\n'
printf '\n'
printf '  1. Add tai to your PATH:\n'
printf '       export PATH="%s:$PATH"\n' "$INSTALL_DIR"
printf '  Add this line to ~/.bashrc or ~/.zshrc to make it permanent.\n'
printf '\n'
printf '  2. Install shell keybinding widget (optional):\n'
printf '  This creates tai-widget at ~/.local/bin for use with a shell keybinding.\n'
printf '  It lets you select a command interactively and inject it into your prompt:\n'
printf '\n'
printf '       tai install bash    # or: tai install zsh\n'
printf '\n'
printf '  3. Enable tab completions (optional):\n'
printf '       tai completion bash >> ~/.bashrc\n'
printf '       tai completion zsh  >> ~/.zshrc\n'
printf '\n'
printf '  4. Reload your shell:\n'
printf '       exec $SHELL\n'
