# HTTP API 규칙 (blog-backend)

코드에서 관찰된 라우팅, 상태 코드, 에러 계약을 정리한다.

## 라우팅

- **Go 1.22 `http.ServeMux` 메서드+패턴 문법** 사용: `mux.HandleFunc("GET /api/posts/{slug}", h.getBySlug)`.
- 라우트 등록은 각 핸들러의 `Register(mux, ...)` 메서드에서만 수행한다.
- **경로 파라미터**: `r.PathValue("slug")`. ID를 받는 라우트가 생기면 `strconv.ParseInt(r.PathValue("id"), 10, 64)` 후 실패 시 400.
- 근거: `internal/post/handler.go`, `internal/category/handler.go`

### 현재 라우트 전체 목록 (코드 확인)

| 메서드/경로                             | 인증           | 용도                              |
| --------------------------------------- | -------------- | --------------------------------- |
| `GET /healthz`                          | -              | 헬스 체크 (`{"status":"ok"}`)     |
| `GET /api/posts`                        | -              | 공개 목록 (`?tag=`, `?category=`) |
| `GET /api/posts/{slug}`                 | -              | 단건 조회 (draft는 404)           |
| `POST /api/posts/{slug}/ask`            | - (rate limit) | 포스트 기반 AI 질문               |
| `GET /api/categories`                   | -              | 카테고리 트리                     |
| `GET /openapi.yaml`, `GET /docs`        | -              | OpenAPI 스펙, Swagger UI          |

## 상태 코드 매핑 (관찰된 계약)

| 코드 | 상황                                       | 응답                                              |
| ---- | ------------------------------------------ | ------------------------------------------------- |
| 200  | 조회 성공, 질문 응답                       | 엔티티 or `{"answer": ...}`                       |
| 400  | 잘못된 JSON 본문 / 검증 실패               | `{"error": "question is required"}` 등            |
| 404  | `ErrNotFound`                              | `{"error": "post not found"}` 등 고정 메시지      |
| 429  | rate limit 초과                            | `{"error": "too many requests, try again later"}` |
| 502  | Anthropic 업스트림 실패                    | `{"error": "failed to get an answer right now"}`  |
| 503  | chat 미설정 (`h.chat == nil`)              | `{"error": "chat feature is not configured"}`     |

- 근거: `internal/post/handler.go`, `internal/chat/ratelimit.go`

## 에러 응답 계약

- 에러 본문은 항상 `{"error": "<메시지>"}` 형태.
  - 도메인 핸들러: `writeError(w, status, msg)` 헬퍼 사용.
- **예상 에러(400/404)는 로그를 남기지 않고**, 예상치 못한 에러만 `slog.Error("<작업명>", "error", err)`로 남긴다.
  - 근거: `internal/post/handler.go` — `getBySlug`의 ErrNotFound 분기에는 로그 없음, 실패 시 500 + `slog.Error("get post", ...)`.
- 500 응답 메시지는 내부 상세를 드러내지 않는 고정 문구 (`failed to list posts` 등), 상세는 로그로만.

## 보안/정책

- **초안 보호**: 공개 라우트는 `publishedOnly=true`로 repository를 호출하고, service가 기본값을 정한다. 초안이 공개 경로에서 404가 되는 것은 000005 RLS 정책과 이중 방어.
- **인증**: 없다. 관리자 로그인(JWT)과 쓰기 라우트는 2026-11-25에 지웠다.
- **Rate limit**: `/api/posts/{slug}/ask`만 `chat.RateLimiter` 적용 — IP당 20회/시간, 키는 `r.RemoteAddr`, 초과 시 429. 비용 발생 라우트에만 적용한다는 의도가 주석에 명시.
- **CORS**: `Access-Control-Allow-Origin: *`, 메서드 `GET, POST, OPTIONS`, 헤더 `Content-Type`. `OPTIONS`는 204 즉시 반환.
- **JSON 계약**: 요청/응답 필드는 camelCase. `omitempty`는 포인터 필드에만 사용.

## TODO

- CORS `*` 허용은 주석에 "tighten to your frontend origin in production"으로 남아 있어 배포 오리진 제한 여부 미정.
- 요청 ID/추적 헤더, 압축(gzip), 등 기타 미들웨어는 존재하지 않음.
