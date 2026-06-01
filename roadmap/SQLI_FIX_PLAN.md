# SQL Injection Remediation Plan — `crud` package

Status: **PROPOSED — not yet applied.** This document is for review before any code changes.

## 1. Scope & root cause

The generated read API binds attacker-controlled query parameters into `GetAllRequest`
(`crud/structs.go`) and forwards them, unvalidated, into raw SQL built with `fmt.Sprintf`.

GORM only parameterizes `?` placeholders — it never escapes **SQL identifiers**
(column names, relation names, sort direction). Every place where a request-supplied
identifier is interpolated into the query string is injectable.

### Affected sinks

| # | Input (request param) | Code | Injected element | Severity |
|---|------------------------|------|------------------|----------|
| 1 | `fields` | `service.go` `tx.Select(fields)` | SELECT column list (GORM marks unknown names `Raw:true`) | High (data exfiltration via subquery) |
| 2 | `filter` | `utils.go` `filterMapper` | column name in `WHERE` | High |
| 3 | `sort` | `utils.go` `sortMapper` | column **and** direction in `ORDER BY` | High |
| 4 | `s` | `utils.go` `searchMapper` | column name in `WHERE`; raw condition fallthrough | High |

### Secondary bugs (fix alongside)

- **Panic / DoS:** unchecked `value.(string)` type assertions in `searchMapper`
  (`$in` branches) crash on non-string JSON values.
- **Raw condition fallthrough:** `tx.Where(whereField, whereVal)` (search shorthand,
  both `$and`/`$or`) passes an attacker string as the entire SQL condition.
- **Correctness (note only):** `router.go` appends `id||$eq||<id>`, but `$eq` is **not**
  a key in `filterConditions` (`eq` is), so the per-id filter is silently dropped and
  `FindOne` ignores the `:id`. Flag for the maintainer; not part of the SQLi fix.

## 2. Strategy

Single principle: **never interpolate a request-supplied identifier unless it has been
resolved against the model schema; always parameterize values; whitelist direction.**

Building blocks (new, in `crud/validation.go`):

```go
// ensureSchema parses the model schema on the statement if needed.
func ensureSchema(tx *gorm.DB) *schema.Schema

// resolveColumn maps a user-supplied field to its canonical DB column name,
// returning ok=false if it is not a real column on the model.
func resolveColumn(tx *gorm.DB, field string) (string, bool)
```

`resolveColumn` uses `tx.Statement.Schema.LookUpField(field)` and returns `f.DBName`.
Because the value comes from the schema, it is then safe to quote with
`tx.Statement.Quote(col)` before splicing. Unknown columns are **rejected** (skipped),
which is the safe default for a generic filter API.

Relation-qualified names (e.g. `author.name`) are out of scope and rejected for now.

## 3. Concrete changes

### 3.1 `Fields` (in `FindTrx` and `FindOne`, `service.go`)

**Before**

```go
if len(api.Fields) > 0 {
    fields := strings.Split(api.Fields, ",")
    tx.Select(fields)
}
```

**After**

```go
if len(api.Fields) > 0 {
    raw := strings.Split(api.Fields, ",")
    fields := make([]string, 0, len(raw))
    for _, f := range raw {
        if col, ok := resolveColumn(tx, strings.TrimSpace(f)); ok {
            fields = append(fields, col) // canonical name -> GORM quotes it (Raw:false)
        }
    }
    if len(fields) > 0 {
        tx.Select(fields)
    }
}
```

Passing canonical schema column names means GORM takes the `LookUpField` branch
(`callbacks/query.go`) and quotes them, instead of the `Raw:true` passthrough.

### 3.2 `sortMapper`

**Before**

```go
func (q *QueryToDBConverter) sortMapper(sorts []string, tx *gorm.DB) {
    for _, sort := range sorts {
        sortParams := strings.Split(sort, SortSeparator)
        if len(sortParams) == 2 {
            tx.Order(fmt.Sprintf("%s %s", sortParams[0], strings.ToLower(sortParams[1])))
        } else {
            tx.Order(fmt.Sprintf("%s desc", sortParams[0]))
        }
    }
}
```

**After**

```go
func (q *QueryToDBConverter) sortMapper(sorts []string, tx *gorm.DB) {
    for _, sort := range sorts {
        sortParams := strings.Split(sort, SortSeparator)
        col, ok := resolveColumn(tx, sortParams[0])
        if !ok {
            continue
        }
        desc := true // default
        if len(sortParams) == 2 && strings.EqualFold(sortParams[1], "asc") {
            desc = false
        }
        tx.Order(clause.OrderByColumn{
            Column: clause.Column{Name: col},
            Desc:   desc,
        })
    }
}
```

