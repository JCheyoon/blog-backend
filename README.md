# blog-backend

A small Go HTTP server for [CHEYOON](https://jcheyoon.github.io), a personal tech blog.
It will host a public, read-only MCP server so readers can connect the blog to
their own AI app and ask it to search and summarize posts.

Posts are written as Markdown in the site repository
([JCheyoon.github.io](https://github.com/JCheyoon/JCheyoon.github.io)). This
server stores nothing: it reads the public post data that the site build
publishes. No database, no external API keys.

## Stack

| Concern | Choice | Why |
| --- | --- | --- |
| Routing | stdlib `net/http` (Go 1.22+) | Method + path-pattern routing ships in stdlib |
| Data | Site-published JSON over HTTP | Posts already live in git; nothing to store here |
| Deployment | Docker, single binary | `CGO_ENABLED=0` static build → tiny Alpine image |
| Hosting | Render (free) | |

## Project layout

```
cmd/
  api/         entrypoint: config, routes, middleware, graceful shutdown
internal/
  platform/    config loading
  ratelimit/   per-IP request limit middleware
```

## Local development

```bash
make run          # or: go run ./cmd/api   (PORT is optional, default 8080)
curl localhost:8080/healthz
```

## API

```
GET  /healthz
POST /mcp        (planned) public read-only MCP server: search_posts, get_post, list_posts
```

## History

On 2026-11-26 the database (Supabase), REST API, "Ask this post" endpoint,
admin API, and OpenAPI docs were removed. The site moved to Markdown in git,
so they were no longer used. They remain in git history.
