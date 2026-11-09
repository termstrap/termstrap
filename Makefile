.DEFAULT_GOAL := help

.PHONY: help test test-v test-e2e test-parity \
       examples examples-local example \
       check render view

# Colors for terminal output
C_CYAN   := \033[1;36m
C_GREEN  := \033[1;32m
C_YELLOW := \033[1;33m
C_RESET  := \033[0m

help:
	@printf '$(C_CYAN)Termstrap (Native PHP) Makefile commands:$(C_RESET)\n\n'
	@printf '  $(C_GREEN)make examples$(C_RESET)             Run all examples with terminal output\n'
	@printf '  $(C_GREEN)make examples-local$(C_RESET)       Run local examples (no network needed)\n'
	@printf '  $(C_GREEN)make example NAME=<name>$(C_RESET)  Run a single example (e.g. NAME=borders)\n'
	@printf '  $(C_GREEN)make render FILE=<file>$(C_RESET)   Render an HTML/Markdown file in terminal\n'
	@printf '  $(C_GREEN)make test$(C_RESET)                 Run PHPUnit test suite\n'
	@printf '  $(C_GREEN)make test-parity$(C_RESET)          Run Go differential parity tests\n'
	@printf '  $(C_GREEN)make check$(C_RESET)                Validate composer and run tests\n\n'

# --- Examples ---

EXAMPLES_LIST := ascii ascii-all-fonts bootstrap borders breakpoints custom-css fiche glamour headings html lipgloss-features media-queries nested nested-no-md shadows spacing spinner-shimmer styling syntax syntax-custom syntax-themes themes torrent

examples: examples-local

examples-local:
	@for ex in $(EXAMPLES_LIST); do \
		printf '\n$(C_CYAN)════════════════════════════════════════════════════════════════════════════════$(C_RESET)\n'; \
		printf '$(C_GREEN) ═══ examples/%s ═══$(C_RESET)\n' "$$ex"; \
		printf '$(C_CYAN)════════════════════════════════════════════════════════════════════════════════$(C_RESET)\n\n'; \
		php examples/run.php "$$ex"; \
	done

# Run a single example: make example NAME=borders
example:
	@if [ -z "$(NAME)" ]; then \
		printf '$(C_YELLOW)Error: Please specify example name: make example NAME=borders$(C_RESET)\n'; \
		printf 'Available examples:\n'; \
		php examples/run.php --list; \
		exit 1; \
	fi
	@php examples/run.php $(NAME)

# --- Render CLI ---

FILE ?=
WIDTH ?=
THEME ?=
ROOT ?=
FLAGS ?=

render:
	@if [ -n "$(FILE)" ]; then \
		php bin/termstrap "$(FILE)" \
			$(if $(WIDTH),-w $(WIDTH)) \
			$(if $(THEME),-t $(THEME)) \
			$(if $(ROOT),-r $(ROOT)) \
			$(FLAGS); \
	elif [ ! -t 0 ]; then \
		php bin/termstrap -; \
	else \
		printf '$(C_YELLOW)Error: Please specify a file$(C_RESET)\n\n'; \
		printf 'Usage:\n'; \
		printf '  $(C_CYAN)make render FILE=<path> [WIDTH=100] [THEME=dracula]$(C_RESET)\n\n'; \
		exit 1; \
	fi

view: render

# --- Tests ---

test:
	@vendor/bin/phpunit

test-v:
	@vendor/bin/phpunit --testdox

test-parity:
	@vendor/bin/phpunit tests/DifferentialParityTest.php

test-e2e:
	@vendor/bin/phpunit tests/E2eExamplesTest.php

check:
	@composer validate --no-check-publish
	@vendor/bin/phpunit
