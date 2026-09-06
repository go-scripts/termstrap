.DEFAULT_GOAL := help

.PHONY: help test test-v test-short test-cover test-cover-html \
       test-unit test-layout test-shadow test-image test-grid test-utils \
       test-e2e test-e2e-v test-e2e-output \
       examples examples-local examples-network example \
       test-examples test-examples-short \
       build check ci lint vet clean render view

# --- Argument passthrough for make <cmd> -- <args> ---
ifeq ($(firstword $(MAKECMDGOALS)),$(filter $(firstword $(MAKECMDGOALS)),render view))
  CMD_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  ifeq ($(firstword $(CMD_ARGS)),--)
    CMD_ARGS := $(wordlist 2,$(words $(CMD_ARGS)),$(CMD_ARGS))
  endif
  $(foreach target,$(MAKECMDGOALS),$(if $(filter-out render view,$(target)),$(eval $(target): _force;@:)))
  _force:
  %:
	@:
endif

# --- Render / CLI ---

FILE ?= $(filepath)
WIDTH ?=
THEME ?=
ROOT ?=
FLAGS ?=

render:
	@if [ -n "$(CMD_ARGS)" ]; then \
		go run ./cmd/termstrap $(CMD_ARGS); \
	elif [ -n "$(FILE)" ]; then \
		go run ./cmd/termstrap \
			$(if $(WIDTH),-w $(WIDTH)) \
			$(if $(THEME),-t $(THEME)) \
			$(if $(ROOT),-r $(ROOT)) \
			$(FLAGS) \
			"$(FILE)"; \
	else \
		printf '$(C_YELLOW)Error: Please specify a file or arguments$(C_RESET)\n\n'; \
		printf 'Usage:\n'; \
		printf '  $(C_CYAN)make render -- <file> [options]$(C_RESET)\n'; \
		printf '  $(C_CYAN)make render FILE=<path> [WIDTH=100] [THEME=dracula]$(C_RESET)\n\n'; \
		printf 'Examples:\n'; \
		printf '  make render -- page.html\n'; \
		printf '  make render -- page.html -w 100 -t tokyonight\n'; \
		printf '  make render -- - < page.html\n'; \
		printf '  make render FILE=page.html WIDTH=100 THEME=dracula\n'; \
		exit 1; \
	fi

view: render
# --- Build ---

build:
	go build ./...

vet:
	go vet ./...

lint: vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

# --- Tests (all) ---

test:
	go test ./... -count=1

test-e2e:
	@./examples/cli/test_e2e.sh

test-e2e-v:
	@./examples/cli/test_e2e.sh -v
test-e2e-output: test-e2e-v

test-v:
	go test ./... -count=1 -v

test-short:
	go test ./... -count=1 -short

test-cover:
	go test ./... -count=1 -coverprofile=cover.out
	go tool cover -func=cover.out | tail -1
	@rm -f cover.out

test-cover-html:
	go test ./... -count=1 -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
	open cover.html
	@rm -f cover.out

test-all: test-unit test-layout test-shadow test-image test-grid test-utils test-deferred

# --- Tests (granular) ---

test-unit:
	go test -v -count=1 -run 'Test(HexToRGB|TrimBlank|PersistColors|AddPadding|WrapLong|StripANSI)' .

test-layout:
	go test -v -count=1 -run 'Test(Border|NoOverflow|ThreeColumn|SingleRow|SideBySide|ColumnWidth|RowLevel|BackgroundColor|Layout_Multiple|ImagePlaceholder)' .

test-shadow:
	go test -v -count=1 -run 'Test(Shadow|CalculateShadow|ApplyShadow)' .

test-image:
	go test -v -count=1 ./image/

test-grid:
	go test -v -count=1 -run 'Test(ResolveCol|ParseGrid|Breakpoint|DetectBreakpoint)' .

test-utils:
	go test -v -count=1 -run 'Test(ExtractSegments|ParseHTML|ResolveStyle|Nested)' .

test-deferred:
	go test -v -count=1 -run 'TestDeferredOverlay' .

# --- Examples ---

EXAMPLES_LOCAL  = breakpoints styling nested borders shadows image/local image/three-columns nested-no-md
EXAMPLES_NET    = . image/detect image/formats image/grid image/markdown image/protocols
EXAMPLES_ALL    = $(EXAMPLES_LOCAL) $(EXAMPLES_NET)

examples: examples-local examples-network

examples-local:
	@for dir in $(EXAMPLES_LOCAL); do \
		printf '\n\033[1;36m═══ examples/%s ═══\033[0m\n' "$$dir"; \
		TERMSTRAP_IMAGE_PROTOCOL=halfblock go run ./examples/$$dir/; \
	done

examples-network:
	@for dir in $(EXAMPLES_NET); do \
		printf '\n\033[1;36m═══ examples/%s ═══\033[0m\n' "$$dir"; \
		TERMSTRAP_IMAGE_PROTOCOL=halfblock go run ./examples/$$dir/; \
	done

# Run a single example: make example NAME=borders
example:
	@TERMSTRAP_IMAGE_PROTOCOL=halfblock go run ./examples/$(NAME)/

# --- Examples via go test ---

test-examples:
	go test -v -count=1 -timeout 300s ./examples/

test-examples-short:
	go test -v -count=1 -short -timeout 120s ./examples/

# --- CI / Check ---

check: build vet test
	@printf '\n\033[1;32m✓ All checks passed.\033[0m\n'

ci: build vet test
	@printf '\n\033[1;32m✓ CI passed.\033[0m\n'

# --- Cleanup ---

