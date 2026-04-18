# OT Calculator Backend (Go + openGauss)

This service stores OT/BREAK entries by AM/PM work session, calculates summaries, persists authoritative results, and optionally stores rendered HTML fragments as cache.

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

## 3) Error format

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
