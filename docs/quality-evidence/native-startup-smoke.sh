#!/usr/bin/env bash
set -euo pipefail
smoke_root=$(mktemp -d /tmp/datapack-native-smoke.XXXXXX)
out_dir="$1"
app_binary="$2"
mkdir -p "$smoke_root/config/datapack" "$smoke_root/cache" "$smoke_root/data"
printf '%s\n' '{"version":1,"build":{"mode":"manual"},"language":"en","theme":"light"}' > "$smoke_root/config/datapack/settings.json"
printf '%s\n' "$smoke_root" > "$out_dir/native-smoke-directory.txt"
export XDG_CONFIG_HOME="$smoke_root/config"
export XDG_CACHE_HOME="$smoke_root/cache"
export XDG_DATA_HOME="$smoke_root/data"
xvfb-run -a -s '-screen 0 1200x900x24' dbus-run-session -- bash -s -- "$app_binary" "$out_dir" <<'INNER'
set -euo pipefail
app_binary="$1"; out_dir="$2"
"$app_binary" > "$out_dir/native-app.log" 2>&1 &
app_pid=$!
trap 'kill -TERM "$app_pid" 2>/dev/null || true; wait "$app_pid" 2>/dev/null || true' EXIT
window_id=""
for attempt in {1..30}; do
 if ! kill -0 "$app_pid" 2>/dev/null; then printf 'native app exited before a visible window\n'; exit 1; fi
 window_id=$(xdotool search --onlyvisible --name '^datapack$' 2>/dev/null | head -n 1 || true)
 if [[ -n "$window_id" ]]; then break; fi
 sleep 0.2
done
if [[ -z "$window_id" ]]; then printf 'no visible native window within six seconds\n'; exit 1; fi
sleep 2
xwininfo -id "$window_id" > "$out_dir/native-window.txt"
import -window "$window_id" "$out_dir/native-welcome.png"
printf 'native window and screenshot captured\n'
# Coordinates come from the inspected 1024x768 welcome screenshot.
xdotool mousemove --window "$window_id" 430 464 click 1
dialog_id=""
for attempt in {1..30}; do
 dialog_id=$(xdotool search --onlyvisible --name '^Open Datapack Project$' 2>/dev/null | head -n 1 || true)
 if [[ -n "$dialog_id" ]]; then break; fi
 sleep 0.2
done
if [[ -z "$dialog_id" ]]; then printf 'native file dialog did not become visible\n'; exit 1; fi
import -window "$dialog_id" "$out_dir/native-file-dialog.png"
xdotool windowfocus "$dialog_id"
xdotool key Escape
for attempt in {1..30}; do
 if ! xdotool search --onlyvisible --name '^Open Datapack Project$' >/dev/null 2>&1; then break; fi
 sleep 0.2
done
if xdotool search --onlyvisible --name '^Open Datapack Project$' >/dev/null 2>&1; then printf 'dialog remained open after Escape\n'; exit 1; fi
sleep 2
import -window "$window_id" "$out_dir/native-after-cancel.png"
printf 'native file dialog opened and Escape cancellation captured\n'
INNER
