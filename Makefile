\
# Simple process management for local development — no Docker needed for a
# two-process (Go + Vite) local app. PIDs/logs are tracked under .run/ so
# `make stop` can reliably find and kill exactly what `make run` started.
#
# All paths below are absolute (built from $(CURDIR)) on purpose: recipe
# lines that background a job (`cmd &`) run that job in its own sub-shell,
# so a `cd` done as part of the backgrounded command does NOT affect the
# cwd of whatever runs next in the same recipe line. Relative paths after
# a backgrounded `cd` are therefore unreliable; absolute paths sidestep the
# issue entirely.

ROOT_DIR     := $(CURDIR)
BACKEND_DIR  := $(ROOT_DIR)/backend
FRONTEND_DIR := $(ROOT_DIR)/frontend
RUN_DIR      := $(ROOT_DIR)/.run

BACKEND_BIN  := $(BACKEND_DIR)/bin/server
BACKEND_PID  := $(RUN_DIR)/backend.pid
FRONTEND_PID := $(RUN_DIR)/frontend.pid
BACKEND_LOG  := $(RUN_DIR)/backend.log
FRONTEND_LOG := $(RUN_DIR)/frontend.log

# Where imported EPUBs are kept. Unset disables persistence entirely, which
# is the right default for tests and for a read-only container; `make run`
# points it at a working directory so books survive a restart. Override it
# from the environment to keep a library somewhere else.
LIBRARY_DIR  ?= $(ROOT_DIR)/data/library

# Source extracts for `make lexicon`, which are ~1 GB each and deliberately
# not committed. The distilled tables they produce are.
LEXICON_SRC  := $(ROOT_DIR)/.lexicon-src
LEXICON_LANGS := fr:French es:Spanish de:German

.PHONY: help install build test run stop up down restart logs clean \
	backend-build frontend-install lexicon

help:
	@echo "Targets:"
	@echo "  make install    Install frontend dependencies (npm install)"
	@echo "  make build      Build backend binary + frontend production bundle"
	@echo "  make test       Run backend Go tests"
	@echo "  make run        Start backend (:8080) and frontend (:5173) in the background"
	@echo "  make stop       Stop whatever 'make run' started"
	@echo "  make restart    make stop && make run"
	@echo "  make logs       Tail backend + frontend logs (Ctrl+C to stop watching)"
	@echo "  make clean      Stop servers and remove build/log artifacts"
	@echo "  make lexicon    Re-download Wiktionary extracts and regenerate"
	@echo "                  the reading dictionaries (slow; rarely needed)"
	@echo ""
	@echo "Aliases: make up == make run, make down == make stop"
	@echo ""
	@echo "Imported EPUBs are stored under $(LIBRARY_DIR)"
	@echo "(override with LIBRARY_DIR=... ; unset SERMOKEY_LIBRARY_DIR to disable)"

install: frontend-install

frontend-install: $(FRONTEND_DIR)/node_modules

$(FRONTEND_DIR)/node_modules:
	cd $(FRONTEND_DIR) && npm install

backend-build:
	cd $(BACKEND_DIR) && go build -o bin/server ./cmd/server

test:
	cd $(BACKEND_DIR) && go test ./...

build: backend-build
	cd $(FRONTEND_DIR) && npm run build

$(RUN_DIR):
	mkdir -p $(RUN_DIR)

run up: backend-build frontend-install $(RUN_DIR)
	@mkdir -p $(LIBRARY_DIR)
	@if [ -f $(BACKEND_PID) ] && kill -0 "$$(cat $(BACKEND_PID))" 2>/dev/null; then \
		echo "backend already running (pid $$(cat $(BACKEND_PID)))"; \
	else \
		echo "starting backend on :8080 ..."; \
		set -m; ( SERMOKEY_LIBRARY_DIR=$(LIBRARY_DIR) nohup $(BACKEND_BIN) < /dev/null > $(BACKEND_LOG) 2>&1 & echo $$! > $(BACKEND_PID) ); \
	fi
	@if [ -f $(FRONTEND_PID) ] && kill -0 "$$(cat $(FRONTEND_PID))" 2>/dev/null; then \
		echo "frontend already running (pid $$(cat $(FRONTEND_PID)))"; \
	else \
		echo "starting frontend on :5173 ..."; \
		set -m; ( cd $(FRONTEND_DIR) && nohup npx vite < /dev/null > $(FRONTEND_LOG) 2>&1 & echo $$! > $(FRONTEND_PID) ); \
	fi
	@sleep 1
	@echo ""
	@echo "Backend:  http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	@echo "Logs: make logs   |   Stop: make stop"

stop down:
	@if [ -f $(FRONTEND_PID) ]; then \
		pid=$$(cat $(FRONTEND_PID)); \
		if kill -0 "$$pid" 2>/dev/null; then kill "$$pid" && echo "stopped frontend (pid $$pid)"; fi; \
		rm -f $(FRONTEND_PID); \
	fi
	@if [ -f $(BACKEND_PID) ]; then \
		pid=$$(cat $(BACKEND_PID)); \
		if kill -0 "$$pid" 2>/dev/null; then kill "$$pid" && echo "stopped backend (pid $$pid)"; fi; \
		rm -f $(BACKEND_PID); \
	fi

restart: stop run

logs:
	@touch $(BACKEND_LOG) $(FRONTEND_LOG)
	@tail -f $(BACKEND_LOG) $(FRONTEND_LOG)

clean: stop
	rm -rf $(RUN_DIR) $(BACKEND_DIR)/bin $(FRONTEND_DIR)/dist

# Regenerates the embedded reading dictionaries from English Wiktionary.
#
# Only needed when refreshing the data or adding a language — the distilled
# tables are committed so an ordinary build never downloads anything. The
# extracts are ~1 GB per language, so they are cached under .lexicon-src/
# and the download is skipped if a file is already there.
#
# Source: English Wiktionary via wiktextract, dual-licensed CC BY-SA and
# GFDL. The attribution that requires is rendered in the reader's word
# panel.
lexicon:
	@mkdir -p $(LEXICON_SRC)
	@for pair in $(LEXICON_LANGS); do \
		code=$${pair%%:*}; name=$${pair##*:}; \
		src=$(LEXICON_SRC)/$$code.jsonl; \
		if [ -s "$$src" ]; then \
			echo "using cached $$src"; \
		else \
			echo "downloading $$name extract (~1 GB) ..."; \
			curl -fL --retry 3 -o "$$src" \
				"https://kaikki.org/dictionary/$$name/kaikki.org-dictionary-$$name.jsonl" || exit 1; \
		fi; \
		echo "generating $$code lexicon ..."; \
		( cd $(BACKEND_DIR) && go run ./cmd/lexgen \
			-in "$$src" -out internal/reading/$$code/data ) || exit 1; \
	done
	@echo ""
	@echo "Regenerated. Run 'make test' to check the corpora still validate."
