# OT Backend API Contract (Frontend Integration)

Base URL (example):
- `http://localhost:8080`

Content-Type:
- Request: `application/json`
- Response: `application/json`

---

## 1) Session lifecycle

### 1.1 Create or get session
- **Method**: `POST`
- **Path**: `/api/sessions`
- **Description**: Create a session if missing, otherwise return existing one.

#### Request Body
```json
{
  "date": "2026-04-26",
  "period": "AM"
}
```

- `date`: `YYYY-MM-DD` or `MM/DD/YYYY`
- `period`: `AM | PM`

#### Success Response `200`
```json
{
  "sessionId": 2026042601,
  "date": "2026-04-26",
  "period": "AM",
  "status": "OPEN"
}
```

#### Frontend notes
- `sessionId` encoding:
  - `YYYYMMDD01` -> AM
  - `YYYYMMDD02` -> PM

---

## 2) Entries API

### 2.1 Replace all entries in a session
- **Method**: `PUT`
- **Path**: `/api/sessions/{sessionId}/entries`
- **Description**: Hard-replace strategy. Existing entries are deleted; payload is inserted.

#### Path Params
- `sessionId` (`int64`): e.g. `2026042601`

#### Request Body
```json
{
  "entries": [
    {
      "id": "4f7f8c84-9204-4e1e-a0f6-fa63bf5ee6a2",
      "employeeId": "A",
      "entryType": "OT",
      "startTime": "19:15",
      "endTime": "21:45"
    },
    {
      "id": "4ab0f26e-8b9b-40a1-a8ab-72e12dd6fca8",
      "employeeId": "A",
      "entryType": "BREAK",
      "startTime": "20:00",
      "endTime": "20:15"
    }
  ]
}
```

#### Field rules
- `id`: UUID string. If omitted/empty, backend auto-generates UUID.
- `employeeId`: `A | B`
- `entryType`: `OT | BREAK`
- `startTime`, `endTime`: `HH:MM` (24-hour)

#### Success Response `200`
```json
{ "status": "ok" }
```

#### Common errors
- `404 NOT_FOUND` (session not found)
- `409 CONFLICT` (session not OPEN)
- `400 VALIDATION_ERROR` (bad fields)

---

## 3) Calculation API

### 3.1 Calculate and persist session result
- **Method**: `POST`
- **Path**: `/api/sessions/{sessionId}/calculate`
- **Description**:
  1. Load entries by `sessionId`
  2. Run engine
  3. Upsert `session_result`
  4. Upsert optional rendered fragment cache

#### Path Params
- `sessionId` (`int64`)

#### Request Body
- Empty body

#### Success Response `200`
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

---

## 4) Result read APIs

### 4.1 Read persisted result rows
- **Method**: `GET`
- **Path**: `/api/sessions/{sessionId}/result`

#### Success Response `200`
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

### 4.2 Read rendered fragment cache (optional)
- **Method**: `GET`
- **Path**: `/api/sessions/{sessionId}/rendered`

#### Query Params
- `employeeId`: `A | B`
- `fragmentType`: e.g. `DAILY_CARD`
- `formatVersion`: integer, e.g. `1`

#### Example URL
`/api/sessions/2026042601/rendered?employeeId=A&fragmentType=DAILY_CARD&formatVersion=1`

#### Success Response `200`
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

#### Common errors
- `404 NOT_FOUND` (fragment not found)
- `400 VALIDATION_ERROR` (missing/invalid query params)

---

## 5) Backward-compatible calculate API

### 5.1 Stateless calculate (no persistence)
- **Method**: `POST`
- **Path**: `/api/calculate`
- **Description**: Runs calculator directly on payload; does not write DB.

#### Request Body
```json
{
  "otEntries": [
    {
      "id": "1",
      "employeeId": "A",
      "date": "2026-04-26",
      "period": "AM",
      "startTime": "19:00",
      "endTime": "21:00"
    }
  ],
  "breakEntries": [
    {
      "id": "2",
      "employeeId": "A",
      "date": "2026-04-26",
      "period": "AM",
      "startTime": "20:00",
      "endTime": "20:15"
    }
  ]
}
```

#### Success Response `200`
- Same shape as `/api/sessions/{sessionId}/calculate`.

---

## 6) Error contract (all endpoints)

### Error JSON
```json
{
  "code": "VALIDATION_ERROR",
  "message": "period must be AM or PM",
  "details": []
}
```

### Error codes
- `VALIDATION_ERROR` -> HTTP 400
- `NOT_FOUND` -> HTTP 404
- `CONFLICT` -> HTTP 409
- `INTERNAL_ERROR` -> HTTP 500

---

## 7) Frontend implementation checklist

1. Create/get session once user picks date + period.
2. Keep returned `sessionId` as primary key in frontend state.
3. Submit all rows through `PUT /entries` (hard replace).
4. Trigger `POST /calculate` after save.
5. Use `dailySummary` for immediate UI display.
6. Optionally fetch `GET /result` to confirm persisted values.
7. Optionally fetch `GET /rendered` if server-rendered HTML is needed.

---

## 8) cURL examples

```bash
# Create session
curl -X POST http://localhost:8080/api/sessions \
  -H 'Content-Type: application/json' \
  -d '{"date":"2026-04-26","period":"AM"}'

# Replace entries
curl -X PUT http://localhost:8080/api/sessions/2026042601/entries \
  -H 'Content-Type: application/json' \
  -d '{"entries":[{"employeeId":"A","entryType":"OT","startTime":"19:15","endTime":"21:45"}]}'

# Calculate + persist
curl -X POST http://localhost:8080/api/sessions/2026042601/calculate

# Read persisted result
curl http://localhost:8080/api/sessions/2026042601/result

# Read rendered fragment
curl 'http://localhost:8080/api/sessions/2026042601/rendered?employeeId=A&fragmentType=DAILY_CARD&formatVersion=1'
```