`clause.OrderByColumn` quotes the column and encodes direction as a bool — no string
interpolation remains.

### 3.3 `filterMapper`

Column resolved + quoted; operator already whitelisted via `filterConditions`; value
parameterized (unchanged for values, which were already safe).

**After (shape)**

```go
func (q *QueryToDBConverter) filterMapper(filters []string, tx *gorm.DB) {
    for _, filter := range filters {
        p := strings.Split(filter, SEPARATOR)
        if len(p) < 2 {
            continue
        }
        operator, ok := filterConditions[p[1]]
        if !ok {
            continue
        }
        col, ok := resolveColumn(tx, p[0])
        if !ok {
            continue
        }
        qc := tx.Statement.Quote(col)

        switch p[1] {
        case NotNullOperator, IsNullOperator:
            tx.Where(fmt.Sprintf("%s %s", qc, operator))
        default:
            if len(p) != 3 {
                continue
            }
            switch p[1] {
            case ContainOperator:
                tx.Where(fmt.Sprintf("%s %s ?", qc, operator), "%"+p[2]+"%")
            case InOperator:
                tx.Where(fmt.Sprintf("%s IN ?", qc), strings.Split(p[2], ","))
            default:
                tx.Where(fmt.Sprintf("%s %s ?", qc, operator), p[2])
            }
        }
    }
}
```

### 3.4 `searchMapper`

Refactor the duplicated `$and`/`$or` bodies into one helper that:

1. resolves + quotes the column (`resolveColumn` + `Quote`); rejects unknown columns,
2. uses only whitelisted operators from `filterConditions`,
3. parameterizes all values,
4. type-checks `value.(string)` before `strings.Split` (fixes the panic),
5. replaces the raw `tx.Where(whereField, whereVal)` shorthand with a quoted
   equality: `tx.Where(fmt.Sprintf("%s = ?", qc), whereVal)`.

```go
// applyCond applies one validated condition, as AND (or==false) or OR (or==true).
func (q *QueryToDBConverter) applyCond(tx *gorm.DB, or bool, field, opKey string, value interface{}) {
    col, ok := resolveColumn(tx, field)
    if !ok {
        return
    }
    qc := tx.Statement.Quote(col)
    where := tx.Where
    if or {
        where = tx.Or
    }

    operator, known := filterConditions[opKey]
    if !known { // equality shorthand: {"name": "john"}
        where(fmt.Sprintf("%s = ?", qc), value)
        return
    }

    switch opKey {
    case NotNullOperator, IsNullOperator:
        where(fmt.Sprintf("%s %s", qc, operator))
    case InOperator:
        s, ok := value.(string)
        if !ok {
            return
        }
        where(fmt.Sprintf("%s IN ?", qc), strings.Split(s, ","))
    case ContainOperator:
        where(fmt.Sprintf("%s %s ?", qc, operator), fmt.Sprintf("%%%v%%", value))
    default:
        where(fmt.Sprintf("%s %s ?", qc, operator), value)
    }
}
```

`searchMapper` then iterates `$and`/`$or` arrays and calls `applyCond`, removing the
two nearly-identical blocks.

> Note: GORM's `tx.Where`/`tx.Or` return a `*gorm.DB`; the helper assigns the method
> value as shown for clarity. Final implementation will call them directly to preserve
> the chained statement.

### 3.5 New file `crud/validation.go`

```go
package crud

import (
    "sync"

    "gorm.io/gorm"
    "gorm.io/gorm/schema"
)

var schemaCache = &sync.Map{}

func ensureSchema(tx *gorm.DB) *schema.Schema {
    if tx.Statement.Schema != nil {
        return tx.Statement.Schema
    }
    if tx.Statement.Model != nil {
        if s, err := schema.Parse(tx.Statement.Model, schemaCache, tx.NamingStrategy); err == nil {
            tx.Statement.Schema = s
        }
    }
    return tx.Statement.Schema
}

func resolveColumn(tx *gorm.DB, field string) (string, bool) {
    s := ensureSchema(tx)
    if s == nil {
        return "", false
    }
    if f := s.LookUpField(field); f != nil && f.DBName != "" {
        return f.DBName, true
    }
    return "", false
}
```

New import added where needed: `gorm.io/gorm/clause` (in `utils.go`).

## 4. Behavior changes (call out for review)

- Unknown / non-existent columns in `fields`, `filter`, `sort`, `s` are **silently
  skipped** rather than erroring. Alternative: return `400`. Recommend skip to preserve
  current lenient API behavior; easy to switch to reject.
