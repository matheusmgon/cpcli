# cpcli desktop (Wails)

**Desktop** client for the Check Point Management API — a graphical UI on
top of the same core as `cpcli`, aiming to offer on Linux/macOS the
experience SmartConsole gives on Windows.

This is a **separate Go module** from the core (`module cpcli/desktop`,
with `replace cpcli => ../`). That way `go build ./...` at the repo root
stays compilable even without the GUI dependencies — only this module
needs them.

## Architecture

```
desktop/main.go          → Wails app; binds service.Service
desktop/frontend/        → Vite + React + TypeScript + Tailwind v4 + shadcn/ui
  src/lib/wailsService.ts  → CpService interface; uses real Wails
                              bindings or falls back to a mock (outside
                              the native shell)
  src/lib/mockService.ts   → mock with sample data, used by vite dev/preview
  wailsjs/                 → generated bindings (wails generate module), gitignored
  dist/                    → Vite build output, embedded via go:embed
                              (gitignored, except dist/.gitkeep — placeholder
                              so the embed does not fail on a fresh checkout
                              before the first build)
        │  window.go.service.Service.*
        ▼
cpcli/service            → UI-facing facade (Login/ListObjects/Publish/…)
        ▼
cpcli/internal/mgmt      → transport/session core (shared with the CLI)
```

Visual direction: corporate navy-blue palette, dark sidebar fixed in
both light and dark themes, lucide-react icons, overview dashboard right
after login.

### Regenerate the Wails bindings

Whenever a `service.Service` method signature changes:

```sh
cd desktop
wails generate module -tags webkit2_41   # the tag applies here too — the command compiles the project
```

Rewrites `frontend/wailsjs/go/service/Service.{js,d.ts}` and
`frontend/wailsjs/go/models.ts`. Idempotent — safe to run any number of
times.

### Run the frontend outside the native shell (fast iteration)

```sh
cd desktop/frontend
npm install
npm run dev       # plain vite dev server — uses the mock service (no window.go)
npm run build     # generates dist/, which Go embeds
npm run preview   # serves the production build locally (also mocked)
```

## Requirements

- Go 1.25+
- Node 20+ (tested with Node 22) and npm
- **Wails CLI:**
  ```sh
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- **System dependencies (Linux/Fedora)** — the Wails WebView:
  ```sh
  sudo dnf install gtk3-devel webkit2gtk4.1-devel
  ```
  (Debian/Ubuntu: `sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev`)
- Run `wails doctor` to check the environment.

## Build & run

Modern distros (Fedora 40+, etc.) only ship **webkit2gtk-4.1**; Wails
looks for 4.0 by default. Pass the `webkit2_41` build tag to use 4.1 —
this applies to `go build`/`go vet`/`wails generate module` in this
module too, not just `dev`/`build`:

```sh
cd desktop
go mod tidy                    # resolves/pins the Wails version and the core replace
wails dev -tags webkit2_41     # real hot-reload (frontend:dev:serverUrl: auto)
wails build -tags webkit2_41   # production binary in build/bin/ (frontend embedded)
```

On systems that still have webkit2gtk-4.0, the tag is optional
(`wails dev` / `wails build`).

**Minimum WebKitGTK requirement:** Tailwind v4 uses CSS features (cascade
layers, `@property`, `color-mix()`) that require WebKitGTK ~2.42+
(webkit2gtk-4.1 on recent distros). Very old distros with webkit2gtk-4.0
(e.g. Ubuntu 22.04, WebKitGTK ~2.36) may render the theme broken.

## Security (npm audit)

`npm audit` reports 1 high-severity advisory in
`react-router`/`react-router-dom` (GHSA-qwww-vcr4-c8h2, CSRF bypass in
**RSC mode**). It does not apply here — this is a pure client-side app
(Vite SPA in a webview, `HashRouter`, no framework mode / SSR / server
actions), which is exactly the mode that is not affected. No published
fix within the range yet (ships in 8.3.0, not released); older versions
(`<7.12.0`) avoid this advisory but reintroduce 4 others (open redirect
XSS, SSR hydration injection, route DoS) already fixed in current
versions — so the choice is to stay on the latest.

## Status

Covers the full `cpcli/service` surface today: login, overview
dashboard, CRUD for the 8 simple object types (hosts, networks, groups,
services TCP/UDP, address ranges, service groups, access roles), Access
Control rules, NAT, Threat Prevention (rules + profiles), HTTPS
Inspection, VPN (star/meshed), install/verify policy, gateway/package
listings, and per-interface anti-spoofing on gateways — always with
explicit publish/discard, like the real SmartConsole. The sidebar
follows the same grouping as SmartConsole (Access Control + NAT + Threat
Prevention + HTTPS Inspection under "Security Policies", instead of
scattered top-level items).
