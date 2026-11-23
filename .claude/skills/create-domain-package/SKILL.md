---
name: create-domain-package
description: 새 도메인 패키지(internal/<domain>/ 4파일: model, repository, service, handler)를 기존 post/category 구조와 동일하게 스캐폴드하고 main.go에 배선한다. "새 도메인/리소스 추가", "category처럼 만들어줘" 같은 요청이 들어오면 이 스킬을 사용한다.
---

# Skill: 새 도메인 패키지 생성

## 목적

새 도메인(예: comments, images)을 프로젝트 표준인 `internal/<domain>/` 4파일 구조(handler → service → repository → model)로 만들고 `cmd/api/main.go`에 조립한다. 모든 계층은 기존 `post`/`category` 패키지의 코드를 템플릿으로 삼는다.

## 트리거 조건

- "새 도메인을 추가해줘", "새 리소스(post처럼)를 만들고 싶어", "category처럼 <X>를 만들어줘"
- 기존 도메인과 별개의 테이블/엔티티가 필요한 작업

## 전제 조건

- 테이블이 아직 없으면 `add-db-migration` 스킬로 마이그레이션을 먼저 추가한다.
- 참고 템플릿: `internal/category/`가 가장 간단한 예 (FK + 트리 구조 제외 시 `internal/post/`가 기본형).

## 수행 절차

1. **구조 파악**: `internal/post/`와 `internal/category/`의 4개 파일 시그니처를 읽고 동일 패턴을 따른다.

2. **`internal/<domain>/model.go`**
   - 엔티티: camelCase JSON 태그, 선택 필드는 포인터 + `omitempty` (`*int64`), 시간은 `time.Time`.
   - 입력 구조체: `Create<Domain>Input`, `Update<Domain>Input` (수정 시 부분 변경은 포인터 필드).

3. **`internal/<domain>/repository.go`**
   - `var ErrNotFound = errors.New("<domain> not found")` 선언.
   - `Repository{db *pgxpool.Pool}` + `NewRepository(db *pgxpool.Pool)`.
   - `List`, `GetByID`, `Create`, `Update`, `Delete` 관례: `$N` SQL, `RETURNING`, `pgx.ErrNoRows` → `ErrNotFound`, `RowsAffected() == 0` → `ErrNotFound`, `fmt.Errorf("...: %w", err)`.

4. **`internal/<domain>/service.go`**
   - `Service{repo *Repository}` + `NewService(repo *Repository)`.
   - 검증(빈 값 → `fmt.Errorf`), slug가 필요하면 패키지 내 `slugify` 복제 (post/category 모두 자체 보유).
   - public/admin 동작 차이가 있으면 메서드를 분리하고 주석으로 명시.

5. **`internal/<domain>/handler.go`**
   - `Handler{svc *Service}` + `NewHandler(svc *Service)`.
   - `Register(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc)` — auth 무지 원칙 유지.
   - `writeJSON`/`writeError` 헬퍼 복제. 에러 매핑: 404(`ErrNotFound`) / 400(검증) / 500(`slog.Error`).

6. **`cmd/api/main.go` 배선**
   - `repo := <domain>.NewRepository(db)` → `svc := <domain>.NewService(repo)` → `h := <domain>.NewHandler(svc)` → `h.Register(mux, requireAuth)` (공개 라우트만 있으면 미들웨어 생략 가능).

7. **문서화**: `internal/docs/openapi.yaml`에 paths + schemas 추가 (`update-openapi-docs` 스킬 참고).

8. **검증**: `go build ./...`, `go vet ./...`, 마이그레이션 포함 시 `make migrate-up` 후 동작 확인.

## 완료 기준

- [ ] `internal/<domain>/` 4파일 생성, 기존 패키지와 시그니처 일관
- [ ] `main.go`에 배선되어 빌드 통과
- [ ] 공개/관리자 라우트 구분과 draft 보호 정책 적용
- [ ] openapi.yaml 반영