- `relation.column` style identifiers in filters/sort are rejected (were never truly
  supported, only accidentally injectable). `join`/`Preload` relation handling is
  unchanged by this plan (it already maps via `Preload`, not raw SQL) — but see §7.

## 5. Test plan

The goal is twofold: (a) prove the injection is closed, and (b) prove the **business
behavior of the read APIs is unchanged** — i.e. the security hardening did not break any
legitimate query feature. Both layers must pass before the fix is considered done.

### 5.1 Test setup

Use a **real SQLite in-memory database** (`gorm.io/driver/sqlite`) so we assert actual
returned rows, not just generated SQL. Seed a deterministic dataset once per suite:

- Model under test: `models.Post` (the model wired in `router.go`). If `Post` lacks
  enough columns to exercise every operator, add a small dedicated test model
  (e.g. `testItem{ID, Name, Status, Price, Stock, CreatedAt, AuthorID}` + a `testAuthor`
  relation) registered only in the test file.
- Seed ~8–10 rows with varied values so ordering, ranges, `IN`, `cont`/ILIKE, null vs
  non-null, and pagination all produce distinguishable results.
- Helper `newTestService(t)` returns a `*Service[testItem]` backed by the seeded DB, plus
  a teardown.
- For a few assertions on the exact SQL string, additionally use
  `db.Session(&gorm.Session{DryRun: true})` and inspect `Statement.SQL` / `Statement.Vars`
  (to confirm identifiers are quoted and values are bound as `?`, not inlined).

Run: `go test ./crud/... -run . -count=1`

### 5.2 Functional / business tests (core features still work)

These MUST pass identically before and after the fix.

**A. Pagination & limit (`Find`)**
1. `Limit=3, Page=1` → returns first 3 rows; `totalRows` equals full dataset count.
2. `Limit=3, Page=2` → returns rows 4–6 (correct `OFFSET`).
3. `Page=0` → no offset applied; limit still respected.
4. `totalPages` math in `router.go` matches `ceil(total/limit)`.

**B. Field selection (`fields`)**
5. `fields=id,name` → only those columns populated; others zero-valued; rows still match.
6. `fields=id, name` (spaces) → trimmed and accepted.
7. Single valid field works.
8. Mixed valid + invalid (`fields=name,bogus`) → keeps `name`, drops `bogus`, query succeeds.

**C. Filtering (`filter`, each operator in `filterConditions`)**
9.  `eq`  — `status||eq||active` returns only active rows.
10. `ne`  — excludes the value.
11. `gt`/`gte`/`lt`/`lte` — numeric range boundaries correct (esp. inclusive vs exclusive).
12. `cont` — `name||cont||oo` matches substring (ILIKE/LIKE), case-insensitive.
13. `$in` — `id||$in||1,2,3` returns exactly those rows; single value also works.
14. `isnull` / `notnull` — partition rows correctly.
15. Multiple filters combine with AND (all conditions applied).

**D. Search (`s` JSON)**
16. `$and` with multiple conditions → intersection.
17. `$or` with multiple conditions → union (verify first-vs-rest `Where`/`Or` chaining
    produces the intended grouping, not accidental precedence changes).
18. Equality shorthand `{"name":"john"}` → parameterized `= ?`, correct rows.
19. Operator forms inside `$and`/`$or` (`eq`, `cont`, `$in`, `isnull`) return same rows
    as the equivalent `filter` form (cross-check B/C).

**E. Sorting (`sort`)**
20. `sort=price,asc` → ascending order.
21. `sort=price,desc` → descending order.
22. `sort=price` (no direction) → defaults to `desc` (preserve current behavior).
23. Multiple sort keys → stable multi-column ordering.

**F. FindOne / `:id` route**
24. Valid id returns the matching row.
25. (Regression guard) document current behavior of the `$eq` filter in `router.go`;
    if the §7 fix is included, assert `:id` actually filters to that row and a wrong id
    returns `record not found`.

**G. Joins / relations (`join` → Preload)**
26. `join=Author` → related entity preloaded and populated.
27. Nested `join=Author.Profile` (if model supports) preloads nested relation.

### 5.3 Security regression tests (injection is closed)

Assert these payloads neither error-inject nor alter results, and (DryRun) never appear
verbatim in `Statement.SQL`:

