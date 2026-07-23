# cpcli desktop (Wails)

Cliente **desktop** para o Check Point Management API — uma UI gráfica sobre o
mesmo core do `cpcli`, com o objetivo de oferecer em Linux/macOS a experiência
que o SmartConsole dá no Windows.

Este é um **módulo Go separado** do core (`module cpcli/desktop`, com
`replace cpcli => ../`). Isso mantém o `go build ./...` da raiz sempre compilável
mesmo sem as dependências de GUI — só este módulo precisa delas.

## Arquitetura

```
desktop/main.go        → app Wails; binda service.Service
desktop/frontend/dist  → UI estática (login + CRUD de hosts), embedada via go:embed
        │  window.go.service.Service.*
        ▼
cpcli/service          → fachada de UI (Login/ListHosts/AddHost/Publish/…)
        ▼
cpcli/internal/mgmt    → core de transporte/sessão (compartilhado com o CLI)
```

O frontend atual é HTML/CSS/JS estático (sem passo de build). Dá pra trocar por
Vite/React/Svelte depois; basta apontar o `frontend:*` do `wails.json` e o
`//go:embed` para o `dist` gerado.

## Pré-requisitos

- Go 1.25+
- **Wails CLI:**
  ```sh
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **Dependências de sistema (Linux/Fedora)** — a WebView do Wails:
  ```sh
  sudo dnf install gtk3-devel webkit2gtk4.1-devel
  ```
  (Debian/Ubuntu: `sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev`)
- Rode `wails doctor` para confirmar que o ambiente está completo.

## Build & run

```sh
cd desktop
go mod tidy        # resolve/pina a versão do Wails e o replace do core
wails dev          # hot-reload durante o desenvolvimento
wails build        # binário de produção em build/bin/
```

## Estado

Scaffold inicial. Cobre: login (com sessão em memória), listar/filtrar hosts,
criar e apagar host, publish e discard. Próximos passos naturais: networks,
groups, services, regras de Access Control e install-policy — todos seguindo o
mesmo padrão, adicionando métodos em `cpcli/service` e telas no frontend.
