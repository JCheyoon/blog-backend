# 코딩 스타일 규칙 (blog-backend)

이 문서의 모든 규칙은 실제 코드에서 확인된 패턴을 근거로 작성했다. 각 항목의 `근거`는 해당 패턴이 관찰된 파일을 가리킨다. 추측성 내용은 포함하지 않았으며, 확인되지 않은 사항은 마지막 `TODO`에 모아 두었다.

## 스택 원칙

- **표준 라이브러리 우선**: `net/http`, `encoding/json`, `log/slog`를 사용한다. 프레임워크(gin/echo/chi), ORM, 외부 SDK는 도입하지 않는다.
  - 근거: `README.md` "Deliberate non-decisions", `go.mod` (의존성은 jwt/pgx/x-crypto 3개뿐)
- **외부 HTTP 통합도 직접 구현**: Anthropic API 클라이언트를 SDK 없이 `net/http`로 직접 작성한다.
  - 근거: `internal/chat/client.go` 주석 "No SDK dependency"
- **정적 빌드**: `CGO_ENABLED=0`로 빌드해 Alpine 이미지에 올린다.
  - 근거: `Dockerfile`

## Go 코드 스타일

- **생성자 패턴**: 구조체는 항상 `New<Type>(...)` 생성자를 통해 만들고, 필드는 비공개로 둔다.
  - 근거: `NewHandler`, `NewService`, `NewRepository`, `NewClient`, `NewRateLimiter`, `NewPostgresPool` — 모든 패키지에서 동일
- **에러 래핑**: DB/외부 호출 에러는 항상 `fmt.Errorf("소문자 작업명: %w", err)`로 래핑한다. 작업명은 로그 메시지와 동일한 표현을 쓴다.
  - 근거: `internal/post/repository.go` (`"list posts: %w"`, `"scan post: %w"`), `internal/chat/client.go` (`"call anthropic api: %w"`)
- **도메인 에러**: "not found" 계열은 패키지 레벨 `var ErrNotFound = errors.New(...)`로 선언하고 `errors.Is`로 비교한다.
  - 근거: `internal/post/repository.go`, `internal/category/repository.go`
- **HTTP 응답 헬퍼**: 패키지마다 `writeJSON(w, status, v)` / `writeError(w, status, msg)`를 자체 정의해 사용한다 (공용 패키지로 추출하지 않음).
  - 근거: `internal/post/handler.go`, `internal/category/handler.go` — 두 곳에 동일한 구현 중복
- **로깅**: `slog` 사용. 메시지는 소문자 동사+명사 형태, `"error"` 키에 에러를 담는다. 예: `slog.Error("list posts", "error", err)`.
  - 근거: `internal/post/handler.go` 전체
- **JSON 디코딩**: `json.NewDecoder(r.Body).Decode(&in)` → 실패 시 400. 인코딩은 `json.NewEncoder(w).Encode(v)`.
  - 근거: `internal/post/handler.go`
- **요청 본문 검증**: `strings.TrimSpace(...) == ""` 검사 후 `fmt.Errorf("title is required")` 형태의 평문 에러를 반환하고, 핸들러가 400으로 매핑한다.
  - 근거: `internal/post/service.go` Create/Update
- **nil 슬라이스 정규화**: 입력이 nil이면 빈 슬라이스로 만들어 JSON에서 `null` 대신 `[]`가 나가게 한다.
  - 근거: `internal/post/service.go` Create (`if in.Tags == nil { in.Tags = []string{} }`)
- **포인터 필드로 부분 수정**: 수정 요청에서 "값을 안 보냄"과 "명시적으로 비움"을 구분할 때 `*string`/`*int64`를 쓴다.
  - 근거: `internal/category/model.go` `UpdateCategoryInput{Name *string}`

## 주석 스타일

- **Go 코드 주석은 영어**, "무엇"보다 "왜(why)"를 설명한다. 미들웨어/서비스 주석에 의도(보안, 비용, 재사용)를 적는 관례가 있다.
  - 근거: `internal/post/handler.go` ask 주석, `internal/post/service.go` 상단 주석, `internal/auth/middleware.go`
- **마이그레이션 SQL 주석은 한국어**를 사용하고, Supabase/RLS 배경을 설명한다.
  - 근거: `migrations/000004_enable_rls.up.sql`, `migrations/000005_private_posts_rls.up.sql`

## TODO (코드베이스에서 확인되지 않음)

- **테스트 관례 없음**: `*_test.go` 파일이 하나도 없다. 테스트 작성 시 규칙을 새로 정의해야 한다.
- **`.env.example` 부재**: `README.md`가 `cp .env.example .env`를 안내하지만 실제 파일이 없다.
- **루트의 `api` 바이너리**: 컴파일된 실행 파일(`api`, ~14MB)이 git에 커밋되어 있고 `.gitignore`(`.env*`, `bin/`)에 포함되지 않는다. 정리/무시 처리 여부 미정.
- **gofmt/goimports 강제 여부**: CI에 린트/포맷 검사가 없어 자동 포맷 도구 규칙을 확인할 수 없다 (코드 자체는 표준 gofmt 형식).