28. `fields=(SELECT password FROM authors LIMIT 1)` → dropped; normal columns returned.
29. `sort=(CASE WHEN (SELECT 1)=1 THEN id ELSE name END)` → dropped; default/again-safe order.
30. `filter=1) OR (1=1||isnull` → dropped; row count unchanged (no boolean bypass).
31. `filter=id;DROP TABLE posts;||eq||1` → dropped; table still intact (query after still works).
32. `s={"$and":[{"1=1)--":{"eq":"x"}}]}` → unknown column dropped; no rows leaked.
33. `s={"$or":[{"id":{"$in":123}}]}` (non-string `$in` value) → **no panic**, condition skipped.
34. Time-based probe `sort=(SELECT 1 FROM pg_sleep(2))` → dropped (no measurable delay).

### 5.4 Pass criteria

- All §5.2 functional tests green **and** their assertions match a baseline captured on
  the pre-fix code for valid inputs (no behavioral drift for legitimate queries).
- All §5.3 security tests green.
- `go vet ./crud/...` and `go test ./crud/... -race` clean.

## 6. End-to-end (HTTP / curl) testing against a real SQLite DB

§5 verifies the service/SQL layer. §6 exercises the **entire stack** — gin routing →
`ShouldBindQuery` → `Service` → GORM → a real SQLite database → JSON response — by sending
**actual HTTP requests with `curl`** and parsing the responses with `jq`. This is the layer
that catches things unit tests can't: query-string binding of repeated `filter=` params
into `[]string`, URL-encoding of the `s` JSON blob, the `page>0` response wrapper vs. raw
array, status codes, and real preload/join JSON shape.

### 6.1 Harness & bootstrap

Reuse the existing `db.OpenTestDB()` (SQLite `file::memory:?cache=shared`) — already in the
codebase and already a dependency (`gorm.io/driver/sqlite`).

- Add a **test entrypoint** that builds the same gin engine as `main.go` (group `crud` +
  `crud.RegisterRoutes`) but backed by the test DB. Either:
  - **(preferred, CI-friendly)** `e2e/server_test.go` that boots the engine via
    `httptest.NewServer(engine)` and exposes its base URL; or
  - an env switch in `main.go` (`APP_ENV=test`) that calls `db.OpenTestDB()` + seeds, so the
    real binary can be launched on `:8081` for `curl`.
- **SQLite gotchas to handle in the harness (important):**
  - Skip `db.AddUUIDExtension()` and the `uuid_generate_v4()` default — that's
    Postgres-only. Seed rows with **explicit** `uuid.UUID` IDs.
  - `AutoMigrate(models.Category{}, models.Post{})` against the SQLite test DB.
  - `cont` maps to `ILIKE` (Postgres). SQLite has no `ILIKE`; document that the e2e suite
    either runs `cont` cases case-insensitively via `LIKE` (SQLite `LIKE` is already
    case-insensitive for ASCII) or asserts substring membership rather than exact operator.
    Flag this as a dialect difference to validate separately on Postgres in CI.

### 6.2 Deterministic seed data

Insert a fixed dataset so every assertion has a known expected result:

- Categories: `Tech` (fixed UUID `C1`), `Food` (fixed UUID `C2`).
- ~8 Posts with known `title` / `price` / `category_id`, including:
  - a spread of prices (e.g. 50, 100, 150, 200, 300, 500) for range/sort tests,
  - at least 2 posts with `category_id = NULL` for `isnull`/`notnull` and join tests,
  - titles containing a common substring (e.g. "go") for `cont`,
  - known IDs so `/crud/:id` and `$in` are deterministic.

### 6.3 Functional curl scenarios (parse + assert with `jq`)

All against `http://localhost:8081/crud`. Use `-G --data-urlencode` for anything with
special characters so the shell/URL encoding is correct.

```bash
BASE=http://localhost:8081/crud

# Pagination (page>0 => wrapped object {data,total,totalPages})
curl -s "$BASE?limit=3&page=1" | jq '.data | length'      # expect 3
curl -s "$BASE?limit=3&page=1" | jq '.total'              # expect 8
curl -s "$BASE?limit=3&page=1" | jq '.totalPages'         # expect 3
curl -s "$BASE?limit=3&page=2" | jq '[.data[].id]'        # expect rows 4-6 (offset works)

# page=0 => raw array (no wrapper)
curl -s "$BASE?limit=100" | jq 'type'                     # expect "array"

# Field selection
curl -s "$BASE?fields=id,title" | jq '.[0] | keys'        # only id,title present

# Filters
curl -s -G "$BASE" --data-urlencode 'filter=price||gte||100' \
                   --data-urlencode 'filter=price||lte||300' | jq '[.[].price] | min, max'
curl -s -G "$BASE" --data-urlencode 'filter=title||cont||go'        | jq 'length'
curl -s -G "$BASE" --data-urlencode 'filter=category_id||isnull'    | jq 'length'   # the 2 null rows
curl -s -G "$BASE" --data-urlencode 'filter=id||$in||<ID1>,<ID2>'   | jq '[.[].id]'

# Sorting
curl -s "$BASE?sort=price,asc"  | jq '[.[].price]'        # ascending
curl -s "$BASE?sort=price,desc" | jq '[.[].price]'        # descending
curl -s "$BASE?sort=price"      | jq '[.[].price]'        # defaults to desc

# Search (s = URL-encoded JSON)
curl -s -G "$BASE" --data-urlencode \
  's={"$or":[{"price":{"gte":"500"}},{"title":{"cont":"sale"}}]}' | jq 'length'
curl -s -G "$BASE" --data-urlencode \
  's={"$and":[{"category_id":{"notnull":""}}]}' | jq 'length'

# Join / Preload
curl -s "$BASE?join=Category" | jq '.[0].category.name'   # related row populated

# FindOne by id
curl -s "$BASE/<ID1>" | jq '.id'                          # expect "<ID1>"
```

