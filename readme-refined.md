# llm_audit — LLM Request Audit Log Service

## Context

In any serious AI-powered product, you can't just fire-and-forget calls to an LLM API.
You need observability: *which model was called, by which feature, with what prompt, how long did it take, did it fail?*

This is a real pattern. Tools like Langfuse, Helicone, and LangSmith exist solely to solve it.
Your task is to build the backend core of such a tool — a small HTTP microservice that **ingests, stores, and exposes LLM call records** from multiple application services.

---

## What You're Building

A Go HTTP microservice called `llm_audit`.

Different product features (e.g. a chatbot, a search ranker, a code assistant) each write their logs
to their own DB table (e.g. `logs_chatbot`, `logs_search`, `logs_code`).
Your service **discovers those tables at startup** and serves all of them through a unified API.

This means: **no hardcoded table names, no hardcoded column lists** — same as in production,
where schemas evolve and teams add new log tables without touching this service.

---

## Database Setup

```sql
-- Example: two feature teams, two tables.
-- Both follow the same column convention but may differ in nullable fields.

CREATE TABLE logs_chatbot (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    request_id  VARCHAR(36)  NOT NULL,
    model       VARCHAR(100) NOT NULL,
    prompt      TEXT,
    response    TEXT,
    status      VARCHAR(20)  DEFAULT 'success',
    tokens_used INT,
    latency_ms  INT,
    error_msg   VARCHAR(500),
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE logs_search (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    request_id  VARCHAR(36)  NOT NULL,
    model       VARCHAR(100) NOT NULL,
    query_text  TEXT,
    result_text TEXT,
    status      VARCHAR(20)  DEFAULT 'success',
    tokens_used INT,
    latency_ms  INT,
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);
```

At startup, discover all tables whose name starts with `logs_` using:
```sql
SHOW TABLES;
SHOW FULL COLUMNS FROM `$table_name`;
```

Tables that don't match the `logs_` prefix are ignored.

---

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | List all discovered log tables |
| `GET` | `/$table` | Query records with optional filters |
| `GET` | `/$table/$id` | Get a single record by ID, or 404 |
| `POST` | `/$table` | Insert a new log record |
| `DELETE` | `/$table/$id` | Delete a record |
| `GET` | `/$table/stats` | Aggregated stats for this table |

---

## Endpoint Details

### `GET /`
Returns the list of discovered log tables.

```
{"tables": ["logs_chatbot", "logs_search"]}
```

---

### `GET /$table?model=gpt-4&status=error&limit=10&offset=0`

Returns a page of records.

- `limit` defaults to 5, `offset` defaults to 0.
- `model` and `status` are **optional filter parameters**.
  Build the `WHERE` clause dynamically based on which query params are present.
  Ignore unknown filter params.
- Only filter on columns that actually exist in the table (use the schema you loaded at startup).

```json
{"records": [{...}, {...}]}
```

---

### `GET /$table/$id`

Returns one record or:
```json
{"error": "record not found"}
```
with HTTP 404.

---

### `POST /$table`

Body is a JSON object. Unknown fields are ignored. The `id` and `created_at` fields are always ignored
(auto-set by the DB).

Respond with:
```json
{"inserted": 42}
```
where `42` is the new record's ID.

Validation rules (same level as the original task):
- Type must match the column type: string → VARCHAR/TEXT, number → INT, null is allowed if the column is nullable.
- If a required (NOT NULL, no DEFAULT) field is missing → return HTTP 422 with `{"error": "field $name is required"}`.

---

### `DELETE /$table/$id`

```json
{"deleted": 1}
```

---

### `GET /$table/stats`

Runs a single aggregated query on the table and returns:

```json
{
  "stats": [
    {"model": "gpt-4",        "count": 120, "avg_latency_ms": 840, "total_tokens": 95000},
    {"model": "gpt-3.5-turbo","count": 430, "avg_latency_ms": 310, "total_tokens": 210000}
  ]
}
```

This is a `SELECT model, COUNT(*), AVG(latency_ms), SUM(tokens_used) FROM ... GROUP BY model` query.
If the table has no `model` column, return `{"error": "stats not supported for this table"}`.

---

## Rules (same spirit as the original)

- **Manual routing only.** No external router libraries. Parse `r.URL.Path` yourself.
- **`database/sql` only.** No ORM, no query builders.
- **No global variables.** Store everything (discovered schema, DB connection) in a struct that lives in a closure.
- **No hardcoded table or column names.** Schema is loaded dynamically at startup.
- **SQL injection prevention.** Table and column names come from your in-memory schema map (safe).
  Values always go through parameterized queries (`?` placeholders).
- On unexpected errors → HTTP 500.
- Requests to unknown tables → HTTP 404 with `{"error": "unknown table"}`.
- All field names in responses match the DB column names exactly.

---

## Key Skills This Trains

| Skill | Where it appears |
|-------|-----------------|
| Manual HTTP routing | Path parsing, method dispatch |
| `database/sql`, no ORM | Every query |
| Dynamic schema introspection | Startup: `SHOW TABLES` + `SHOW FULL COLUMNS` |
| Dynamic query construction | `WHERE` clause built from query params; `INSERT` columns built from body |
| `interface{}` / type assertions | Scanning unknown column types from rows |
| NULL handling | Nullable columns, `sql.NullString`, etc. |
| JSON encode/decode | Request body parsing, response building |
| SQL injection awareness | Whitelisting column names from schema map |
| Aggregation queries | `stats` endpoint |

---

## Local Setup

```bash
docker run -p 3306:3306 \
  -v $(PWD):/docker-entrypoint-initdb.d \
  -e MYSQL_ROOT_PASSWORD=1234 \
  -e MYSQL_DATABASE=golang \
  -d mysql
```

Put your `CREATE TABLE` statements in a `.sql` file in the same directory so Docker runs them on first start.

---

## Hints

- `SHOW FULL COLUMNS FROM` gives you column name, type, nullable flag, and whether there's a default.
  Use that to build your validation logic — no hardcoded type lists.
- `rows.ColumnTypes()` is still useful when scanning query results dynamically.
- For the dynamic `WHERE` clause: collect valid filter keys (those present in the table schema AND in the request query params), then build `WHERE col1=? AND col2=?` with a `[]interface{}` args slice.
- For `INSERT`: same pattern — collect columns from request body that exist in the schema, excluding `id` and auto-timestamp fields.
- `json.Number` (or `UseNumber()` on the decoder) prevents float64 surprises when decoding integers.
- The `stats` endpoint path (`/$table/stats`) must be disambiguated from `/$table/$id` — handle it before the ID route.
