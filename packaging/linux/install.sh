#!/usr/bin/env bash
# Install Script Launcher for this user: the program in ~/.local/bin and its entry in the applications menu.
# Usage: bash install.sh   (from the unpacked folder)
set -e
cd "$(dirname "$0")"
mkdir -p "$HOME/.local/bin" "$HOME/.local/share/applications"
install -m 755 script-launcher "$HOME/.local/bin/script-launcher"
sed "s|@BIN@|$HOME/.local/bin/script-launcher|" script-launcher.desktop > "$HOME/.local/share/applications/script-launcher.desktop"
echo "Script Launcher instalado. Búscalo en el menú de aplicaciones como «Script Launcher»."