Each line gets an `assert_eq <expected> <actual>` in the script (helper below).

### 6.4 Security curl scenarios (injection closed, data intact)

```bash
# Subquery exfiltration via fields -> dropped, normal columns returned, HTTP 200
curl -s -o /dev/null -w '%{http_code}\n' -G "$BASE" \
  --data-urlencode 'fields=id,(SELECT name FROM categories LIMIT 1)'   # expect 200
curl -s -G "$BASE" --data-urlencode 'fields=id,(SELECT name FROM categories LIMIT 1)' \
  | jq '.[0] | has("name") | not'                                       # no leaked column

# ORDER BY injection -> dropped, no 500
curl -s -o /dev/null -w '%{http_code}\n' -G "$BASE" \
  --data-urlencode 'sort=(CASE WHEN 1=1 THEN id ELSE title END)'        # expect 200

# Boolean-bypass via filter -> row count unchanged
curl -s -G "$BASE" --data-urlencode 'filter=1) OR (1=1||isnull' | jq 'length'  # == baseline

# Destructive attempt then verify table still intact
curl -s -G "$BASE" --data-urlencode 'filter=id;DROP TABLE posts;--||eq||1' >/dev/null
curl -s "$BASE?limit=1" | jq 'length'                                   # still > 0

# Non-string $in must not 500 (panic fix)
curl -s -o /dev/null -w '%{http_code}\n' -G "$BASE" \
  --data-urlencode 's={"$or":[{"price":{"$in":123}}]}'                  # expect 200
```

### 6.5 Orchestration

Deliver `scripts/e2e.sh` that:

1. starts the test server (test entrypoint or `httptest`-backed binary) on `:8081`,
2. polls `GET /` until healthy (the `{"message":"ok 5"}` root handler),
3. runs §6.3 + §6.4 cases through `curl | jq`, comparing against expected values with an
   `assert_eq`/`assert_contains` helper,
4. prints a pass/fail summary and **exits non-zero on the first failed assertion**,
5. tears the server down on exit (trap).

Requirements: `curl` and `jq` available in CI. Keep it runnable locally with a single
`bash scripts/e2e.sh`.

### 6.6 Pass criteria

- Every §6.3 request returns the exact parsed value expected from the seed data.
- Every §6.4 request returns HTTP `200` with benign results, leaks no extra columns/rows,
  and the dataset remains intact after destructive payloads.
- `scripts/e2e.sh` exits `0` in CI.

## 7. Out of scope / follow-ups

- `relationsMapper` (`join` param) uses `Preload` with title-cased segments — not raw
  SQL, lower risk — but unvalidated relation names should still be allow-listed against
  `Schema.Relationships.Relations`. Recommend a follow-up PR.
- `router.go` `$eq` vs `eq` correctness bug (see §1).
- Consider centralizing the operator whitelist and exposing a per-model column
  allow-list in the code generator templates so all generated services inherit this.

## 8. Files touched (when applied)

- `crud/service.go` — `Fields` validation in `FindTrx` and `FindOne`.
- `crud/utils.go` — rewrite `filterMapper`, `sortMapper`, `searchMapper`; add `clause` import.
- `crud/validation.go` — **new**, schema-based identifier resolver.
- `crud/crud_api_test.go` — **new**, functional (§5.2) + security regression (§5.3) tests
  against a seeded SQLite in-memory DB.
- `e2e/server_test.go` — **new**, boots the gin engine over `db.OpenTestDB()` for HTTP-level
  tests (§6).
- `scripts/e2e.sh` — **new**, curl + jq end-to-end suite (§6.3–§6.5).
- `go.mod` — `gorm.io/driver/sqlite` is already present; no new dependency required.
