---
name: add-db-migration
description: golang-migrate 스키마 마이그레이션 쌍(up/down)을 추가하고 이 프로젝트의 인덱스/RLS 관례를 적용한다. "테이블 추가", "컬럼 추가", "스키마 변경", "RLS 정책" 요청 시 사용한다.
---

# Skill: DB 마이그레이션 추가

## 목적

스키마 변경을 이 프로젝트의 마이그레이션 관례(버전 번호, up/down 쌍, 인덱스, Supabase RLS)에 맞게 추가하고 관련 Go 코드(repository/model)까지 동기화한다.

## 트리거 조건

- "테이블을 추가해줘", "posts에 컬럼을 추가", "스키마 변경", "RLS 정책 추가/수정"
- 새 도메인 생성 시 테이블이 필요한 경우 (`create-domain-package`와 함께)

## 전제 조건

- golang-migrate CLI가 설치되어 있어야 로컬 검증 가능 (`make migrate-up`).

## 수행 절차

1. **다음 버전 번호 확인**: `ls migrations/`에서 가장 큰 `NNNNNN`을 찾고 `+1` 사용 (현재 000005까지 존재 → 다음은 000006).

2. **up/down 쌍 생성**: `migrations/NNNNNN_<설명>.up.sql` + `.down.sql` (설명은 `<동사>_<대상>`: `create_xxx_table`, `add_column_to_xxx`).

3. **up.sql 관례 적용**:
   - 컬럼: `BIGSERIAL PRIMARY KEY`, `TEXT`, `BOOLEAN DEFAULT false`, `TIMESTAMPTZ NOT NULL DEFAULT now()`.
   - FK: `REFERENCES <table>(id) ON DELETE SET NULL`.
   - 인덱스: 정렬 조회는 복합 인덱스(`idx_<table>_<col1>_<col2>`), 배열은 GIN, FK 컬럼은 단일 인덱스.
   - 컬럼 추가는 `ALTER TABLE ... ADD COLUMN` (000003 패턴).

4. **down.sql 작성**: up과 대칭 — `DROP TABLE`, `DROP COLUMN`, `DROP INDEX`, `DROP POLICY`.

5. **Supabase RLS 정책 (public 테이블일 때)**:
   - `ALTER TABLE ... ENABLE ROW LEVEL SECURITY;`
   - 공개 읽기: `CREATE POLICY "Enable read access for all users" ... FOR SELECT USING (true);`
   - 비공개(초안 등) 데이터: SELECT 정책을 `USING (published = true)`처럼 좁게 (`000005_private_posts_rls` 패턴).
   - 정책명은 "Enable read ... for all users" 형태.
   - **중요**: Go 백엔드는 postgres 소유자 역할로 접속해 RLS를 우회하므로 관리자 CRUD에 영향 없음 — 이 내용을 한국어 주석으로 명시하는 관례.
   - RLS 전용 마이그레이션도 up/down 쌍으로 (000004/000005처럼).

6. **기존 마이그레이션 수정 금지**: append-only. 변경은 새 버전으로.

7. **Go 코드 동기화**: repository.go 쿼리/컬럼, model.go 필드 반영 (스키마와 코드가 어긋나면 빌드는 되지만 런타임 에러).

8. **검증**: `make docker-up` 후 `make migrate-up` → 필요 시 `make migrate-down`으로 롤백 확인 → `go build ./...`.

## 완료 기준

- [ ] up/down 쌍 생성, 번호가 기존 최대+1
- [ ] 인덱스/FK/RLS 관례 적용
- [ ] repository/model과 스키마 일치
- [ ] 로컬 `migrate-up` 성공 확인 (가능한 경우)
