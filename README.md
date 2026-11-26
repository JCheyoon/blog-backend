# blog-backend

Go REST API for [blog-frontend](#) — a personal tech blog / portfolio site.
Read-only, built deliberately minimal: no framework, no ORM, no auth. Posts are
written in the site repository, so this API only serves published posts and the
"Ask this post" endpoint.

## Stack

| Concern    | Choice                             | Why                                                                                |
| ---------- | ---------------------------------- | ---------------------------------------------------------------------------------- |
| Routing    | stdlib `net/http` (Go 1.22+)       | Method + path-pattern routing shipped in stdlib — no need for gin/echo/chi         |
| Database   | PostgreSQL + `pgx`                 | Raw SQL, no ORM — explicit queries, no hidden N+1s                                 |
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

**"Ask my blog" (on hold)** — a blog-wide version of the above that can
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
internal/
  post/       model, repository (SQL), service (business rules), handler (HTTP)
  category/   same layout, category tree
  chat/       Anthropic client, per-IP rate limiter for /ask
  platform/   config loading, db connection pool
migrations/   versioned SQL schema
```

Each domain (`post`, `category`) is self-contained: handler → service → repository.
`main.go` only assembles them. Business logic stays in the service layer, so
another entrypoint could reuse it without duplicating it. A `cmd/mcp-server`
that imports `post.Service` was planned this way. It is on hold.

## Local development

```bash
# create .env with: DATABASE_URL, ANTHROPIC_API_KEY (PORT is optional, default 8080)
# see internal/platform/config.go for what each one is used for
# get ANTHROPIC_API_KEY from console.anthropic.com

make docker-up      # builds and starts the api container only; DATABASE_URL must point at an existing Postgres (e.g. Supabase)
make migrate-up      # applies schema (run once, needs golang-migrate CLI)
make run
```

API is now live at `http://localhost:8080`.

## API

```
GET    /healthz
GET    /api/posts               ?tag= ?category=
GET    /api/posts/{slug}
GET    /api/categories
POST   /api/posts/{slug}/ask     { "question": "..." } -> { "answer": "..." }, rate-limited
```

## Deliberate non-decisions

- **No ORM / query builder.** SQL is explicit in `repository.go` — easier to
  reason about and to optimize later (indexes are already added for the
  published+created_at listing query and the tags GIN index).
- **No auth.** There is nothing to write through this API: posts are
  written in the site repository. The admin login and write routes were
  removed on 2026-11-25 along with the retired admin UI.
- **No framework.** Go 1.22's `http.ServeMux` handles method + path params
  natively now, which covers everything this API needs.

## Roadmap

As of 2026-11-23, blog posts are written as Markdown in the site repository
([JCheyoon.github.io](https://github.com/JCheyoon/JCheyoon.github.io)) and are
published through a reviewed pull request. This API is no longer the content
source. See that repo's `docs/SPEC.md` and `docs/ROADMAP.md`.

- [x] "Ask this post" — `POST /api/posts/{slug}/ask` is implemented. A widget
      on the site is an optional item in the site's last phase. It will need
      to read post content from the site instead of this database.
- [ ] Translation — moved to the site repo. An English draft is generated
      for one post at a time and published only after human review. There is
      no auto-translation on publish.
- [ ] LinkedIn sharing — moved to the site repo. The post text is reviewed in
      the publish PR, and a GitHub Action posts it after merge. No MCP server
      is needed for this.
- On hold: "Ask my blog" (pgvector + Voyage embeddings) and `cmd/mcp-server`.
- Not needed: an image upload endpoint. Images are committed with posts in
  the site repo.
