# Makefile — automatiza build/test/dev do cpcli (CLI) e do app desktop (Wails).
#
# O módulo desktop/ é um módulo Go separado (replace cpcli => ../) porque
# depende do runtime do Wails + CGO/webkit2gtk; por isso os alvos "desktop-*"
# e "frontend-*" sempre entram na pasta desktop/ antes de rodar o comando.
#
# WEBKIT_TAG: esta e outras distros recentes (Fedora 40+, etc.) só têm
# webkit2gtk-4.1 — o Wails procura a 4.0 por padrão. Veja desktop/README.md.

WEBKIT_TAG := webkit2_41

.DEFAULT_GOAL := help

.PHONY: help build test vet tidy install \
        desktop-tidy desktop-generate desktop-build desktop-dev \
        desktop-build-check desktop-vet \
        frontend-install frontend-build frontend-typecheck frontend-dev \
        all clean

help: ## Lista os alvos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## --- cpcli (core / CLI) -------------------------------------------------------

build: ## Compila o cpcli (binário CLI) em ./cpcli
	go build -o cpcli ./cmd/cpcli

test: ## Roda os testes do módulo core
	go test ./...

vet: ## go vet no módulo core
	go vet ./...

tidy: ## go mod tidy no módulo core
	go mod tidy

install: ## Instala o cpcli via go install (fica em $$GOPATH/bin)
	go install ./cmd/cpcli

## --- desktop (Wails) ----------------------------------------------------------

desktop-tidy: ## go mod tidy no módulo desktop
	cd desktop && go mod tidy

desktop-generate: ## Regenera os bindings TS do Wails (desktop/frontend/wailsjs)
	cd desktop && wails generate module -tags $(WEBKIT_TAG)

desktop-build: frontend-install ## Compila o app desktop (React embutido no binário Go)
	mkdir -p desktop/frontend/dist && touch desktop/frontend/dist/.gitkeep
	cd desktop && wails build -tags $(WEBKIT_TAG)

desktop-dev: ## Sobe o app desktop em modo dev (hot-reload via Vite)
	cd desktop && wails dev -tags $(WEBKIT_TAG)

desktop-build-check: ## go build (sem empacotar) do módulo desktop — checagem rápida
	cd desktop && go build -tags $(WEBKIT_TAG) ./...

desktop-vet: ## go vet do módulo desktop
	cd desktop && go vet -tags $(WEBKIT_TAG) ./...

## --- frontend (Vite/React, dentro de desktop/frontend) -------------------------

frontend-install: ## npm install do frontend
	cd desktop/frontend && npm install

frontend-build: frontend-install ## Build de produção do frontend (gera desktop/frontend/dist)
	cd desktop/frontend && npm run build

frontend-typecheck: frontend-install ## tsc --noEmit do frontend
	cd desktop/frontend && npm run typecheck

frontend-dev: frontend-install ## Vite dev server isolado (mock service, sem shell nativo)
	cd desktop/frontend && npm run dev

## --- geral ----------------------------------------------------------------------

all: build desktop-build ## Compila o CLI e o app desktop

clean: ## Remove artefatos de build (cpcli, desktop/build, dist gerado)
	rm -f cpcli desktop/desktop
	rm -rf desktop/build/bin
	rm -rf desktop/frontend/dist
	mkdir -p desktop/frontend/dist
	touch desktop/frontend/dist/.gitkeep
