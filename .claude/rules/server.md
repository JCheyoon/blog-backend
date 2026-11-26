# 서버 규칙 (blog-backend)

블로그의 공개 읽기 전용 MCP 서버(사이트 SPEC D18)를 만드는 Go HTTP 서버다. DB와 외부 API 키는 없다.

## 구조
```
cmd/api/main.go       서버 실행, 라우트 조립, 로깅·CORS, graceful shutdown
internal/platform/    설정(환경 변수)
internal/ratelimit/   IP당 요청 횟수 제한 미들웨어
```
- `main.go`는 조립만 한다. 새 기능은 `internal/<이름>/` 패키지에 두고 `Register(mux, ...)`로 라우트를 등록한다.
- 미들웨어는 `Register`의 파라미터로 받는다(패키지는 미들웨어 구현을 모른다).
- 로직(검색, 글 찾기)은 HTTP 처리와 분리해 HTTP 없이 테스트할 수 있게 한다.

## 데이터
- 글의 원본은 사이트 저장소의 Markdown이다. 서버는 사이트가 공개하는 글 데이터 파일(`posts.json`)을 HTTP로 읽고, 저장하지 않는다.
- draft는 사이트 빌드에서 빠지므로 서버에 들어오지 않는다.

## API
| 경로 | 용도 |
|---|---|
| `GET /healthz` | 상태 확인 (`{"status":"ok"}`) |
| `POST /mcp` | 공개 MCP 서버(예정). 도구: `search_posts`, `get_post`, `list_posts` |

- 읽기 전용. 쓰기 도구와 인증은 두지 않는다.
- `/mcp`에는 `internal/ratelimit`을 씌우고, 검색어 길이·결과 개수·본문 크기에 상한을 둔다.
- 에러: 일반 HTTP는 `{"error": "<메시지>"}`, MCP는 JSON-RPC 2.0 에러 형식. 500 계열은 고정 문구만 내보내고 상세는 로그로 남긴다.
- 라우트나 MCP 도구를 바꾸면 이 표와 `README.md`의 API 목록을 같이 고친다.

### `/mcp`를 만들 때 할 일 (TODO)
- 횟수 제한 키: Render 같은 프록시 뒤에서는 `RemoteAddr`가 프록시 주소다. 실제 클라이언트 IP를 어떤 헤더(`X-Forwarded-For` 등)에서 어떻게 믿고 꺼낼지 Render 문서로 확인해 정한다.
- CORS: 브라우저 기반 MCP 클라이언트를 받으려면 `Mcp-Session-Id`, `MCP-Protocol-Version` 요청 헤더를 허용하고 `Access-Control-Expose-Headers: Mcp-Session-Id`를 추가한다(MCP 규격으로 확인).

## 코드 스타일
- 표준 라이브러리 우선(`net/http`, `encoding/json`, `log/slog`). 프레임워크·ORM 없음. 의존성 추가는 먼저 묻는다.
- 생성자 `New<Type>`, 필드는 비공개.
- 에러는 `fmt.Errorf("소문자 작업명: %w", err)`로 감싸고, "없음"은 패키지 `ErrNotFound` + `errors.Is`.
- 로그는 `slog`, 메시지는 소문자 동사+명사, 에러는 `"error"` 키.
- JSON 필드는 camelCase. MCP 도구 이름은 snake_case(`search_posts`).
- Go 주석은 영어로 "왜"를 쓴다.
- 테스트는 표준 `testing` 패키지. 첫 테스트를 쓸 때 관례(테이블 드리븐 등)를 여기 적는다.

## 설정과 배포
- 환경 변수: `PORT`(선택, 기본 8080)만. 추가하면 이 문서와 README를 같은 커밋에서 고친다.
- 시크릿은 `.env`에만 둔다(gitignore). `.env.example`은 만들지 않는다.
- `Dockerfile`: `golang:1.22-alpine` 빌더 → `CGO_ENABLED=0` → `alpine:3.20`. 호스팅은 Render 무료 플랜(설정은 사람이 함).

## 검증
- `go build ./...`, `go vet ./...`, 테스트가 있으면 `go test ./...`.
- 이 컴퓨터에는 Go가 없다. Docker로 확인한다: `docker run --rm -v "$PWD":/src -w /src golang:1.22-alpine go build ./...`
- 실행하지 못한 검증은 PR에 "미확인"으로 적는다.
