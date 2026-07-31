# cpcli — Check Point SmartConsole alternative in Go

A **cross-platform** alternative to Check Point's SmartConsole, written in
Go: a **CLI** (`cpcli`) and a **desktop app** (Wails + React) to manage a
Check Point Security Management server via the **Check Point Management
API**.

Built because the official SmartConsole is Windows-only. Here you can
manage objects, rules (Access Control, NAT, Threat Prevention, HTTPS
Inspection), gateways, VPN, install policy and read logs — all on Linux
and macOS.

Built on top of Check Point's official SDK:
[`github.com/CheckPointSW/cp-mgmt-api-go-sdk`](https://github.com/CheckPointSW/cp-mgmt-api-go-sdk).

---

## Table of contents

- [Components](#components)
- [Requirements](#requirements)
- [Build](#build)
- [Quick start](#quick-start)
- [CLI guide](#cli-guide)
- [Desktop app guide](#desktop-app-guide)
- [Project layout](#project-layout)
- [Tests](#tests)
- [Contributing](#contributing)

---

## Components

| Module | Description |
|---|---|
| `cmd/cpcli` | CLI (`cpcli`) — commands for full CRUD and policy operations |
| `service/` | Bindable facade — `Service` groups UI-facing operations |
| `internal/mgmt/` | Core transport/session on top of the official SDK |
| `desktop/` | Desktop app (Wails + React + TypeScript + Tailwind) on top of the facade |

The CLI and the desktop app share the same core (`internal/mgmt`) and the
same facade (`service/`). Any feature added there shows up in both
clients.

---

## Requirements

**CLI only:**
- Go 1.25+
- Network access to a Check Point Security Management (version ≥ R80.10)

**For the desktop app on top of that:**
- Node 20+ and npm (tested with Node 22)
- [Wails CLI](https://wails.io) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- WebView system dependencies (Linux):
  ```sh
  # Fedora
  sudo dnf install gtk3-devel webkit2gtk4.1-devel
  # Debian/Ubuntu
  sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
  ```
- Run `wails doctor` to check your environment.

---

## Build

### CLI

```sh
git clone https://github.com/matheusmgon/cpcli.git
cd cpcli
go build -o cpcli ./cmd/cpcli
./cpcli --help
```

### Desktop app

```sh
cd desktop
# Distros with webkit2gtk-4.1 (Fedora 40+, Ubuntu 24.04+):
wails build -tags webkit2_41
# Distros still on webkit2gtk-4.0:
wails build
```

Binary output: `desktop/build/bin/cpcli-desktop`.

For development (real hot-reload):
```sh
cd desktop
wails dev -tags webkit2_41
```

To iterate on the frontend only (with mocked data, no firewall needed):
```sh
cd desktop/frontend
npm install
npm run dev
```

---

## Quick start

**CLI — create an internal network, an HTTPS rule and install the policy:**

```sh
# 1. Login (password via prompt or CPCLI_PASSWORD env)
./cpcli login --server 192.0.2.10 --user admin --insecure

# 2. Create a network object
./cpcli network add --name lan-internal \
    --field subnet4=10.0.10.0 --field mask-length4=24

# 3. Create an HTTP/HTTPS rule
./cpcli rule add --layer Network --name "web out" \
    --action accept \
    --field 'source=["lan-internal"]' \
    --field 'service=["http","https"]' \
    --field 'track={"type":"Log"}'

# 4. Publish the changes
./cpcli session publish

# 5. Install the policy
./cpcli policy install --package Standard --target CheckPointA
```

**Desktop app — same thing, with a UI:** open `cpcli-desktop`, log in,
navigate to *Object Explorer → Networks* → **Add**, then *Security
Policies → Access Control* → **New rule**, click **Publish** at the top,
then *Install Policy*.

---

## CLI guide

Every command accepts `--help` to see its flags and subcommands.

### Session / authentication

Every operation requires login first. The session is saved to
`~/.config/cpcli/session-<profile>.json` and reused by later commands.
**No password is ever written to disk** — only the session token.

```sh
# Login with user/password (interactive prompt)
./cpcli login --server 192.0.2.10 --user admin

# Login reading the password from an env var (handy in scripts)
CPCLI_PASSWORD='my-secret' ./cpcli login --server 192.0.2.10 --user admin

# Login via API key
./cpcli login --server 192.0.2.10 --api-key 'AAAA-BBBB-CCCC'

# Read-only login
./cpcli login --server 192.0.2.10 --user admin --read-only

# Continue the last session (useful when another admin left an object locked)
./cpcli login --server 192.0.2.10 --user admin --continue-last-session

# Multi-Domain: pick a specific domain
./cpcli login --server 192.0.2.10 --user admin --domain "Corp"

# Skip TLS fingerprint verification (lab only)
./cpcli login --server 192.0.2.10 --user admin --insecure

# Logout
./cpcli logout
```

**Fingerprint security:** on the first connection the CLI asks you to
confirm the server's SHA1 fingerprint (same as `mgmt_cli`). Once accepted,
it is saved to `~/.config/cpcli/fingerprints.json` for later runs.

### Objects (hosts, networks, groups, services)

Every object type shares the same subcommands: `list`, `show`, `add`,
`set`, `delete`.

**Host** (single IP):
```sh
./cpcli host list
./cpcli host add --name web-01 --field ip-address=10.0.0.11
./cpcli host set  --name web-01 --field comments='"Web server"'
./cpcli host show --name web-01
./cpcli host delete --name web-01
```

**Network** (subnet):
```sh
./cpcli network add --name lan-corp \
    --field subnet4=10.0.0.0 --field mask-length4=24
./cpcli network list
```

**Group:**
```sh
./cpcli group add --name grp-servers \
    --field 'members=["web-01","db-01"]'
```

**Services (TCP/UDP/ICMP/other):**
```sh
./cpcli service-tcp add --name svc-8080 --field port=8080
./cpcli service-udp add --name svc-syslog --field port=514
./cpcli service-icmp list
```

**Address range:**
```sh
./cpcli address-range add --name dhcp-pool \
    --field ipv4-address-first=10.0.5.10 \
    --field ipv4-address-last=10.0.5.100
```

**Other supported types:** `service-group`, `access-role`,
`security-zone`, `dns-domain`, `application-site`, `tag`, `time`,
`wildcard`, `dynamic-object`.

The `--field` flag takes JSON — for complex values, use single quotes
around JSON with double quotes:
```sh
--field 'members=["a","b","c"]'
--field 'track={"type":"Log"}'
```

### Access Control

```sh
# List layers
./cpcli rule layers

# List rules in a layer
./cpcli rule list --layer Network

# Add a rule
./cpcli rule add --layer Network \
    --name "web out" \
    --action accept \
    --position top \
    --field 'source=["lan-internal"]' \
    --field 'destination=["Any"]' \
    --field 'service=["http","https"]' \
    --field 'track={"type":"Log"}'

# Edit (by name or uid)
./cpcli rule set --layer Network --name "web out" \
    --field 'service=["http","https","ssh"]'

# Show details
./cpcli rule show --layer Network --name "web out"

# Delete
./cpcli rule delete --layer Network --name "web out"
```

Actions: `accept`, `drop`, `reject`, `ask`.
Positions: `top`, `bottom`, a number (e.g. `3`), `above:<uid>`, `below:<uid>`.

### NAT

```sh
# List
./cpcli nat list --package Standard

# Hide NAT (mask an entire network behind the gateway itself)
./cpcli nat add --package Standard \
    --field 'method="hide"' \
    --field 'original-source="lan-internal"' \
    --field 'translated-source="CheckPointA"' \
    --field 'position="bottom"'

# Static NAT (1:1 mapping)
./cpcli nat add --package Standard \
    --field 'method="static"' \
    --field 'original-source="web-01"' \
    --field 'translated-source="pub-web-01"'

./cpcli nat delete --package Standard --rule-number 1
```

### Threat Prevention

```sh
./cpcli threat layer list
./cpcli threat rule list --layer 'Standard Threat Prevention'
./cpcli threat rule add --layer 'Standard Threat Prevention' \
    --name "servers strict" \
    --field 'protected-scope=["grp-servers"]' \
    --field 'action="Strict"' \
    --field 'position="top"'
./cpcli threat profile list
```

### HTTPS Inspection

```sh
./cpcli https layer list
./cpcli https rule list --layer 'Default Layer'
./cpcli https rule add --layer 'Default Layer' \
    --name "bypass banks" \
    --field 'action="Bypass"' \
    --field 'site-category=["Financial Services"]' \
    --field 'position="top"'
```

### Gateways and interfaces

```sh
# List all gateways/servers
./cpcli gateway list

# Show details (interfaces, blades, VPN, etc.)
./cpcli gateway show --name CheckPointA

# Change a blade (e.g. enable Application Control)
./cpcli gateway set --name CheckPointA --field application-control=true

# --- Interfaces ---
# Mark eth1 as internal with network defined by the interface mask + anti-spoofing
./cpcli gateway interface set --gateway CheckPointA --interface eth1 \
    --field 'topology="internal"' \
    --field 'topology-settings={"ip-address-behind-this-interface":"network defined by the interface ip and net mask"}' \
    --field 'anti-spoofing=true'

# Mark eth0 as external
./cpcli gateway interface set --gateway CheckPointA --interface eth0 \
    --field 'topology="external"'
```

### VPN Communities

```sh
./cpcli vpn meshed list
./cpcli vpn meshed add --field 'name="branches"' \
    --field 'gateways=["gw-sp","gw-rio"]'

./cpcli vpn star list
./cpcli vpn star add --field 'name="hub-spoke"' \
    --field 'center-gateways=["gw-hq"]' \
    --field 'satellite-gateways=["gw-sp","gw-rio"]'
```

### Policy installation

```sh
# List available packages
./cpcli policy package list

# Verify without installing
./cpcli policy verify --package Standard

# Install on one or more gateways
./cpcli policy install --package Standard --target CheckPointA
./cpcli policy install --package Standard --target GW-SP --target GW-RJ
```

### Object search

```sh
# Search all types by substring
./cpcli find --filter "web"

# Restrict to a type
./cpcli find --filter "web" --type host

# Find where an object is referenced (rules, groups, etc.)
./cpcli find where-used --name lan-internal
./cpcli find where-used --name lan-internal --indirect  # includes references via groups
```

### Session (publish, discard)

Check Point works transactionally: `add`/`set`/`delete` stay pending
until you **publish**.

```sh
./cpcli session show      # show number of pending changes
./cpcli session publish   # commit
./cpcli session discard   # cancel everything pending
```

### Raw — any API command

For any Management API command the CLI does not expose directly:

```sh
# Inline JSON body
./cpcli raw show-object --json '{"uid":"xxx","details-level":"full"}'

# From a file
./cpcli raw add-time --file config/time-obj.json

# Paginated listing command
./cpcli raw show-security-zones --list --container-key objects

# Async task without waiting (returns the task-id)
./cpcli raw install-policy --no-wait \
    --json '{"policy-package":"Standard","targets":["gw"]}'
```

### Profiles (multi-server)

To manage multiple Management servers:

```sh
./cpcli --profile prod login --server 10.0.0.1 --user admin
./cpcli --profile lab  login --server 192.0.2.10 --user admin --insecure

./cpcli --profile prod rule list --layer Network
./cpcli --profile lab  rule list --layer Network
```

Each profile owns its `session-<profile>.json`.

### Shell autocompletion

```sh
./cpcli completion bash > /etc/bash_completion.d/cpcli   # or zsh, fish, powershell
```

---

## Desktop app guide

After `wails build -tags webkit2_41`, the binary is
`desktop/build/bin/cpcli-desktop`. It opens a native window (WebKitGTK on
Linux).

**Layout:**
- **Left sidebar:** Dashboard, Gateways & Servers, Security Policies
  (Access Control, NAT, Threat Prevention, HTTPS Inspection, Install
  Policy), Monitoring (Logs & Monitor), VPN Communities, Object Explorer
  (hosts, networks, groups, services, ...).
- **Top bar:** connected server indicator, pending changes counter,
  **Publish** and **Discard** buttons, dark/light theme toggle.

**Typical flow** (same as the CLI): make a change → top bar shows "N
pending changes" → click **Publish** → go to **Install Policy** to apply
on the gateway.

**Highlights:**
- **SmartConsole-style Object Picker** on rule fields (source/destination/
  service): text search, categories, inline object creation without
  leaving the dialog.
- **Get Interfaces** on the gateway's Interfaces tab: mirrors the
  SmartConsole button that syncs topology via SIC.
- **Auto-classify topology:** one click marks every interface as
  `internal` with "network by mask"; you then manually flip the
  Internet-facing one to `external`.
- **Logs & Monitor:** reads `fw log` directly from the gateway via
  `run-script` — works on Standalone installs (no Smart-1 Log Server
  needed). Row-expandable table with full detail (GCP Cloud Logging
  style). Dedicated column for the NAT-translated tuple.
- **Per-interface anti-spoofing** with action (`prevent`/`detect`) and
  tracking (`none`/`log`/`alert`), free from the `set-simple-gateway`
  footguns (which overwrite the entire interfaces array).

---

## Project layout

```
cmd/cpcli/              # CLI entrypoint
internal/
  cli/                  # cobra subcommands (login, rule, nat, host, ...)
  mgmt/                 # core: client, session, fingerprint pinning,
                        # error extraction, rulebase resolver with dictionary
service/                # bindable facade (Login, ListObjects, AddRule, ...)
                        # this is what the desktop consumes via Wails bindings
desktop/
  main.go               # Wails app; binds service.Service
  wails.json            # Wails config
  frontend/             # Vite + React + TS + Tailwind + shadcn/ui
    src/
      pages/            # one page per screen (AccessRulesPage, NatPage, ...)
      components/
        shared/         # ObjectPicker, RulebaseTable, EntityFormDialog, ...
        ui/             # shadcn wrappers (Button, Dialog, Table, ...)
      lib/
        wailsService.ts # CpService interface + real bindings
        mockService.ts  # in-memory mock for dev without a firewall
      config/           # objectKinds, objectCategories, nav
      stores/           # zustand (session)
```

---

## Tests

```sh
# Unit tests for the core and the facade
go test ./...

# With the race detector
go test -race ./...

# Single package
go test -v ./internal/mgmt
go test -v ./service
```

The core (`internal/mgmt`) exposes a `caller` interface so a fake SDK can
be injected — most tests exercise pagination, error extraction and
session handling without a live server.

---

## Contributing

- Issues and PRs are welcome for any part (CLI, service, desktop).
- When adding a new operation: implement in `service/service.go` first
  (for reuse), then expose it in the CLI (`internal/cli/`) and optionally
  in the desktop (`desktop/frontend/`).
- `go fmt ./... && go vet ./...` must be clean.
- If you change a `service.Service` method signature, run
  `wails generate module -tags webkit2_41` inside `desktop/` to regenerate
  the TS bindings.

---

## License

MIT — see [LICENSE](LICENSE).

---

## Disclaimer

This is an **independent, unofficial** project. It is not affiliated with,
sponsored by, endorsed by, or otherwise associated with **Check Point
Software Technologies Ltd.**

*"Check Point"*, *"SmartConsole"* and other product names are trademarks
or registered trademarks of Check Point Software Technologies Ltd. They
are used here descriptively (nominative fair use) to indicate
compatibility with the official, documented **Check Point Management
API** — with no intent to cause confusion about origin or authorship.

**Third-party components:**

- [`github.com/CheckPointSW/cp-mgmt-api-go-sdk`](https://github.com/CheckPointSW/cp-mgmt-api-go-sdk) —
  Check Point's official SDK, licensed under **Apache License 2.0**.
  Consumed as a Go module dependency; not redistributed.
- [Wails](https://wails.io), [React](https://react.dev),
  [Tailwind CSS](https://tailwindcss.com), [shadcn/ui](https://ui.shadcn.com),
  [lucide-react](https://lucide.dev) — all under permissive licenses
  (MIT or similar).

If you are from Check Point and would prefer any part of this project to
be renamed or adjusted, please open an issue on the repository.
