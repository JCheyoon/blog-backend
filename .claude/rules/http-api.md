# HTTP API 규칙 (blog-backend)

코드에서 관찰된 라우팅, 상태 코드, 에러 계약을 정리한다.

## 라우팅

- **Go 1.22 `http.ServeMux` 메서드+패턴 문법** 사용: `mux.HandleFunc("GET /api/posts/{slug}", h.getBySlug)`.
- 라우트 등록은 각 핸들러의 `Register(mux, ...)` 메서드에서만 수행한다.
- **경로 파라미터**: `r.PathValue("slug")`. ID는 `strconv.ParseInt(r.PathValue("id"), 10, 64)` 후 실패 시 400 (`invalid post id` / `invalid category id`).
- 근거: `internal/post/handler.go`, `internal/category/handler.go`

### 현재 라우트 전체 목록 (코드 확인)

| 메서드/경로                             | 인증           | 용도                              |
| --------------------------------------- | -------------- | --------------------------------- |
| `GET /healthz`                          | -              | 헬스 체크 (`{"status":"ok"}`)     |
| `POST /api/auth/login`                  | -              | JWT 발급                          |
| `GET /api/posts`                        | -              | 공개 목록 (`?tag=`, `?category=`) |
| `GET /api/posts/{slug}`                 | -              | 단건 조회 (draft는 404)           |
| `POST /api/posts`                       | O              | 생성 (201)                        |
| `PUT /api/posts/{id}`                   | O              | 수정                              |
| `DELETE /api/posts/{id}`                | O              | 삭제 (204)                        |
| `POST /api/posts/{slug}/ask`            | - (rate limit) | 포스트 기반 AI 질문               |
| `GET /api/admin/posts`                  | O              | 초안 포함 전체 목록               |
| `GET /api/admin/posts/{id}`             | O              | 초안 포함 단건                    |
| `GET /api/categories`                   | -              | 카테고리 트리                     |
| `GET/POST /api/admin/categories`        | O              | 관리 목록/생성                    |
| `PUT/DELETE /api/admin/categories/{id}` | O              | 수정/삭제                         |
| `GET /openapi.yaml`, `GET /docs`        | -              | OpenAPI 스펙, Swagger UI          |

## 상태 코드 매핑 (관찰된 계약)

| 코드 | 상황                                       | 응답                                              |
| ---- | ------------------------------------------ | ------------------------------------------------- |
| 200  | 조회/수정 성공, 로그인 성공                | 엔티티 or `{"token": ...}`                        |
| 201  | 생성 성공                                  | 생성된 엔티티                                     |
| 204  | DELETE 성공                                | 본문 없음                                         |
| 400  | 잘못된 JSON 본문 / 잘못된 ID / 검증 실패   | `{"error": <서비스 에러 메시지>}`                 |
| 401  | 자격 증명 불일치 / Bearer 누락 / 토큰 무효 | `{"error": "invalid credentials"}` 등             |
| 404  | `ErrNotFound`                              | `{"error": "post not found"}` 등 고정 메시지      |
| 429  | rate limit 초과                            | `{"error": "too many requests, try again later"}` |
| 502  | Anthropic 업스트림 실패                    | `{"error": "failed to get an answer right now"}`  |
| 503  | chat 미설정 (`h.chat == nil`)              | `{"error": "chat feature is not configured"}`     |

- 근거: `internal/post/handler.go`, `internal/auth/handler.go`, `internal/auth/middleware.go`, `internal/chat/ratelimit.go`

## 에러 응답 계약

- 에러 본문은 항상 `{"error": "<메시지>"}` 형태.
  - 도메인 핸들러: `writeError(w, status, msg)` 헬퍼 사용.
  - auth 계열: `http.Error(w, \`{"error":"..."}\`, status)`로 raw JSON 문자열 직접 사용.
- **예상 에러(400/404)는 로그를 남기지 않고**, 예상치 못한 에러만 `slog.Error("<작업명>", "error", err)`로 남긴다.
  - 근거: `internal/post/handler.go` — `getBySlug`의 ErrNotFound 분기에는 로그 없음, 실패 시 500 + `slog.Error("get post", ...)`.
- 500 응답 메시지는 내부 상세를 드러내지 않는 고정 문구 (`failed to list posts` 등), 상세는 로그로만.

## 보안/정책

- **초안 보호**: 공개 라우트는 `publishedOnly=true`로 repository를 호출하고, service가 기본값을 정한다. 초안이 공개 경로에서 404가 되는 것은 000005 RLS 정책과 이중 방어.
- **인증**: `Authorization: Bearer <JWT>` 헤더, `strings.CutPrefix(header, "Bearer ")`로 파싱. 미들웨어는 검증만 하고 사용자 식별 정보를 컨텍스트에 주입하지 않는다 (현재 단일 관리자이므로).
- **Rate limit**: `/api/posts/{slug}/ask`만 `chat.RateLimiter` 적용 — IP당 20회/시간, 키는 `r.RemoteAddr`, 초과 시 429. 비용 발생 라우트에만 적용한다는 의도가 주석에 명시.
- **CORS**: `Access-Control-Allow-Origin: *`, 메서드 `GET, POST, PUT, DELETE, OPTIONS`, 헤더 `Content-Type, Authorization`. `OPTIONS`는 204 즉시 반환.
- **JSON 계약**: 요청/응답 필드는 camelCase. `omitempty`는 포인터 필드에만 사용.

## TODO

- CORS `*` 허용은 주석에 "tighten to your frontend origin in production"으로 남아 있어 배포 오리진 제한 여부 미정.
- 요청 ID/추적 헤더, 압축(gzip), 등 기타 미들웨어는 존재하지 않음.
