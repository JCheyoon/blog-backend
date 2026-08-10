# 데이터베이스 규칙 (blog-backend)

PostgreSQL + pgx (raw SQL, ORM 없음) 관례를 정리한다.

## pgx 사용 관례 (repository 계층)

- `Repository` 구조체가 `*pgxpool.Pool`을 보유하고 `NewRepository(db *pgxpool.Pool)`로 주입받는다.
- **SQL은 백틱 멀티라인 문자열**, 키워드 대문자, `$N` 플레이스홀더. ORM/쿼리 빌더 없음.
- **에러 래핑**: 모든 쿼리 에러는 `fmt.Errorf("<소문자 작업명>: %w", err)`.
- **다중 행**: `rows.Next()` 루프 + `defer rows.Close()` + 마지막에 `rows.Err()` 확인.
- **단건 없음**: `errors.Is(err, pgx.ErrNoRows)` → 패키지 `ErrNotFound`.
- **쓰기**: `RETURNING id, created_at, updated_at` 등 서버 생성 필드를 회수해 입력 구조체에 반영.
- **삭제**: `tag.RowsAffected() == 0`이면 `ErrNotFound`.
- **선택 필터**: 빈 문자열 파라미터로 조건 off — `($1 = '' OR $1 = ANY(p.tags))`.
- **공개/비공개**: `(NOT $3 OR p.published = true)` — publishedOnly가 false면 조건 무시.
- 근거: `internal/post/repository.go`, `internal/category/repository.go` 전체

## 스키마 관례 (migrations)

- **테이블**: `BIGSERIAL PRIMARY KEY`, `TEXT`, `TEXT[]`(tags), `BOOLEAN`, `TIMESTAMPTZ NOT NULL DEFAULT now()`.
  - `created_at`/`updated_at`은 모든 테이블에 `TIMESTAMPTZ DEFAULT now()`.
- **FK**: `REFERENCES <table>(id) ON DELETE SET NULL`.
- **인덱스 관례**:
  - 정렬 조회용 복합 인덱스: `idx_posts_published_created_at ON posts (published, created_at DESC)`
  - 배열 컬럼은 GIN: `idx_posts_tags ON posts USING GIN (tags)`
  - FK 컬럼은 단일 인덱스: `idx_categories_parent_id`, `idx_posts_category_id`
- 근거: `migrations/000001_create_posts_table.up.sql`, `000002`, `000003`

## 마이그레이션 운영 규칙

- golang-migrate 형식: `NNNNNN_<설명>.up.sql` + `.down.sql` **쌍** 필수 (현재 최대 번호: 000005).
- down은 up과 대칭으로 작성 (`DROP TABLE`, `DROP COLUMN`, `DROP INDEX`, `DROP POLICY`).
- **기존 마이그레이션은 수정하지 않는다** (append-only). 변경은 새 버전 파일로.
- 적용/롤백: `make migrate-up` / `make migrate-down` (Makefile이 `$DATABASE_URL` 사용).
- 근거: `migrations/` 디렉터리, `Makefile`

## Supabase RLS 관례 (000004, 000005에서 확인)

- Supabase REST API(anon key) 노출 표면을 RLS로 통제한다.
- **정책**: 공개 읽기용 `CREATE POLICY "Enable read access for all users" ... FOR SELECT USING (true)`.
- **비공개 데이터**: 000005에서 `posts`의 공개 SELECT 정책을 내리고 `USING (published = true)`로 교체 → 초안은 default-deny.
- **Go 백엔드는 postgres 소유자 역할로 접속해 RLS를 우회**하므로 관리자 CRUD에는 영향 없음 — 이를 마이그레이션 주석에 명시하는 관례.
- 마이그레이션 주석은 Supabase 경고/배경을 설명한다.
- 근거: `migrations/000004_enable_rls.up.sql`, `migrations/000005_private_posts_rls.up.sql`

## TODO

- **테스트 DB**: 테스트 파일이 없어 테스트용 DB(격리, 트랜잭션 롤백 등) 전략이 정의되지 않았다.
- **pgvector**: README의 "Ask my blog" 계획에 `pgvector` + 임베딩 검색이 예정되어 있으나 구현 전이라 관련 규칙 없음.
