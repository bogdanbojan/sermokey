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

.PHONY: help install build test run stop up down restart logs clean \
	backend-build frontend-install

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
	@echo ""
	@echo "Aliases: make up == make run, make down == make stop"

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
	@if [ -f $(BACKEND_PID) ] && kill -0 "$$(cat $(BACKEND_PID))" 2>/dev/null; then \
		echo "backend already running (pid $$(cat $(BACKEND_PID)))"; \
	else \
		echo "starting backend on :8080 ..."; \
		set -m; ( nohup $(BACKEND_BIN) < /dev/null > $(BACKEND_LOG) 2>&1 & echo $$! > $(BACKEND_PID) ); \
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
