# API Agent — Field Service Management API

## Role
This is the API Agent workspace. It owns everything under `field_service_management_api/`.
Do NOT modify frontend files from here. Do NOT modify FSM schema migrations from here.

---

## Tech Stack
| Layer | Library |
|-------|---------|
| Language | Go 1.22+ |
| HTTP router | github.com/go-chi/chi/v5 |
| CORS | github.com/go-chi/cors |
| JWT | github.com/golang-jwt/jwt/v5 |
| Database driver | github.com/jackc/pgx/v5 (pgxpool) |
| Env loading | github.com/joho/godotenv |
| Password hashing | golang.org/x/crypto/bcrypt |

---

## Project Structure
```
field_service_management_api/
├── cmd/
│   └── server/
│       └── main.go          # Entry point: wires config, DB, router
├── internal/
│   ├── auth/
│   │   └── jwt.go           # GenerateToken / ValidateToken, Claims type
│   ├── config/
│   │   └── config.go        # Load() reads DATABASE_URL, JWT_SECRET, PORT
│   ├── db/
│   │   └── db.go            # Connect() returns *pgxpool.Pool
│   ├── handlers/
│   │   ├── auth.go          # Register, Login http.HandlerFunc factories
│   │   └── helpers.go       # writeJSON, WriteError (exported), writeError
│   ├── middleware/
│   │   └── auth.go          # JWTAuth(secret) middleware, GetClaims(r)
│   └── models/
│       └── user.go          # User, LoginRequest, RegisterRequest, AuthResponse, UserInfo
├── sql/                     # Raw SQL migrations (schema only, run manually)
├── .env.example
├── .gitignore
├── go.mod
└── CLAUDE.md
```

---

## Database Notes
- **All tables live in the `fsm` schema** — including `fsm.users` (auth). Never reference `public.*` tables.
- `fsm.users` columns: `id bigint`, `username text`, `email text`, `password text`, `created_at timestamptz`
- Always qualify table names with `fsm.` prefix (e.g., `fsm.users`, `fsm.technicians`, `fsm.work_orders`)
- Use `pgxpool` (never `database/sql`) for all queries

---

## Environment Variables
| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Full pgx-compatible connection string |
| `JWT_SECRET` | Yes | HMAC-SHA256 signing secret, min 32 chars |
| `PORT` | No | HTTP listen port (default: `8080`) |

Copy `.env.example` to `.env` and fill in real values. Never commit `.env`.

---

## JWT Claims Shape
```go
type Claims struct {
    UserID   int64  `json:"user_id"`
    Email    string `json:"email"`
    Username string `json:"username"`
    jwt.RegisteredClaims        // Includes ExpiresAt (24h), IssuedAt, Subject
}
```
Signing algorithm: **HS256**. Token lifetime: **24 hours**.

---

## HTTP Status Codes Convention
| Situation | Code |
|-----------|------|
| Success (read) | 200 |
| Created | 201 |
| Bad request / validation | 400 |
| Unauthorized / invalid token | 401 |
| Forbidden | 403 |
| Not found | 404 |
| Conflict (duplicate) | 409 |
| Internal error | 500 |

---

## Coding Rules
1. **Read before editing** — always read the file before making changes.
2. **Build must pass** — after every change, run `go build ./...` from the module root. Fix any error before declaring done.
3. **No import cycles** — `middleware` imports `handlers` (for `WriteError`) and `auth`; `handlers` imports `auth`, `config`, `models`; nothing imports `middleware` or `handlers` from `auth`/`config`/`models`.
4. **Exported helpers** — `handlers.WriteError` is exported so `middleware` can call it without a cycle.
5. **bcrypt** — always use `bcrypt.DefaultCost`. Never store plaintext passwords.
6. **Error messages** — login failures must return a generic message to prevent user enumeration.
7. **Context propagation** — pass `context.Background()` for DB queries in handlers (no request context timeout yet).
8. **Formatting** — run `gofmt -w .` or let the editor handle it; PRs must be gofmt-clean.
9. **Debugging tasks** — when asked to add debug logging or inspect values, use the simplest possible approach (e.g. `fmt.Printf("%+v\n", val)`). Do NOT restructure the code, add new imports, or use `io.ReadAll`/body-cloning patterns unless explicitly asked. Remove debug lines when done.
