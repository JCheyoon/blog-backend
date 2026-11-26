# Project Guide (blog-backend)

## 현재 상태 (먼저 읽을 것)

- Go 표준 라이브러리로 만든 REST API다. DB는 Supabase Postgres, 호스팅은 Render 무료 플랜이다.
- **블로그 글의 원본은 이 저장소가 아니다.** 글은 사이트 저장소([JCheyoon.github.io](https://github.com/JCheyoon/JCheyoon.github.io))의 Markdown이 기준이다(사이트 `docs/SPEC.md` D1). DB에는 글이 없다(2026-11-23 확인).
- 이 저장소는 유지보수 모드다.
  - 남은 역할: "이 글에 질문하기"(`POST /api/posts/{slug}/ask`). 사이트에서 쓰려면 글 본문을 DB 대신 사이트에서 가져오게 바꿔야 한다(사이트 로드맵 Phase 5, 선택).
  - 보류: "Ask my blog", `cmd/mcp-server`.
  - 쓰기 API는 없다. 관리자 화면(은퇴한 SPA)이 쓰던 로그인·글/카테고리 쓰기 API와 인증 코드는 2026-11-25에 지웠다. 공개 읽기(`GET /api/posts`, `GET /api/posts/{slug}`, `GET /api/categories`)와 질문하기만 남았다.
  - 새 기능은 사이트 로드맵(`docs/ROADMAP.md`)에 있을 때만 만든다.
- push와 PR은 `create-pr` 스킬로 한다(한글 PR, AI 작성 표시 없음). push 전에 반드시 PR 초안을 보여 주고 승인을 받는다. merge와 배포는 사람이 한다(`.claude/rules/git-workflow.md`).

## 작업 원칙

모든 작업에 적용한다. LLM이 자주 하는 실수(근거 없는 가정, 과한 설계, 요청과 무관한 수정)를 줄이기 위한 원칙이다.

### 1. 코딩 전에 생각한다
가정하지 않는다. 헷갈리는 것을 숨기지 않는다. 장단점을 드러낸다.
- 가정은 명시한다. 확실하지 않으면 추측하지 말고 묻는다.
- 해석이 여러 가지면 하나를 몰래 고르지 말고, 해석들을 제시한다.
- 더 단순한 방법이 있으면 말한다. 필요하면 반대 의견을 낸다.
- 헷갈리면 멈추고, 무엇이 불분명한지 말하고 묻는다.
- 이 저장소에서는:
  - API 계약(라우트, JSON 필드, 상태 코드)이나 DB 스키마를 바꾸는 결정은 구현 전에 묻는다.
  - 사이트와 맞물리는 작업(질문 위젯 등)은 사이트 SPEC을 먼저 확인한다.

### 2. 단순함이 먼저다
문제를 푸는 최소한의 코드만 쓴다. 미래를 짐작한 코드는 쓰지 않는다.
- 요청받지 않은 기능은 넣지 않는다.
- 한 번만 쓰는 코드에 추상화를 만들지 않는다.
- 요청받지 않은 "유연성"이나 "설정 가능성"을 넣지 않는다.
- 일어날 수 없는 상황에 대한 에러 처리를 넣지 않는다.
- 200줄을 50줄로 쓸 수 있으면 다시 쓴다.
- 이 저장소에서는: 표준 라이브러리 우선, ORM·프레임워크·SDK 없음이라는 기존 선택을 유지한다(README "Deliberate non-decisions"). 의존성을 추가하려면 먼저 묻는다.
- 기준: 시니어 엔지니어가 보고 "너무 복잡하다"고 할까? 그렇다면 단순하게 고친다.

### 3. 수술하듯 고친다
꼭 필요한 곳만 건드린다. 내가 만든 흔적만 정리한다.
- 주변 코드, 주석, 포맷을 "개선"하지 않는다.
- 고장 나지 않은 것을 리팩터링하지 않는다.
- 내 취향과 달라도 기존 스타일에 맞춘다.
- 관련 없는 죽은 코드를 발견하면 알리기만 하고 지우지 않는다.
- 내 변경 때문에 쓰이지 않게 된 import, 변수, 함수는 지운다. 원래 있던 죽은 코드는 요청이 없으면 지우지 않는다.
- 이 저장소에서는:
  - 패키지마다 `writeJSON`/`writeError`(`coding-style.md`), `slugify`(`create-domain-package` 스킬)를 복제해 둔 것은 의도된 관례다. 공통 패키지로 "정리"하지 않는다.
  - 기존 마이그레이션 파일은 고치지 않는다(append-only, `database.md`).
- 기준: 바뀐 모든 줄이 사용자의 요청으로 설명되어야 한다.

### 4. 목표를 정하고 검증될 때까지 반복한다
성공 기준을 먼저 정한다. 지시를 검증 가능한 목표로 바꾼다.
- 검증 명령: `go build ./...`, `go vet ./...`. 테스트가 생기면 `go test ./...`도 실행한다.
- 아직 `*_test.go`가 없다. 버그를 고치거나 검증을 추가할 때는 표준 `testing` 패키지로 재현 테스트를 먼저 쓰고, 통과하게 만든다. 첫 테스트를 쓸 때 테스트 관례를 `coding-style.md`의 TODO 자리에 함께 정한다.
- 이 컴퓨터에는 Go가 설치되어 있지 않다(2026-11-23 확인). 검증 순서:
  1. Go를 설치했다면 위 명령으로 검증한다.
  2. 없으면 Docker를 켜고 `docker build .`로 컴파일을 확인한다. Dockerfile이 Go 1.22 이미지에서 빌드한다.
  3. 둘 다 못 했으면 PR에 "미확인"으로 적는다. 실행하지 않은 검증을 통과라고 쓰지 않는다.
- 라우트나 스키마를 바꾸면 `internal/docs/openapi.yaml`을 같이 고친다(`update-openapi-docs` 스킬).
- 마이그레이션은 up/down 쌍과 번호를 확인한다. 실제 DB 적용은 사람이 한다.
- 여러 단계 작업은 짧은 계획을 먼저 적는다.
  ```
  1. [단계] → 확인: [무엇으로]
  2. [단계] → 확인: [무엇으로]
  ```

### 적용 범위
- 이 원칙은 속도보다 신중함을 우선한다. 오타나 한 줄 수정처럼 사소한 작업에는 판단껏 적용한다.
- 목표는 사소한 작업을 느리게 만드는 것이 아니라, 큰 작업에서 비싼 실수를 줄이는 것이다.

## 구조와 명령

- `cmd/api/`: 진입점. 설정, DB, 핸들러를 조립하고 서버를 띄운다
- `internal/<도메인>/`: `model`, `repository`(SQL), `service`(비즈니스 규칙), `handler`(HTTP) 4개 파일. `main.go`는 조립만 한다(`architecture.md`)
- `internal/platform/`: 설정과 DB 풀, `internal/docs/`: OpenAPI와 Swagger UI, `migrations/`: SQL 마이그레이션
- 명령(`Makefile`): `make run`, `make build`, `make docker-up`, `make migrate-up`, `make migrate-down`
  - Makefile이 `.env`를 읽으므로 `.env`가 없으면 make가 실패한다. 필요한 변수는 README의 Local development와 `config-and-secrets.md`에 있다.

## Claude Workspace (.claude/)

- 규칙(`.claude/rules/`): `architecture`, `coding-style`, `naming-convention`, `http-api`, `database`, `config-and-secrets`, `git-workflow`
- 스킬(`.claude/skills/<이름>/SKILL.md`): `add-api-endpoint`, `add-db-migration`, `create-domain-package`, `update-openapi-docs`, `create-pr`
- `.claude/settings.json`: 권한 기본값(allow/ask/deny)
