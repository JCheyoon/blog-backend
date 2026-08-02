# blog-backend

Go REST API for [blog-frontend](#) — a personal tech blog / portfolio site.
Admin-only, CRUD built deliberately minimal: no framework, no ORM, no multi-user auth system, because none of those are
needed for a single-author blog.

## Stack

| Concern    | Choice                             | Why                                                                                |
| ---------- | ---------------------------------- | ---------------------------------------------------------------------------------- |
| Routing    | stdlib `net/http` (Go 1.22+)       | Method + path-pattern routing shipped in stdlib — no need for gin/echo/chi         |
| Database   | PostgreSQL + `pgx`                 | Raw SQL, no ORM — explicit queries, no hidden N+1s                                 |
| Auth       | JWT + bcrypt, single admin         | One author, one login. A `users` table would be over-engineering here              |
| Migrations | plain SQL files + `golang-migrate` | Versioned schema without a heavyweight migration framework                         |
| Deployment | Docker, single binary              | `CGO_ENABLED=0` static build → tiny Alpine image                                   |
| Hosting    | Render (free)                      | Only platform with a real, no-expiry-adjacent free web service tier as of 2026     |
| DB hosting | Supabase Postgres (free)           | No 30-day expiry, unlike Render's free Postgres                                    |
| Chat       | Anthropic API, plain `net/http`    | No SDK dependency - a hand-written client is the entire integration surface needed |

## AI features

**"Ask this post" (implemented)** — `POST /api/posts/{slug}/ask`. Answers a
reader's question using _only_ that post's content as context, via the
Anthropic API. Rate-limited to 20 requests/IP/hour (`internal/chat/ratelimit.go`)
since every call costs real money. This is deliberately single-turn and
scoped to one post — no vector search needed yet, the whole post fits in
the prompt.

**"Ask my blog" (planned)** — a blog-wide version of the above that can
answer questions spanning multiple posts. Needs semantic search across all
posts, which the single-post version doesn't: `pgvector` extension on the
same Supabase Postgres (no separate vector DB), Voyage AI for embeddings
(Anthropic's recommended embeddings partner - Claude itself doesn't produce
embeddings), chunk-and-embed on publish, cosine-similarity search for
top-k chunks, then the same Claude call pattern with retrieved chunks as
context and post citations in the answer.

## Project layout

```
cmd/
  api/        entrypoint: wires config, db, handlers, starts the server
  hashpw/     CLI to generate a bcrypt hash for your admin password
internal/
  post/       model, repository (SQL), service (business rules), handler (HTTP)
  auth/       JWT issuing/parsing, login handler, auth middleware
  platform/   config loading, db connection pool
migrations/   versioned SQL schema
```

Each domain (`post`, `auth`) is self-contained: handler → service → repository.
`main.go` only assembles them. This matters for what comes next — a planned
`cmd/mcp-server` binary will import `post.Service` directly and expose it as
MCP tools, reusing the exact same business logic instead of duplicating it.

## Local development

```bash
cp .env.example .env
go run ./cmd/hashpw "your-admin-password"   # copy output into ADMIN_PASSWORD_HASH in .env
# get ANTHROPIC_API_KEY from console.anthropic.com

make docker-up      # starts postgres + api (or point DATABASE_URL at Supabase directly)
make migrate-up      # applies schema (run once, needs golang-migrate CLI)
make run
```

API is now live at `http://localhost:8080`.

## API

```
GET    /healthz
POST   /api/auth/login
GET    /api/posts               ?tag=
GET    /api/posts/{slug}
POST   /api/posts                (auth required)
PUT    /api/posts/{id}           (auth required)
DELETE /api/posts/{id}           (auth required)
POST   /api/posts/{slug}/ask     { "question": "..." } -> { "answer": "..." }, rate-limited
```

## Deliberate non-decisions

- **No ORM / query builder.** SQL is explicit in `repository.go` — easier to
  reason about and to optimize later (indexes are already added for the
  published+created_at listing query and the tags GIN index).
- **No multi-user auth.** Admin credentials live in env vars, checked with
  bcrypt. Adding a `users` table for one person would be premature.
- **No framework.** Go 1.22's `http.ServeMux` handles method + path params
  natively now, which covers everything this API needs.

## Roadmap

- [ ] "Ask this post" chatbot, scoped per post
- [ ] "Ask my blog" — pgvector + Voyage embeddings, blog-wide semantic Q&A
- [ ] Image upload endpoint (presigned URL to object storage)
- [ ] `cmd/mcp-server` — expose `post.Service` as MCP tools (search, summarize,
      draft LinkedIn post) for AI clients like Claude Desktop
- [ ] Auto-translation pipeline on publish
