---
name: update-openapi-docs
description: API 변경 시 internal/docs/openapi.yaml을 코드와 동기화한다. 엔드포인트 추가/변경/삭제나 요청·응답 스키마 변경 시 사용한다 (add-api-endpoint의 완료 기준 일부).
---

# Skill: OpenAPI 문서 갱신

## 목적

핸드메인테인드(`//go:embed`로 서빙)되는 `internal/docs/openapi.yaml`이 Go 코드의 실제 라우트/스키마와 항상 일치하도록 유지한다. 문서만 최신이면 `/docs`(Swagger UI)와 `/openapi.yaml`이 자동으로 반영된다.

## 트리거 조건

- 엔드포인트 추가/변경/삭제 후 (add-api-endpoint, create-domain-package 완료 기준)
- 요청/응답 JSON 스키마 변경 (필드 추가, 타입 변경)
- 인증 요구사항 변경 (공개 ↔ 관리자)

## 전제 조건

- 변경된 Go 코드(handler.go의 라우트, model.go의 JSON 태그)를 확인할 수 있어야 한다.

## 수행 절차

1. **현재 스펙 확인**: `internal/docs/openapi.yaml` 열기. tags: `auth`, `posts`, `categories`, `chat`.

2. **paths 동기화**: 코드의 `mux.HandleFunc("<METHOD> <path>", ...)` 목록과 일치하는지 대조.
   - 경로 파라미터는 `{slug}`/`{id}` 형태.
   - 공개 라우트: `security` 없음.
   - 관리자 라우트(`/api/admin/...` 또는 auth로 감싼 라우트): `security: [bearerAuth: []]`.

3. **요청 본문**: 인라인 객체보다 `components/schemas`의 Input 스키마 `$ref` 사용 (PostInput 패턴).
   - `required: [title]`처럼 필수 필드 명시.

4. **응답 문서화**: 코드에서 실제로 쓰는 상태 코드만 기술 (200/201/204/400/401/404/429/502/503).
   - 에러 응답은 `$ref: "#/components/schemas/Error"`.
   - 204는 content 없음.

5. **스키마 필드**: Go JSON 태그와 camelCase 일치 확인. `omitempty` 필드는 `required`에 넣지 않는다. `categoryId` 같은 nullable은 타입만 명시.

6. **검증**:
   - YAML 들여쓰기/문법 확인 (`go build ./...` — go:embed로 문법이 틀리면 빌드가 깨지지는 않지만, 가능하면 파서로 확인).
   - 가능하면 `make run` 후 브라우저 `http://localhost:8080/docs`에서 스키마 확인.

## 완료 기준

- [ ] paths/스키마가 Go 코드와 일치 (라우트 누락/잔재 없음)
- [ ] 인증(security) 표시가 실제 미들웨어 적용과 일치
- [ ] 상태 코드가 핸들러의 실제 분기와 일치
