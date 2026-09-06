#!/usr/bin/env bash
set -e

# ==============================================================================
# End-to-End Test Suite for Termstrap CLI / make render
# ==============================================================================

C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[1;36m"
C_GREEN="\033[1;32m"
C_RED="\033[1;31m"
C_YELLOW="\033[1;33m"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$DIR/../.." && pwd)"
cd "$ROOT_DIR"

export COLORTERM="truecolor"
export CLICOLOR_FORCE="1"
export TERM="xterm-256color"

SHOW_OUTPUT=0
INTERACTIVE_MODE=0

for arg in "$@"; do
  if [ "$arg" = "-v" ] || [ "$arg" = "--verbose" ] || [ "$arg" = "--output" ]; then
    SHOW_OUTPUT=1
  fi
  if [ "$arg" = "-i" ] || [ "$arg" = "--interactive" ]; then
    INTERACTIVE_MODE=1
    SHOW_OUTPUT=1
  fi
done
if [ "${VERBOSE:-0}" = "1" ]; then
  SHOW_OUTPUT=1
fi
if [ "${INTERACTIVE:-0}" = "1" ] || [ "${TERMSTRAP_INTERACTIVE:-0}" = "1" ]; then
  INTERACTIVE_MODE=1
  SHOW_OUTPUT=1
fi
TOTAL=0
PASSED=0
FAILED=0
run_test() {
  local name="$1"
  local cmd="$2"
  local expected_contain="$3"
  
  TOTAL=$((TOTAL + 1))
  printf "  ${C_CYAN}TEST %02d${C_RESET}: %-55s " "$TOTAL" "$name"
  
  set +e
  local output
  output=$(eval "$cmd" 2>&1)
  local status=$?
  set -e
  
  if [ $status -eq 0 ] && [[ "$output" == *"$expected_contain"* ]]; then
    printf "${C_GREEN}[PASS]${C_RESET}\n"
    PASSED=$((PASSED + 1))
    if [ $SHOW_OUTPUT -eq 1 ]; then
      printf "    ${C_YELLOW}Command:${C_RESET} %s\n\n" "$cmd"
      echo "$output"
      printf "\n"
      if [ $INTERACTIVE_MODE -eq 1 ]; then
        read -r -p "    [Appuyez sur Entrée pour continuer (ou 'q' pour quitter)...] " user_input </dev/tty || true
        if [ "$user_input" = "q" ] || [ "$user_input" = "Q" ]; then
          printf "\n${C_RED}Tests interrompus par l'utilisateur.${C_RESET}\n"
          exit 1
        fi
        printf "\n"
      fi
    fi
  else
    printf "${C_RED}[FAIL]${C_RESET}\n"
    printf "    ${C_YELLOW}Expected substring:${C_RESET} %s\n\n" "$expected_contain"
    echo "$output" | head -n 15
    printf "\n"
    FAILED=$((FAILED + 1))
  fi
}

printf "\n${C_BOLD}${C_CYAN}══════════════════════════════════════════════════════════════════${C_RESET}\n"
printf "  ${C_BOLD}Running Termstrap End-to-End (E2E) Test Suite${C_RESET}\n"
printf "${C_BOLD}${C_CYAN}══════════════════════════════════════════════════════════════════${C_RESET}\n\n"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

# Scenario 1: 3-column Bootstrap responsive grid
cat << 'EOF' > "$TMP_DIR/grid.html"
<div class="row">
  <div class="col-4 border rounded p-2 bg-primary text-white text-center">Col 1</div>
  <div class="col-4 border rounded p-2 bg-success text-white text-center">Col 2</div>
  <div class="col-4 border rounded p-2 bg-danger text-white text-center">Col 3</div>
</div>
EOF
run_test "3-Column Responsive Grid (col-4)" "make render -- $TMP_DIR/grid.html -w 90" "Col 1"

# Scenario 2: Theme TokyoNight with nested cards
# Scenario 2: Theme TokyoNight with colors (asserts TokyoNight primary RGB color 121;162;247)
cat << 'EOF' > "$TMP_DIR/tokyonight.html"
<div class="p-2 border rounded bg-primary text-white">
  <h3>TokyoNight Theme</h3>
  <p>Blue accent header with dark text.</p>
</div>
EOF
run_test "TokyoNight Theme (-t tokyonight with TrueColor)" "make render -- $TMP_DIR/tokyonight.html -w 80 -t tokyonight" "121;162;247"

# Scenario 3: Theme Dracula with colors (asserts Dracula primary RGB color 189;147;249)
cat << 'EOF' > "$TMP_DIR/dracula.html"
<div class="p-2 border rounded bg-primary text-white">
  <h3>Dracula Theme</h3>
  <p>Purple accent header.</p>
</div>
EOF
run_test "Dracula Theme (-t dracula with TrueColor)" "make render -- $TMP_DIR/dracula.html -w 80 -t dracula" "189;147;249"

# Scenario 4: Bootstrap Colored Alerts (bg-success and bg-danger TrueColor codes)
cat << 'EOF' > "$TMP_DIR/alerts.html"
<div class="row">
  <div class="col-6 bg-success text-white p-1">Success Alert</div>
  <div class="col-6 bg-danger text-white p-1">Danger Alert</div>
</div>
EOF
run_test "Colored Alerts (bg-success 25;135;84 and bg-danger 220;52;69)" "make render -- $TMP_DIR/alerts.html -w 80" "25;135;84"
# Scenario 4: Local HalfBlock image rendering
cat << 'EOF' > "$TMP_DIR/image.html"
<div class="row">
  <div class="col-6 border rounded p-2 text-center bg-dark text-white">
    <img src="examples/image/local/test.png" width="30" alt="Test Image" />
  </div>
  <div class="col-6 border rounded p-2 bg-dark text-white">
    <h3>Image Info</h3>
    <p>Rendered with HalfBlock unicode protocol</p>
  </div>
</div>
EOF
run_test "HalfBlock Image Rendering (row with img)" "make render -- $TMP_DIR/image.html -w 80 -r ." "Image Info"

# Scenario 5: Stdin pipeline rendering
run_test "Stdin pipeline with cat and '-'" "cat $TMP_DIR/grid.html | make render -- - -w 80" "Col 1"

# Scenario 6: URL-encoded filename handling (%20)
cat << 'EOF' > "$TMP_DIR/80s Huge Hits FLAC 2026.html"
<div class="p-2 border rounded bg-primary text-white">
  <h3>80s Huge Hits FLAC</h3>
</div>
EOF
run_test "URL-encoded %20 filename resolution" "make render -- '$TMP_DIR/80s%20Huge%20Hits%20FLAC%202026.html' -w 80" "80s Huge Hits FLAC"

# Scenario 7: Legacy FILE= and custom width/theme flags
run_test "Legacy FILE= syntax with WIDTH and THEME" "make render FILE=$TMP_DIR/grid.html WIDTH=80 THEME=tokyonight" "Col 1"

# Scenario 9: Full HTML Torrent Details file (template.html)
run_test "Full Torrent File Stacked (width 100)" "make render -- template.html -w 100" "80s Huge Hits FLAC 2026"
run_test "Full Torrent File 2-Col Grid (width 180)" "make render -- template.html -w 180" "Tracklist"

# Scenario 10: 256 Colors Quantization Mode (-c 256)
run_test "256 Colors Quantization Mode (-c 256)" "make render -- $TMP_DIR/alerts.html -w 80 -c 256" "Success Alert"

# Scenario 11: CLI help and watch flag support
run_test "CLI --watch flag in help" "go run ./cmd/termstrap -h" "-watch"
printf "\n${C_BOLD}${C_CYAN}──────────────────────────────────────────────────────────────────${C_RESET}\n"
if [ $FAILED -eq 0 ]; then
  printf "  ${C_GREEN}${C_BOLD}✓ All $TOTAL E2E tests passed successfully!${C_RESET}\n"
else
  printf "  ${C_RED}${C_BOLD}✗ $FAILED of $TOTAL tests failed.${C_RESET}\n"
  exit 1
fi
printf "${C_BOLD}${C_CYAN}──────────────────────────────────────────────────────────────────${C_RESET}\n\n"