clean:
	@rm -f cover.out cover.html

# --- Help ---

# Colors
C_RESET  = \033[0m
C_BOLD   = \033[1m
C_CYAN   = \033[1;36m
C_GREEN  = \033[1;32m
C_YELLOW = \033[1;33m
C_DIM    = \033[2m
C_WHITE  = \033[1;37m
C_MAG    = \033[1;35m

help:
	@printf '\n'
	@printf '  $(C_CYAN)╔══════════════════════════════════════════════════════════════════╗$(C_RESET)\n'
	@printf '  $(C_CYAN)║$(C_RESET)  $(C_BOLD)$(C_WHITE)termstrap$(C_RESET) — Bootstrap-like layout for the terminal          $(C_CYAN)║$(C_RESET)\n'
	@printf '  $(C_CYAN)╚══════════════════════════════════════════════════════════════════╝$(C_RESET)\n'
	@printf '\n'
	@printf '  $(C_GREEN)RENDER & CLI$(C_RESET)\n'
	@printf '  $(C_CYAN)make render FILE=<path>$(C_RESET)   Render an HTML file in terminal (e.g. WIDTH=100 THEME=tokyonight)\n'
	@printf '  $(C_CYAN)make view FILE=<path>$(C_RESET)     Alias for make render\n'
	@printf '\n'
	@printf '  $(C_GREEN)BUILD & LINT$(C_RESET)\n'
	@printf '  $(C_CYAN)make build$(C_RESET)               Compile all packages                          \n'
	@printf '  $(C_CYAN)make vet$(C_RESET)                 Run go vet static analysis                    \n'
	@printf '  $(C_CYAN)make lint$(C_RESET)                Run vet + staticcheck (if installed)           \n'
	@printf '\n'
	@printf '  $(C_GREEN)TESTS — ALL$(C_RESET)\n'
	@printf '  $(C_CYAN)make test$(C_RESET)                Run all tests                                 \n'
	@printf '  $(C_CYAN)make test-e2e$(C_RESET)            Run End-to-End bash test suite for make render\n'
	@printf '  $(C_CYAN)make test-e2e-v$(C_RESET)          Run E2E test suite with full visual render output\n'
	@printf '  $(C_CYAN)make test-e2e-output$(C_RESET)     Alias for make test-e2e-v\n'
	@printf '  $(C_CYAN)make test-v$(C_RESET)              Run all tests (verbose)                       \n'
	@printf '  $(C_CYAN)make test-short$(C_RESET)          Run all tests, skip network examples          \n'
	@printf '  $(C_CYAN)make test-cover$(C_RESET)          Run tests + print coverage summary            \n'
	@printf '  $(C_CYAN)make test-cover-html$(C_RESET)     Run tests + open HTML coverage in browser     \n'
	@printf '\n'
	@printf '  $(C_GREEN)TESTS — GRANULAR$(C_RESET)\n'
	@printf '  $(C_CYAN)make test-unit$(C_RESET)           Utils: hexToRGB, stripANSI, trimBlankLines…   \n'
	@printf '  $(C_CYAN)make test-layout$(C_RESET)         Layout: borders, overflow, column widths      \n'
	@printf '  $(C_CYAN)make test-shadow$(C_RESET)         Shadows: metrics, rendering, overflow         \n'
	@printf '  $(C_CYAN)make test-image$(C_RESET)          Image subpackage: protocols, resize, detect   \n'
	@printf '  $(C_CYAN)make test-grid$(C_RESET)           Grid: parseGrid, breakpoints, resolveColSpan  \n'
	@printf '  $(C_CYAN)make test-utils$(C_RESET)          Parsing: extractSegments, resolveStyle, nested\n'
	@printf '  $(C_CYAN)make test-deferred$(C_RESET)       Deferred overlay: cursor sequences, fallback  \n'
	@printf '\n'
	@printf '  $(C_GREEN)EXAMPLES — RUN WITH OUTPUT$(C_RESET)\n'
	@printf '  $(C_CYAN)make examples$(C_RESET)            Run all examples (local + network)            \n'
	@printf '  $(C_CYAN)make examples-local$(C_RESET)      Run offline examples only                     \n'
	@printf '  $(C_CYAN)make examples-network$(C_RESET)    Run examples that fetch remote images         \n'
	@printf '  $(C_CYAN)make example NAME=borders$(C_RESET) Run a single example by name                 \n'
	@printf '\n'
	@printf '  $(C_GREEN)EXAMPLES — VIA GO TEST$(C_RESET)\n'
	@printf '  $(C_CYAN)make test-examples$(C_RESET)       Run all examples as tests (pass/fail)         \n'
	@printf '  $(C_CYAN)make test-examples-short$(C_RESET) Run local examples as tests (no network)     \n'
	@printf '\n'
	@printf '  $(C_GREEN)CI & WORKFLOW$(C_RESET)\n'
	@printf '  $(C_CYAN)make check$(C_RESET)               Build + vet + test (pre-commit)               \n'
	@printf '  $(C_CYAN)make ci$(C_RESET)                  Full CI pipeline: build + vet + test           \n'
	@printf '  $(C_CYAN)make clean$(C_RESET)               Remove generated files (cover.out, cover.html)\n'
	@printf '\n'
	@printf '  $(C_DIM)Available examples: breakpoints, styling, nested, borders, shadows,$(C_RESET)\n'
	@printf '  $(C_DIM)image/local, image/three-columns, image/detect, image/formats, image/grid,$(C_RESET)\n'
	@printf '  $(C_DIM)image/markdown, image/protocols$(C_RESET)\n'
	@printf '\n'
