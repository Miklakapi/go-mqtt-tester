SHELL := /usr/bin/bash
.ONESHELL:
.SHELLFLAGS := -euo pipefail -c

ifneq ("$(wildcard .env)","")
	include .env
	export $(shell sed -n 's/^\([^#][A-Za-z0-9_]*\)=.*/\1/p' .env)
endif

GO ?= go

.PHONY: help run build

help:
	@echo "Targets:"
	@echo "  make run"
	@echo "  make build"

run:
	$(GO) run ./cmd/main.go

build:
	go build -o ./bin/app ./cmd/main.go
