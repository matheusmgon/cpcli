# cpcli desktop (Wails)

Cliente **desktop** para o Check Point Management API — uma UI gráfica sobre o
mesmo core do `cpcli`, com o objetivo de oferecer em Linux/macOS a experiência
que o SmartConsole dá no Windows.

Este é um **módulo Go separado** do core (`module cpcli/desktop`, com
`replace cpcli => ../`). Isso mantém o `go build ./...` da raiz sempre compilável
mesmo sem as dependências de GUI — só este módulo precisa delas.

## Arquitetura

```
desktop/main.go          → app Wails; binda service.Service
desktop/frontend/        → Vite + React + TypeScript + Tailwind v4 + shadcn/ui
  src/lib/wailsService.ts  → interface CpService; usa os bindings reais do
                              Wails ou cai num mock (fora do shell nativo)
  src/lib/mockService.ts   → mock com dados de exemplo, usado por vite dev/preview
  wailsjs/                 → bindings gerados (wails generate module), gitignored
  dist/                    → build do Vite, embedado via go:embed (gitignored,
                              exceto dist/.gitkeep — placeholder pro embed não
                              quebrar num checkout limpo antes do 1º build)
        │  window.go.service.Service.*
        ▼
cpcli/service            → fachada de UI (Login/ListObjects/Publish/…)
        ▼
cpcli/internal/mgmt      → core de transporte/sessão (compartilhado com o CLI)
```

Direção visual: paleta azul-marinho corporativo (não o laranja/vermelho da
versão inicial), sidebar escura fixa nos dois temas claro/escuro, ícones
lucide-react, dashboard de overview logo após o login. Decisões completas de
design em `/home/matheus/.claude/plans/abundant-sauteeing-snowflake.md`.

### Regenerar os bindings do Wails

Sempre que a assinatura de algum método de `service.Service` mudar:

```sh
cd desktop
wails generate module -tags webkit2_41   # a tag vale aqui também — o comando compila o projeto
```

Reescreve `frontend/wailsjs/go/service/Service.{js,d.ts}` e
`frontend/wailsjs/go/models.ts`. Idempotente — pode rodar quantas vezes quiser.

### Rodar o frontend fora do shell nativo (iteração rápida)

```sh
cd desktop/frontend
npm install
npm run dev       # vite dev server comum — usa o mock service (sem window.go)
npm run build     # gera dist/, o que o Go embeda
npm run preview   # serve o build de produção localmente (também em mock)
```

## Pré-requisitos

- Go 1.25+
- Node 20+ (testado com Node 22) e npm
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

Distros modernas (Fedora 40+, etc.) só têm **webkit2gtk-4.1**; o Wails procura a
4.0 por padrão. Passe a build tag `webkit2_41` para usar a 4.1 — isso vale
também para `go build`/`go vet`/`wails generate module` neste módulo, não só
`dev`/`build`:

```sh
cd desktop
go mod tidy                    # resolve/pina a versão do Wails e o replace do core
wails dev -tags webkit2_41     # hot-reload de verdade (frontend:dev:serverUrl: auto)
wails build -tags webkit2_41   # binário de produção em build/bin/ (frontend embedado)
```

Em sistemas que ainda tenham a webkit2gtk-4.0, a tag é dispensável
(`wails dev` / `wails build`).

**Requisito mínimo de WebKitGTK**: Tailwind v4 usa recursos de CSS (cascade
layers, `@property`, `color-mix()`) que exigem WebKitGTK ~2.42+ (webkit2gtk-4.1
em distros recentes). Distros muito antigas com webkit2gtk-4.0 (ex: Ubuntu
22.04, WebKitGTK ~2.36) podem renderizar o tema quebrado.

## Segurança (npm audit)

`npm audit` acusa 1 advisory de severidade alta em `react-router`/`react-router-dom`
(GHSA-qwww-vcr4-c8h2, CSRF bypass em **modo RSC**). Não se aplica aqui — este é
um app client-side puro (Vite SPA num webview, `HashRouter`, sem framework
mode/SSR/server actions), que é exatamente o modo que não é afetado. Não há
correção publicada ainda dentro da faixa (ships na 8.3.0, ainda não lançada);
versões mais antigas (`<7.12.0`) evitam esse advisory mas reintroduzem 4
outros (open redirect XSS, injeção via SSR hydration, DoS de rota) já
corrigidos nas versões atuais — por isso a escolha é ficar na mais recente.

## Estado

Cobre a superfície completa do `cpcli/service` hoje: login, dashboard de
overview, CRUD dos 8 tipos de objeto simples (hosts, networks, groups,
services TCP/UDP, address ranges, service groups, access roles), Access
Control rules, NAT, Threat Prevention (regras + profiles), HTTPS Inspection,
VPN (star/meshed), install/verify de política, leitura de gateways/packages,
e anti-spoofing por interface de gateway — sempre com publish/discard
explícitos, como no SmartConsole real. A sidebar segue a mesma agrupação do
SmartConsole (Access Control + NAT + Threat Prevention + HTTPS Inspection sob
"Security Policies", em vez de itens soltos).
