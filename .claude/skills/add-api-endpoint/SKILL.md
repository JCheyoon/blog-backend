---
name: add-api-endpoint
description: 기존 도메인에 새 REST 엔드포인트를 추가한다 (handler → service → repository 3계층 + 라우팅 + openapi 동기화). "새 API 엔드포인트", "POST /api/... 추가", "엔드포인트 만들어줘" 같은 요청이 들어오면 이 스킬을 사용한다.
---

# Skill: REST API 엔드포인트 추가

## 목적

blog-backend의 기존 3계층 패턴(handler → service → repository)을 그대로 따라 새 엔드포인트를 추가한다. 라우팅 문법, 에러 계약, 공개/비공개 정책, openapi 문서화까지 프로젝트 관례와 일치시킨다.

## 트리거 조건

- "새 API 엔드포인트를 추가해줘", "`POST /api/xxx` 만들어줘" 등 엔드포인트 추가/수정 요청
- 기존 라우트의 동작 변경(파라미터, 응답 스키마, 인증 여부) 요청
- 신규 도메인(새 디렉터리)이 필요한 경우 → `create-domain-package` 스킬 사용

## 전제 조건

- 대상 도메인 패키지가 `internal/<domain>/`에 이미 존재하는가? (없으면 `create-domain-package` 먼저)
- 요청이 데이터 변경(CRUD)인가, 조회/액션인가?

## 수행 절차

1. **계층 결정**
   - 데이터 조작(생성/수정/삭제/조회): repository → service → handler 순으로 추가.
   - 조회 전용이고 비즈니스 규칙이 없으면 service는 repository에 위임만 한다 (`List` 패턴).

2. **repository.go** — 메서드 추가
   - 백틱 멀티라인 SQL, `$N` 플레이스홀더, `fmt.Errorf("<작업명>: %w", err)` 래핑.
   - 단건 조회: `pgx.ErrNoRows` → `ErrNotFound`. 쓰기: `RETURNING`으로 서버 생성 필드 회수. 삭제: `RowsAffected() == 0` → `ErrNotFound`.
   - draft 노출 위험이 있으면 `publishedOnly` 파라미터 추가 (`(NOT $N OR published = true)` 패턴).

3. **service.go** — 메서드 추가
   - 검증 먼저: `strings.TrimSpace(...) == ""` → `fmt.Errorf("... is required")`.
   - nil 슬라이스 정규화 (`if in.Tags == nil { in.Tags = []string{} }`), 필요 시 `slugify`.
   - admin 전용(초안 포함) 메서드는 주석으로 admin-only임을 명시 (`GetBySlugAdmin` 패턴).

4. **handler.go** — 메서드 + 라우트 등록
   - 메서드명: 소문자 단일 동사 (`create`, `list`, `ask` 등).
   - `json.NewDecoder(r.Body).Decode(&in)` → 실패 시 400. 서비스 호출 → `writeJSON`/`writeError`.
   - 에러 매핑: `errors.Is(err, ErrNotFound)` → 404, 검증 에러 → 400(`err.Error()`), 그 외 → 500 + `slog.Error("<작업명>", "error", err)`.
   - `Register`에 라우트 추가:
     - 공개: `mux.HandleFunc("GET /api/posts/{slug}", h.xxx)`
     - 관리자: `mux.HandleFunc("...", authMiddleware(h.xxx))`, 경로는 `/api/admin/...` 관례
     - 비용 발생 라우트(외부 API 호출): rate limit 미들웨어 파라미터 사용 (post 패키지의 `askRateLimit` 패턴)

5. **openapi.yaml 동기화** — `update-openapi-docs` 스킬 참고.

6. **검증**
   - `go build ./...` 또는 `make build` 통과.
   - `go vet ./...`.
   - 가능하면 로컬에서 `make run` 후 curl로 수동 확인 (로그인 → 토큰 → 엔드포인트 호출).

## 완료 기준

- [ ] go build / go vet 통과
- [ ] 공개/관리자 분리, draft 노출 여부 확인
- [ ] 에러 응답이 `{"error": ...}` 계약을 따름
- [ ] openapi.yaml에 라우트와 스키마가 반영됨
