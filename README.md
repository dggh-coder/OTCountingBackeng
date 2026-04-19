# OT Calculator Backend (Go + openGauss)


## 0) New Ubuntu setup (Podman + tools)

On a fresh Ubuntu host:

```bash
sudo apt update
sudo apt install -y podman curl ca-certificates postgresql-client
```

Verify:
```bash
podman --version
psql --version
```

> If `podman` is missing in your distro repo, enable the official Ubuntu updates repo first, then reinstall.

This service stores OT/BREAK entries by AM/PM work session, calculates summaries, persists authoritative results, and optionally stores rendered HTML fragments as cache.


## 0b) New openEuler setup (Podman + tools)

On a fresh openEuler server:

```bash
sudo dnf -y update
sudo dnf -y install podman git curl ca-certificates postgresql
```

Enable/start Podman service pieces (safe to run even if already enabled):
```bash
sudo systemctl enable --now podman.socket || true
sudo systemctl enable --now podman.service || true
```

Verify install:
```bash
podman --version
psql --version
```

(Optional, rootless quality-of-life)
```bash
# allows lingering user services after logout
sudo loginctl enable-linger "$USER"
```

## 1) Quick start

### Prerequisites
- Go 1.23+
- openGauss (PostgreSQL wire compatible)

### Install dependencies
```bash
go mod tidy
```

### Configure database
You can configure DB in either of these ways:

1. **Single DSN** via `DATABASE_URL` (recommended)
2. **Split openGauss env vars**:
   - `OPENGAUSS_HOST`
   - `OPENGAUSS_PORT`
   - `OPENGAUSS_USER`
   - `OPENGAUSS_PASSWORD`
   - `OPENGAUSS_DBNAME`
   - `OPENGAUSS_SSLMODE`

> If you are running in a container, put your openGauss password in `OPENGAUSS_PASSWORD` (or in `DATABASE_URL`), then start the server.

If your container uses `GS_PASSWORD` (like `opengauss-server:latest`), export that into app env:
```bash
export OPENGAUSS_PASSWORD="$GS_PASSWORD"
```

Example:
```bash
export OPENGAUSS_HOST=127.0.0.1
export OPENGAUSS_PORT=5432
export OPENGAUSS_USER=postgres
export OPENGAUSS_PASSWORD='your_password_here'
export OPENGAUSS_DBNAME=postgres
export OPENGAUSS_SSLMODE=disable
```

Or DSN style:
```bash
export DATABASE_URL='postgres://postgres:your_password_here@127.0.0.1:5432/postgres?sslmode=disable'
```

### Run migration
Apply SQL in:
- `internal/db/migrations/001_init.sql`

Example:
```bash
psql "$DATABASE_URL" -f internal/db/migrations/001_init.sql
```

### Start service
```bash
go run ./cmd/server
```

Default listen address: `:8080` (override with `SERVER_ADDR`).

---

## 2) API list + methods + responses

Base URL examples assume `http://localhost:8080`.

### A. Create or get session
- **Method**: `POST`
- **Path**: `/api/sessions`

Request:
```json
{ "date": "2026-04-26", "period": "AM" }
```

Response:
```json
{
  "sessionId": 2026042601,
  "date": "2026-04-26",
  "period": "AM",
  "status": "OPEN"
}
```

### B. Replace entries for a session
- **Method**: `PUT`
- **Path**: `/api/sessions/{sessionId}/entries`

Request:
```json
{
  "entries": [
    { "id":"1", "employeeId":"A", "entryType":"OT", "startTime":"19:15", "endTime":"21:45" },
    { "id":"2", "employeeId":"A", "entryType":"BREAK", "startTime":"20:00", "endTime":"20:15" }
  ]
}
```

Response:
```json
{ "status": "ok" }
```

### C. Calculate and persist
- **Method**: `POST`
- **Path**: `/api/sessions/{sessionId}/calculate`

Response shape (same contract):
```json
{
  "dailySummary": [
    {
      "employeeId": "A",
      "dateLabel": "2026-04-26",
      "rate20Minutes": 135,
      "rate20RoundedHours": 3,
      "rate15Minutes": 135,
      "rate15RoundedHours": 3,
      "calculatedAtUnixSec": 1770000000
    }
  ],
  "monthlySummary": [
    {
      "employeeId": "A",
      "dateLabel": "2026-04-26",
      "rate20Minutes": 135,
      "rate20RoundedHours": 3,
      "rate15Minutes": 135,
      "rate15RoundedHours": 3,
      "calculatedAtUnixSec": 1770000000
    }
  ]
}
```

### D. Read persisted result rows
- **Method**: `GET`
- **Path**: `/api/sessions/{sessionId}/result`

Response:
```json
{
  "results": [
    {
      "sessionId": 2026042601,
      "employeeId": "A",
      "dateLabel": "2026-04-26",
      "rate20Minutes": 135,
      "rate20RoundedHours": 3,
      "rate15Minutes": 135,
      "rate15RoundedHours": 3,
      "calculatedAt": "2026-04-26T10:00:00Z"
    }
  ]
}
```

### E. Read optional rendered fragment cache
- **Method**: `GET`
- **Path**: `/api/sessions/{sessionId}/rendered?employeeId=A&fragmentType=DAILY_CARD&formatVersion=1`

Response:
```json
{
  "sessionId": 2026042601,
  "employeeId": "A",
  "fragmentType": "DAILY_CARD",
  "formatVersion": 1,
  "contentHtml": "<div class=\"daily-card\"><strong>A</strong><span>135 mins</span></div>",
  "lastCalculated": "2026-04-26T10:00:00Z"
}
```

### F. Backward-compatible calculate only (no persistence)
- **Method**: `POST`
- **Path**: `/api/calculate`

