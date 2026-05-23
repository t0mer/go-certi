# go-certi

**go-certi** is a Go rewrite and significant feature expansion of the original Python [t0mer/certi](https://github.com/t0mer/certi) project — an **SSL Certificate Transparency log monitor** that tracks certificates issued for domains you care about and alerts you when something new appears.

---

## Overview

go-certi watches [Certificate Transparency](https://certificate.transparency.dev/) logs for your domains. Every time a new certificate is issued — by Let's Encrypt, a commercial CA, or anyone else — go-certi discovers it, stores it, and notifies you through the channel of your choice. Runs as a single statically-linked binary with an embedded React web UI.

---

## Features

- **CT Log Monitoring** — Fetches certificates from [sslmate Cert Spotter](https://sslmate.com/ct_search_api/) (primary) with [crt.sh](https://crt.sh) as fallback. Supports anonymous access; optional API key unlocks higher rate limits.
- **Scheduler** — Per-domain or global cron-based scan schedule (e.g. `@every 2h`, `@daily`, standard cron expressions).
- **Notifications** — Pluggable notification channels:
  - 🔔 **Shoutrrr** — Telegram, Slack, Discord, email, and [30+ more services](https://containrrr.dev/shoutrrr/services/)
  - 💬 **GreenAPI** — WhatsApp via [green-api.com](https://green-api.com)
  - 💬 **WaWeb** — WhatsApp via [go-whatsapp-web-multidevice](https://github.com/aldinokemal/go-whatsapp-web-multidevice)
- **Certificate Details** — Stores CN, SANs, issuer (friendly name + full DN), issue date, expiry date, revocation status, and source.
- **Web UI** — Mobile-first React + Tailwind UI with light/dark/system theme. All data accessible without any tooling.
- **REST API** — Full OpenAPI 3 / Swagger spec at `/swagger/index.html`.
- **Auth** — Optional UI login (JWT, bcrypt) and optional API token protection. Both off by default for homelab use.
- **Single binary** — Embedded SQLite database (no CGO — pure Go), embedded frontend. No external dependencies at runtime.
- **Multi-arch** — Cross-compiled for `linux/amd64`, `linux/arm64`, `linux/armv7`, `linux/armhf`, `linux/arm`, `windows/amd64`, `windows/arm64`.
- **Docker** — Multi-stage distroless image for `linux/amd64`, `linux/arm64`, `linux/arm/v7`.

---

## Screenshots

### Dashboard
Overview of monitored FQDNs, total certificates discovered, notification channels, and schedules. Recent certificates are listed with issuer and expiry.

![Dashboard](docs/screenshots/dashboard.png)

---

### FQDNs
Add domains to monitor. Enable/disable monitoring per domain, toggle notifications, trigger an immediate scan, or remove a domain.

![FQDNs](docs/screenshots/fqdns.png)

---

### Certificates
Paginated list of all discovered certificates. Shows subject CN, issuer (friendly name + full DN), issue date → expiry date with relative countdown, SANs, source, and a **Revoked** badge when applicable. Filter by FQDN or search by CN/SAN/CA.

![Certificates](docs/screenshots/certificates.png)

---

### Notification Channels
Configure reusable notification targets. Supports Shoutrrr (Telegram, Slack, …), GreenAPI (WhatsApp), and WaWeb (WhatsApp). Each channel has a **Test** button to verify delivery.

![Channels](docs/screenshots/channels.png)

---

### Schedules
Define cron-based scan schedules. Supports robfig/cron syntax: `@every 1h`, `@daily`, `0 */4 * * *`, etc. Mark one as the default — FQDNs without a custom schedule inherit it.

![Schedules](docs/screenshots/schedules.png)

---

### Settings
- **Theme** — Light / Dark / System
- **sslmate API key** — leave blank to use crt.sh (no key required)
- **UI Authentication** — Require login with username + password
- **API Token Protection** — Require `Authorization: Bearer <token>` on all API requests; rotate token at any time

![Settings](docs/screenshots/settings.png)

---

### Login
Shown when UI authentication is enabled. Issues a signed JWT in an HttpOnly cookie on success.

![Login](docs/screenshots/login.png)

---

### Swagger / API Docs
Interactive OpenAPI 3 documentation for all endpoints, available at `/swagger/index.html`.

![Swagger](docs/screenshots/swagger.png)

---

## Quick Start

### Binary

Download the latest release for your platform from the [GitHub Releases](https://github.com/t0mer/go-certi/releases) page.

```bash
# Example: Linux amd64
curl -L https://github.com/t0mer/go-certi/releases/latest/download/go-certi_linux_amd64 -o go-certi
chmod +x go-certi
./go-certi
```

Open **http://localhost:8111** in your browser.

### Docker

```bash
docker run -d \
  --name go-certi \
  -p 8111:8111 \
  -v go-certi-data:/data \
  techblog/go-certi:latest
```

### Docker Compose

```yaml
services:
  go-certi:
    image: techblog/go-certi:latest
    ports:
      - "8111:8111"
    volumes:
      - go-certi-data:/data
    environment:
      GO_CERTI_SSLMATE_API_KEY: "your-key-here"  # optional
    restart: unless-stopped

volumes:
  go-certi-data:
```

---

## Configuration

All settings can be provided as CLI flags **or** environment variables. **Environment variables always win over flags** — ideal for container deployments.

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--port` | `GO_CERTI_PORT` | `8111` | HTTP server port |
| `--conf` | `GO_CERTI_CONF` | `$XDG_CONFIG_HOME/go-certi` | Config + database directory |
| `--sslmate-api-key` | `GO_CERTI_SSLMATE_API_KEY` | _(none)_ | sslmate Cert Spotter API key |
| `--reset-password` | `GO_CERTI_RESET_PASSWORD` | — | Generate a new password, print it, exit |
| `--reset-api-token` | `GO_CERTI_RESET_API_TOKEN` | — | Generate a new API token, print it, exit |
| `--version` | — | — | Print version and exit |

Config and the SQLite database are stored in the `--conf` directory:

```
<conf>/
├── config.json      # Runtime config
└── go-certi.db      # SQLite database (WAL mode)
```

### Resetting credentials offline

```bash
# Reset the login password
./go-certi --conf /data --reset-password

# Reset the API token
./go-certi --conf /data --reset-api-token
```

---

## Notification Channel Config

### Shoutrrr (Telegram, Slack, Discord, email, and more)

```json
{ "url": "telegram://token@telegram?chats=123456789" }
```

See the [Shoutrrr service docs](https://containrrr.dev/shoutrrr/services/) for all supported services and URL formats.

### GreenAPI (WhatsApp)

```json
{
  "instance_id": "your-instance-id",
  "api_token_instance": "your-api-token",
  "chat_id": "972501234567@c.us",
  "api_url": "https://api.green-api.com"
}
```

### WaWeb (WhatsApp via go-whatsapp-web-multidevice)

```json
{
  "base_url": "http://your-waweb-host:3000",
  "phone": "+972501234567",
  "auth": "basic dXNlcjpwYXNz"
}
```

---

## API Reference

Full interactive documentation is available at `/swagger/index.html`.

**Base path:** `/api/v1`

| Method | Endpoint | Description |
|---|---|---|
| `GET` / `POST` | `/fqdns` | List / create FQDNs |
| `GET` / `PUT` / `DELETE` | `/fqdns/:id` | Get / update / delete FQDN |
| `POST` | `/fqdns/:id/scan` | Trigger immediate CT scan |
| `GET` | `/certificates` | List certificates (`?fqdn=`, `?page=`, `?page_size=`) |
| `GET` | `/certificates/:id` | Get certificate |
| `GET` | `/certificates/cas` | List distinct certificate authorities |
| `GET` / `POST` | `/channels` | List / create notification channels |
| `GET` / `PUT` / `DELETE` | `/channels/:id` | Get / update / delete channel |
| `POST` | `/channels/:id/test` | Send a test notification |
| `GET` / `POST` | `/schedules` | List / create schedules |
| `GET` / `PUT` / `DELETE` | `/schedules/:id` | Get / update / delete schedule |
| `GET` / `PUT` | `/settings` | Get / update application settings |
| `POST` | `/settings/api-token/rotate` | Rotate the API token |
| `POST` | `/auth/login` | Login — returns JWT cookie + token |
| `POST` | `/auth/logout` | Clear session cookie |
| `GET` | `/auth/me` | Currently authenticated username |
| `GET` | `/healthz` | Liveness probe (no auth, always 200) |
| `GET` | `/readyz` | Readiness probe (DB ping) |

### Authentication

When **API Token Protection** is enabled, pass the token as a Bearer header:

```
Authorization: Bearer <token>
```

When **UI Authentication** is enabled, the JWT from `/auth/login` can also be used as a Bearer token for programmatic access.

---

## Building from Source

**Prerequisites:** Go 1.25+, Node.js 20+

```bash
git clone https://github.com/t0mer/go-certi
cd go-certi

# Build frontend
cd web && npm install && npm run build && cd ..

# Build binary (embeds the frontend)
go build -o go-certi ./cmd/go-certi

# Run
./go-certi --conf ./data
```

### Cross-compile all platforms

```bash
VERSION=1.0.0 BUILD_MODE=prod bash scripts/build.sh
# Binaries written to dist/
```

### Run tests

```bash
go test ./... -race
```

### Regenerate Swagger docs

```bash
go tool swag init -g cmd/go-certi/main.go -d .,internal/api -o docs/
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25+ |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO) |
| DB queries | [sqlc](https://github.com/sqlc-dev/sqlc) (type-safe generated code) |
| Migrations | Hand-rolled embedded runner (`go:embed`) |
| Scheduler | [robfig/cron/v3](https://github.com/robfig/cron) |
| Notifications | [containrrr/shoutrrr](https://github.com/containrrr/shoutrrr) + GreenAPI HTTP + WaWeb HTTP |
| CT sources | [sslmate Cert Spotter API](https://sslmate.com/ct_search_api/) + [crt.sh](https://crt.sh) fallback |
| Auth | [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) + [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) |
| API docs | [swaggo/swag](https://github.com/swaggo/swag) (OpenAPI 3 / Swagger UI) |
| Frontend | React 18 + Vite + TypeScript + Tailwind CSS + shadcn/ui |
| Logging | `log/slog` (structured, stdlib) |

---

## Credits

Inspired by the original [t0mer/certi](https://github.com/t0mer/certi) Python project by Tomer Klein.

---

## License

See [LICENSE](LICENSE).
