# CheckPoint SmartConsole em Go

Uma alternativa **cross-platform** ao SmartConsole da Check Point, escrita
em Go: um cliente **CLI** (`cpcli`) e um **app desktop** (Wails + React)
para administrar um Check Point Security Management via **Check Point
Management API**.

Feito porque a SmartConsole oficial é Windows-only. Aqui você configura
objetos, regras (Access Control, NAT, Threat Prevention, HTTPS Inspection),
gateways, VPN, instala política e lê logs — tudo em Linux e macOS.

Construído sobre o SDK oficial da Check Point:
[`github.com/CheckPointSW/cp-mgmt-api-go-sdk`](https://github.com/CheckPointSW/cp-mgmt-api-go-sdk).

---

## Índice

- [Componentes](#componentes)
- [Pré-requisitos](#pré-requisitos)
- [Build](#build)
- [Guia rápido](#guia-rápido)
- [Guia do CLI](#guia-do-cli)
- [Guia do app desktop](#guia-do-app-desktop)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Testes](#testes)
- [Contribuindo](#contribuindo)

---

## Componentes

| Módulo | Descrição |
|---|---|
| `cmd/cpcli` | CLI (`cpcli`) — comandos para todo o CRUD e operações de política |
| `service/` | Fachada bindável — `Service` reúne as operações da UI |
| `internal/mgmt/` | Core de transporte/sessão sobre o SDK oficial |
| `desktop/` | App desktop (Wails + React + TypeScript + Tailwind) sobre a fachada |

CLI e app desktop compartilham o mesmo core (`internal/mgmt`) e a mesma
fachada (`service/`). Qualquer feature adicionada ali aparece nos dois
lados.

---

## Pré-requisitos

**Só o CLI:**
- Go 1.25+
- Rede até um Check Point Security Management (versão ≥ R80.10)

**Para o app desktop além disso:**
- Node 20+ e npm (testado com Node 22)
- [Wails CLI](https://wails.io) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Dependências de sistema da WebView (Linux):
  ```sh
  # Fedora
  sudo dnf install gtk3-devel webkit2gtk4.1-devel
  # Debian/Ubuntu
  sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
  ```
- Rode `wails doctor` para conferir o ambiente.

---

## Build

### CLI

```sh
git clone https://github.com/matheusmgon/cpcli.git
cd cpcli
go build -o cpcli ./cmd/cpcli
./cpcli --help
```

### App desktop

```sh
cd desktop
# Distros com webkit2gtk-4.1 (Fedora 40+, Ubuntu 24.04+):
wails build -tags webkit2_41
# Distros ainda em webkit2gtk-4.0:
wails build
```

O binário sai em `desktop/build/bin/cpcli-desktop`.

Para desenvolvimento (hot-reload de verdade):
```sh
cd desktop
wails dev -tags webkit2_41
```

Para iterar só no frontend (com dados mockados, sem conectar em firewall):
```sh
cd desktop/frontend
npm install
npm run dev
```

---

## Guia rápido

**CLI — criando uma rede interna, uma regra HTTPS e instalando política:**

```sh
# 1. Login (senha via prompt ou env CPCLI_PASSWORD)
./cpcli login --server 192.0.2.10 --user admin --insecure

# 2. Criar objeto de rede
./cpcli network add --name lan-interna \
    --field subnet4=10.0.10.0 --field mask-length4=24

# 3. Criar regra HTTP/HTTPS
./cpcli rule add --layer Network --name "web out" \
    --action accept \
    --field 'source=["lan-interna"]' \
    --field 'service=["http","https"]' \
    --field 'track={"type":"Log"}'

# 4. Publicar as mudanças
./cpcli session publish

# 5. Instalar a política
./cpcli policy install --package Standard --target CheckPointA
```

**App desktop — mesma coisa, com UI:** abrir `cpcli-desktop`, logar,
navegar em *Object Explorer → Networks* → **Adicionar**, depois
*Security Policies → Access Control* → **Nova regra**, botão **Publish** no
topo, depois *Instalar política*.

---

## Guia do CLI

Todos os comandos aceitam `--help` para ver flags e subcomandos.

### Sessão / autenticação

Todo o cliente exige login primeiro. A sessão fica salva em
`~/.config/cpcli/session-<profile>.json` e é reutilizada pelos comandos
seguintes. **Nenhuma senha é gravada em disco** — só o token de sessão.

```sh
# Login com usuário/senha (prompt interativo)
./cpcli login --server 192.0.2.10 --user admin

# Login lendo senha da env var (útil em scripts)
CPCLI_PASSWORD='meu-segredo' ./cpcli login --server 192.0.2.10 --user admin

# Login via API key
./cpcli login --server 192.0.2.10 --api-key 'AAAA-BBBB-CCCC'

# Login somente-leitura
./cpcli login --server 192.0.2.10 --user admin --read-only

# Continuar a última sessão (útil quando outro admin deixou objeto locked)
./cpcli login --server 192.0.2.10 --user admin --continue-last-session

# Multi-Domain: escolher domínio específico
./cpcli login --server 192.0.2.10 --user admin --domain "Corp"

# Desabilitar verificação de fingerprint TLS (lab only)
./cpcli login --server 192.0.2.10 --user admin --insecure

# Sair
./cpcli logout
```

**Segurança do fingerprint:** na primeira conexão, o CLI pede confirmação
do fingerprint SHA1 do servidor (igual ao `mgmt_cli`). Aceito, fica
gravado em `~/.config/cpcli/fingerprints.json` para as próximas vezes.

### Objetos (hosts, networks, groups, serviços)

Cada tipo tem os mesmos subcomandos: `list`, `show`, `add`, `set`, `delete`.

**Host** (endereço IP único):
```sh
./cpcli host list
./cpcli host add --name web-01 --field ip-address=10.0.0.11
./cpcli host set  --name web-01 --field comments='"Servidor web"'
./cpcli host show --name web-01
./cpcli host delete --name web-01
```

**Network** (subrede):
```sh
./cpcli network add --name lan-corporativa \
    --field subnet4=10.0.0.0 --field mask-length4=24
./cpcli network list
```

**Group:**
```sh
./cpcli group add --name grp-servidores \
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

**Outros tipos suportados:** `service-group`, `access-role`,
`security-zone`, `dns-domain`, `application-site`, `tag`, `time`,
`wildcard`, `dynamic-object`.

O flag `--field` aceita JSON — para valores complexos, cite entre aspas
simples envolvendo JSON com aspas duplas:
```sh
--field 'members=["a","b","c"]'
--field 'track={"type":"Log"}'
```

### Access Control

```sh
# Listar layers
./cpcli rule layers

# Listar regras de uma layer
./cpcli rule list --layer Network

# Adicionar regra
./cpcli rule add --layer Network \
    --name "web out" \
    --action accept \
    --position top \
    --field 'source=["lan-interna"]' \
    --field 'destination=["Any"]' \
    --field 'service=["http","https"]' \
    --field 'track={"type":"Log"}'

# Editar (por nome ou uid)
./cpcli rule set --layer Network --name "web out" \
    --field 'service=["http","https","ssh"]'

# Ver detalhes
./cpcli rule show --layer Network --name "web out"

# Apagar
./cpcli rule delete --layer Network --name "web out"
```

Ações possíveis: `accept`, `drop`, `reject`, `ask`.
Posições: `top`, `bottom`, número (ex: `3`), `above:<uid>`, `below:<uid>`.

### NAT

```sh
# Lista
./cpcli nat list --package Standard

# Hide NAT (mascara toda uma rede atrás do próprio gateway)
./cpcli nat add --package Standard \
    --field 'method="hide"' \
    --field 'original-source="lan-interna"' \
    --field 'translated-source="CheckPointA"' \
    --field 'position="bottom"'

# Static NAT (mapeamento 1:1)
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
    --field 'protected-scope=["grp-servidores"]' \
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

### Gateways e interfaces

```sh
# Listar todos os gateways/servidores
./cpcli gateway list

# Ver detalhe (interfaces, blades, VPN, etc.)
./cpcli gateway show --name CheckPointA

# Alterar blade (ex: habilitar Application Control)
./cpcli gateway set --name CheckPointA --field application-control=true

# --- Interfaces ---
# Marcar eth1 como internal com rede definida pela máscara + anti-spoofing
./cpcli gateway interface set --gateway CheckPointA --interface eth1 \
    --field 'topology="internal"' \
    --field 'topology-settings={"ip-address-behind-this-interface":"network defined by the interface ip and net mask"}' \
    --field 'anti-spoofing=true'

# Marcar eth0 como external
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

### Instalação de política

```sh
# Lista pacotes disponíveis
./cpcli policy package list

# Verificar sem instalar
./cpcli policy verify --package Standard

# Instalar em um ou mais gateways
./cpcli policy install --package Standard --target CheckPointA
./cpcli policy install --package Standard --target GW-SP --target GW-RJ
```

### Busca de objetos

```sh
# Busca em todos os tipos por substring
./cpcli find --filter "web"

# Restringe a um tipo
./cpcli find --filter "web" --type host

# Descobre onde um objeto é referenciado (rules, groups, etc)
./cpcli find where-used --name lan-interna
./cpcli find where-used --name lan-interna --indirect  # inclui referências via grupos
```

### Sessão (publish, discard)

O Check Point trabalha em modo transacional: `add`/`set`/`delete` ficam
pendentes até você **publicar**.

```sh
./cpcli session show      # mostra número de mudanças pendentes
./cpcli session publish   # efetiva
./cpcli session discard   # cancela tudo pendente
```

### Raw — qualquer comando da API

Para qualquer comando da Management API que o CLI ainda não expõe
diretamente:

```sh
# Comando com corpo JSON inline
./cpcli raw show-object --json '{"uid":"xxx","details-level":"full"}'

# De um arquivo
./cpcli raw add-time --file config/time-obj.json

# Comando de listagem paginada
./cpcli raw show-security-zones --list --container-key objects

# Task assíncrona sem esperar (retorna task-id)
./cpcli raw install-policy --no-wait \
    --json '{"policy-package":"Standard","targets":["gw"]}'
```

### Perfis (multi-servidor)

Para gerenciar vários management servers:

```sh
./cpcli --profile prod  login --server 10.0.0.1 --user admin
./cpcli --profile lab   login --server 192.0.2.10 --user admin --insecure

./cpcli --profile prod rule list --layer Network
./cpcli --profile lab  rule list --layer Network
```

Cada perfil tem seu próprio `session-<profile>.json`.

### Autocomplete de shell

```sh
./cpcli completion bash > /etc/bash_completion.d/cpcli   # ou zsh, fish, powershell
```

---

## Guia do app desktop

Após `wails build -tags webkit2_41`, o binário é
`desktop/build/bin/cpcli-desktop`. Abre uma janela nativa (WebKitGTK no
Linux).

**Layout:**
- **Sidebar esquerda:** Dashboard, Gateways & Servers, Security Policies
  (Access Control, NAT, Threat Prevention, HTTPS Inspection, Instalar
  política), Monitoramento (Logs & Monitor), VPN Communities, Object
  Explorer (hosts, networks, groups, services, ...).
- **Topo:** indicador do server conectado, contador de mudanças pendentes,
  botões **Publish** e **Discard**, toggle de tema escuro/claro.

**Fluxo típico** (mesmo do CLI): fazer mudança → topo mostra "N mudanças
pendentes" → clicar **Publish** → ir em **Instalar política** para efetivar
no gateway.

**Recursos que valem destacar:**
- **Object picker estilo SmartConsole** em campos de regra (origem/destino/
  serviço): busca por texto, categorias, criação inline de objeto sem sair
  do diálogo.
- **Get Interfaces** na aba Interfaces do gateway: reproduz o botão do
  SmartConsole que sincroniza a topologia via SIC.
- **Auto-classify topologia:** um clique marca todas as interfaces como
  `internal` com "rede pela máscara"; depois você marca manualmente a
  Internet-facing como `external`.
- **Logs & Monitor:** lê `fw log` direto do gateway via `run-script` —
  funciona em Standalone (sem Smart-1 log server). Tabela clicável, expande
  linha com detalhe completo abaixo (estilo GCP Cloud Logging). Coluna
  dedicada de NAT traduzido.
- **Anti-spoofing por interface** com ação (`prevent`/`detect`) e tracking
  (`none`/`log`/`alert`) sem os footguns do `set-simple-gateway` original
  (que sobrescreve toda a array de interfaces).

---

## Estrutura do projeto

```
cmd/cpcli/              # entrypoint do CLI
internal/
  cli/                  # subcomandos cobra (login, rule, nat, host, ...)
  mgmt/                 # core: cliente, sessão, fingerprint pinning,
                        # error extraction, rulebase resolver com dicionário
service/                # fachada bindable (Login, ListObjects, AddRule, ...)
                        # é o que o desktop consome via Wails bindings
desktop/
  main.go               # app Wails; binda service.Service
  wails.json            # config do Wails
  frontend/             # Vite + React + TS + Tailwind + shadcn/ui
    src/
      pages/            # uma page por tela (AccessRulesPage, NatPage, ...)
      components/
        shared/         # ObjectPicker, RulebaseTable, EntityFormDialog, ...
        ui/             # wrappers shadcn (Button, Dialog, Table, ...)
      lib/
        wailsService.ts # interface CpService + bindings reais
        mockService.ts  # mock in-memory para dev sem firewall
      config/           # objectKinds, objectCategories, nav
      stores/           # zustand (session)
```

---

## Testes

```sh
# Unit tests do core e da fachada
go test ./...

# Com race detector
go test -race ./...

# Só pacote específico
go test -v ./internal/mgmt
go test -v ./service
```

O core (`internal/mgmt`) tem uma interface `caller` que permite injetar
um SDK fake — muitos testes exercitam paginação, extração de erro e
sessão sem precisar de servidor.

---

## Contribuindo

- Issues e PRs bem-vindos em qualquer parte (CLI, service, desktop).
- Ao adicionar operação nova: implementar em `service/service.go` primeiro
  (para reuso), depois expor no CLI (`internal/cli/`) e opcionalmente no
  desktop (`desktop/frontend/`).
- `go fmt ./... && go vet ./...` deve passar limpo.
- Se alterar assinatura de método em `service.Service`, rode
  `wails generate module -tags webkit2_41` dentro de `desktop/` para
  regenerar os bindings TS.

---

## Licença

MIT — veja [LICENSE](LICENSE).

---

## Aviso legal / Disclaimer

Este projeto é uma iniciativa **independente e não oficial**. Não é afiliado,
patrocinado, endossado ou de qualquer forma associado à **Check Point Software
Technologies Ltd.**

*"Check Point"*, *"SmartConsole"* e demais nomes de produto são marcas
registradas ou marcas comerciais da Check Point Software Technologies Ltd.
São usados aqui apenas em caráter descritivo/nominativo (*nominative fair
use*), para indicar compatibilidade com a **Check Point Management API**
oficial e documentada — sem intenção de causar confusão sobre a origem ou
autoria.

**Componentes de terceiros:**

- [`github.com/CheckPointSW/cp-mgmt-api-go-sdk`](https://github.com/CheckPointSW/cp-mgmt-api-go-sdk) —
  SDK oficial da Check Point, licenciado sob **Apache License 2.0**.
  Consumido como dependência Go, sem redistribuição.
- [Wails](https://wails.io), [React](https://react.dev),
  [Tailwind CSS](https://tailwindcss.com), [shadcn/ui](https://ui.shadcn.com),
  [lucide-react](https://lucide.dev) — todos com licenças permissivas
  (MIT/similar).

Se você é da Check Point e prefere que este projeto seja renomeado ou tenha
qualquer ajuste, abra uma issue no repositório.

