# 아키텍처 규칙 (blog-backend)

## 핵심: 도메인별 3계층 + 조립은 main에서만

각 도메인 패키지는 `handler → service → repository` 3계층과 model을 하나의 디렉터리에 둔다. `main.go`는 이들을 조립(wiring)하는 역할만 한다.

- 근거: `internal/post/`(model.go, repository.go, service.go, handler.go), `internal/category/` 동일 구조, `cmd/api/main.go` 주석 "main.go only assembles them"

### 계층 책임 (코드에서 확인된 구분)

| 계층            | 책임                                           | 금지               |
| --------------- | ---------------------------------------------- | ------------------ |
| `model.go`      | 엔티티 + 입력 구조체, JSON 태그                | 로직 없음          |
| `repository.go` | 순수 SQL 데이터 접근                           | 비즈니스 규칙 없음 |
| `service.go`    | 검증, slug 생성, 공개/비공개 정책 결정         | HTTP 지식 없음     |
| `handler.go`    | JSON 직렬화, 상태 코드, 에러 매핑, 라우트 등록 | SQL 없음           |

- 근거: `internal/post/service.go` 상단 주석 "handlers stay thin and the repository stays a pure data-access layer"

## 조립 규칙 (cmd/api/main.go)

- 의존성은 생성자 인자로 주입한다: `repo → svc → handler` 순서로 생성 후 `Register(mux, ...)` 호출.
- 근거: `cmd/api/main.go`의 `post.NewRepository(db)` → `post.NewService(postRepo)` → `post.NewHandler(postSvc, chatClient)`

### 패키지 간 무지(ignorance) 원칙

- 도메인 패키지는 **auth 구현을 알지 못한다**. 인증 미들웨어는 `Register`의 파라미터로 주입받는다.
  - `post.Handler.Register(mux, authMiddleware, askRateLimit)` — /ask 라우트만 rate limit 적용
  - `category.Handler.Register(mux, authMiddleware)`
  - `auth.Handler.Register(mux)` — 인증 불필요
- 근거: `internal/post/handler.go` Register 주석 "this package stays unaware of how auth works"

### 미들웨어는 함수형 데코레이터

- `func(http.HandlerFunc) http.HandlerFunc` 형태의 순수 함수 체인: `withCORS(withLogging(mux))`.
- 근거: `cmd/api/main.go`, `internal/auth/middleware.go`

## 공개/비공개 (draft) 분리 규칙

- 공개 라우트(`/api/...`)는 `published = true`인 게시물만 반환한다. 초안(draft)은 절대 공개 경로로 새지 않는다.
- 관리자 라우트(`/api/admin/...`)는 auth 미들웨어 뒤에 있으며 초안을 포함한다.
- **정책은 repository의 `publishedOnly` 파라미터로 구현**하고, service가 기본값을 결정한다 (예: `GetBySlug`은 항상 published-only, `GetBySlugAdmin`은 전체).
- 근거: `internal/post/repository.go` (`AND (NOT $2 OR published = true)`), `internal/post/service.go`, `internal/post/handler.go` 라우트 목록

## 라이프사이클 (cmd/api/main.go)

- `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`으로 graceful shutdown.
- `srv.Shutdown`에 5초 타임아웃.
- 서버 타임아웃: Read 10s / Write 10s.
- 로거: `slog.New(slog.NewJSONHandler(os.Stdout, nil))` 후 `slog.SetDefault`.

## 향후 재사용 계획 (현재 아키텍처의 존재 이유)

- README Roadmap에 `cmd/mcp-server`가 `post.Service`를 직접 import해 MCP 도구로 노출할 계획이 명시되어 있다. 따라서 **비즈니스 로직은 반드시 service 계층**에 있어야 하며 handler에 넣으면 안 된다.
- 근거: `README.md` Project layout, `cmd/api/main.go` 주석

## TODO

- **테스트 구조**: 테스트 파일이 없어 레이어별 테스트 전략(테이블 드리븐/통합 테스트 등)이 정의되지 않았다.
- **커밋 컨벤션**: git 로그에 `test`, `debug`, `fix yml` 같은 임시 커밋이 다수 존재해 명확한 컨벤션을 확인할 수 없다.
