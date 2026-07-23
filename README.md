# cpcli

Cliente de linha de comando, em Go, para o **Check Point Management API**
(a mesma API que o SmartConsole usa por baixo dos panos) — pensado para
Linux e macOS, onde não existe um cliente nativo do SmartConsole.

Construído sobre o SDK oficial da Check Point:
[`github.com/CheckPointSW/cp-mgmt-api-go-sdk`](https://github.com/CheckPointSW/cp-mgmt-api-go-sdk).

## Build

```sh
go build -o cpcli ./cmd/cpcli
```

## Uso básico

```sh
# 1. Login (a senha é lida via prompt interativo, ou da env var CPCLI_PASSWORD)
./cpcli login --server 192.168.56.10 --user admin

# Na primeira conexão você vai ver o fingerprint TLS do servidor e precisa
# confirmar (digitando "y") — assim como o mgmt_cli oficial. Ele fica salvo
# em ~/.config/cpcli/fingerprints.json para as próximas vezes.

# 2. Criar um objeto host
./cpcli host add --name web-01 --field ip-address=10.0.0.10

# 3. Publicar a mudança (obrigatório para ela valer)
./cpcli session publish

# 4. Instalar a política nos gateways
./cpcli policy install --package Standard --target gw-lab

# 5. Sair
./cpcli logout
```

A sessão (`sid`) fica salva em `~/.config/cpcli/session-default.json` entre
execuções — cada comando é um processo novo, então não é preciso logar de
novo a cada chamada. Nenhuma senha é gravada em disco.

## Múltiplos servidores

Use `--profile <nome>` em qualquer comando para manter sessões separadas:

```sh
./cpcli --profile lab login --server 192.168.56.10 --user admin
./cpcli --profile prod login --server 10.1.1.1 --user admin
./cpcli --profile lab host list
```

## Comandos disponíveis

| Comando | Cobre |
|---|---|
| `login` / `logout` | Autenticação e sessão |
| `session publish/discard/show` | Publicar ou descartar mudanças pendentes |
| `host`, `network`, `group`, `service-tcp`, `service-udp` | `add`/`show`/`set`/`delete`/`list` de objetos |
| `rule` | Regras de Access Control (`add`/`show`/`set`/`delete`/`list`), `rule layers` |
| `nat` | Regras de NAT (`add`/`show`/`set`/`delete`/`list`) |
| `policy install` / `policy verify` | Instalar/verificar um pacote de política nos gateways |
| `vpn meshed` / `vpn star` | Comunidades VPN site-to-site |
| `task <task-id>` | Status de uma task assíncrona |
| `raw <command>` | Executa **qualquer** comando da Management API não coberto acima |

Rode `./cpcli <comando> --help` para ver as flags de cada um.

### Campos livres com `--field`

A maioria dos comandos de objeto/regra aceita `--field chave=valor` (repetível)
para qualquer campo da API que não tenha uma flag dedicada. Valores em JSON
válido (`true`, `123`, `["a","b"]`, `{"x":1}`) são interpretados como tal;
o resto vira string:

```sh
./cpcli rule add --layer Network --name "allow-web" \
  --action accept \
  --field source='["any"]' \
  --field destination='["web-01"]' \
  --field service='["http","https"]'
```

### Escape hatch: `raw`

Para qualquer comando da API ainda não modelado (ex: threat prevention,
identity awareness, gateways-and-servers, etc.):

```sh
./cpcli raw show-gateways-and-servers --list --details-level standard
./cpcli raw add-simple-gateway --json '{"name":"gw2","ip-address":"10.0.0.2"}'
```

## Segurança

- A senha nunca é gravada em disco — só o token de sessão (`sid`) retornado
  pelo `login`.
- O fingerprint TLS do servidor é verificado por padrão (trust-on-first-use,
  igual ao SSH): na primeira conexão você confirma o fingerprint mostrado,
  e ele fica fixado para o mesmo servidor daí em diante. Use `--insecure`
  no `login` para pular essa checagem (não recomendado fora de laboratório).
- Toda alteração (`add`/`set`/`delete`) fica pendente até rodar
  `cpcli session publish`, e só afeta os gateways depois de
  `cpcli policy install`.

## Autoconclusão de shell

```sh
./cpcli completion bash > /etc/bash_completion.d/cpcli   # ou zsh/fish
```