Request:
```json
{
  "otEntries": [
    {"id":"1", "employeeId":"A", "startTime":"19:00", "endTime":"21:00"}
  ],
  "breakEntries": [
    {"id":"2", "employeeId":"A", "startTime":"20:00", "endTime":"20:15"}
  ]
}
```

Response: same `dailySummary` / `monthlySummary` structure as calculate endpoint above.

---


## 4) Run backend in Podman

Yes — this backend runs fine in Podman.

### Build image
```bash
podman build -t ot-backend:latest -f Containerfile .
```

This `Containerfile` is intentionally pinned to fully-qualified base image names and is safe for Podman hosts with no short-name registry aliases.

If you see `short-name ... did not resolve to an alias`, your server blocks unqualified images. This repo already uses fully-qualified base images in `Containerfile`; if pulling manually, use full names like `docker.io/library/golang:1.23-alpine`.

### Option A: run backend on host network (Linux)
This is easiest when your openGauss container already publishes `5432` to host.

```bash
podman run --rm -it \
  --name ot-backend \
  --network host \
  -e SERVER_ADDR=':8080' \
  -e OPENGAUSS_HOST='127.0.0.1' \
  -e OPENGAUSS_PORT='5432' \
  -e OPENGAUSS_USER='postgres' \
  -e GS_PASSWORD='xxxxxx' \
  -e OPENGAUSS_DBNAME='postgres' \
  -e OPENGAUSS_SSLMODE='disable' \
  ot-backend:latest
```

Then call API at `http://127.0.0.1:8080`.

### Option B: run backend + openGauss in same podman network
```bash
podman network create ot-net

# openGauss (name: opengauss)
podman run -d \
  --name opengauss \
  --network ot-net \
  --privileged=true \
  --shm-size=1g \
  -e GS_PASSWORD='xxxxxx' \
  -v /data/opengauss:/var/lib/opengauss \
  -p 5432:5432 \
  docker.io/opengauss/opengauss-server:latest

# backend
podman run --rm -it \
  --name ot-backend \
  --network ot-net \
  -p 8080:8080 \
  -e SERVER_ADDR=':8080' \
  -e OPENGAUSS_HOST='opengauss' \
  -e OPENGAUSS_PORT='5432' \
  -e OPENGAUSS_USER='postgres' \
  -e GS_PASSWORD='xxxxxx' \
  -e OPENGAUSS_DBNAME='postgres' \
  -e OPENGAUSS_SSLMODE='disable' \
  ot-backend:latest
```



### End-to-end on new openEuler (copy/paste)
```bash
# 1) clone project
git clone <YOUR_REPO_URL> ot-backend
cd ot-backend

# 2) build backend image
podman build -t ot-backend:latest -f Containerfile .

# 3) create podman network
podman network create ot-net || true

# 4) run openGauss
podman run -d \
  --name opengauss \
  --network ot-net \
  --privileged=true \
  --shm-size=1g \
  -e GS_PASSWORD='xxxxxx' \
  -v /data/opengauss:/var/lib/opengauss \
  -p 5432:5432 \
  docker.io/opengauss/opengauss-server:latest

# 5) apply migration from host
export DATABASE_URL='postgres://postgres:xxxxxx@127.0.0.1:5432/postgres?sslmode=disable'
psql "$DATABASE_URL" -f internal/db/migrations/001_init.sql

# 6) run backend container
podman run --rm -it \
  --name ot-backend \
  --network ot-net \
  -p 8080:8080 \
  -e SERVER_ADDR=':8080' \
  -e OPENGAUSS_HOST='opengauss' \
  -e OPENGAUSS_PORT='5432' \
  -e OPENGAUSS_USER='postgres' \
  -e GS_PASSWORD='xxxxxx' \
  -e OPENGAUSS_DBNAME='postgres' \
  -e OPENGAUSS_SSLMODE='disable' \
  ot-backend:latest
```

### End-to-end on new Ubuntu (copy/paste)
```bash
# 1) clone project
git clone <YOUR_REPO_URL> ot-backend
cd ot-backend

# 2) build backend image
podman build -t ot-backend:latest -f Containerfile .

# 3) create podman network
podman network create ot-net || true

# 4) run openGauss
podman run -d \
  --name opengauss \
  --network ot-net \
  --privileged=true \
  --shm-size=1g \
  -e GS_PASSWORD='xxxxxx' \
  -v /data/opengauss:/var/lib/opengauss \
  -p 5432:5432 \
  docker.io/opengauss/opengauss-server:latest

# 5) apply migration from host
export DATABASE_URL='postgres://postgres:xxxxxx@127.0.0.1:5432/postgres?sslmode=disable'
psql "$DATABASE_URL" -f internal/db/migrations/001_init.sql

# 6) run backend container
podman run --rm -it \
  --name ot-backend \
  --network ot-net \
  -p 8080:8080 \
  -e SERVER_ADDR=':8080' \
  -e OPENGAUSS_HOST='opengauss' \
  -e OPENGAUSS_PORT='5432' \
  -e OPENGAUSS_USER='postgres' \
  -e GS_PASSWORD='xxxxxx' \
  -e OPENGAUSS_DBNAME='postgres' \
  -e OPENGAUSS_SSLMODE='disable' \
  ot-backend:latest
```

### Migration from host
After DB is up, run:
```bash
export DATABASE_URL='postgres://postgres:xxxxxx@127.0.0.1:5432/postgres?sslmode=disable'
psql "$DATABASE_URL" -f internal/db/migrations/001_init.sql
```

## 5) Error format

```json
{
  "code": "VALIDATION_ERROR",
  "message": "period must be AM or PM",
  "details": []
}
```

HTTP status usage:
- `400` validation
- `404` not found
- `409` conflict
- `500` internal/server/db
