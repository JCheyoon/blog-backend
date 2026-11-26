# 명명 규칙 (blog-backend)

모든 규칙은 코드에서 확인된 명명을 기준으로 한다.

## 파일/디렉터리

| 대상           | 규칙                                                        | 근거                                                                   |
| -------------- | ----------------------------------------------------------- | ---------------------------------------------------------------------- |
| 도메인 패키지  | `internal/<domain>/`                                        | `internal/post`, `internal/category`, `internal/chat`                   |
| 인프라/공통    | `internal/platform/` (config, db)                           | `internal/platform/config.go`, `db.go`                                 |
| 문서           | `internal/docs/`                                            | openapi 스펙 + Swagger UI                                              |
| 도메인 내 파일 | `model.go`, `repository.go`, `service.go`, `handler.go` 4종 | `internal/post/`, `internal/category/`                                 |
| 실행 진입점    | `cmd/<name>/main.go`                                        | `cmd/api`                                                              |
| 마이그레이션   | `NNNNNN_<스네이크 설명>.up.sql` / `.down.sql` 쌍            | `migrations/000001_create_posts_table.up.sql` 등                       |

## Go 식별자

- **생성자**: `New<Type>` — `NewHandler`, `NewService`, `NewRepository`, `NewClient`, `NewRateLimiter`, `NewPostgresPool`
- **Repository 메서드**: CRUD 동사 — `List`, `GetBySlug` (+ 인자로 동작 변형: `GetBySlug(ctx, slug, publishedOnly)`). 쓰기가 생기면 `GetByID`, `Create`, `Update`, `Delete`
- **Service 메서드**: 비즈니스 의미의 이름 — `Tree`, `List`, `GetBySlug`
- **Handler 메서드**: 소문자 단일 동사(명사) — `list`, `getBySlug`, `ask`, `tree`
- **도메인 에러**: `ErrNotFound` (패키지 레벨)
- **입력 구조체**: `Create<Domain>Input`, `Update<Domain>Input` (2026-11-25에 지운 관리자 쓰기 API에서 쓰던 관례. 새 쓰기 API를 만들면 따른다)
- **핸들러 의존 인터페이스**: 소문자 동사/명사 — `asker` (post.Handler가 chat 의존성에 대해 정의)
- **내부 HTTP 타입**: 소문자 요청/응답 구조체 — `askRequest`, `askResponse`

## API (JSON)

- **JSON 필드는 camelCase**: `createdAt`, `updatedAt`, `categoryId`, `parentId`
- **선택/Nullable 필드**: 포인터 + `omitempty` — `CategoryID *int64 json:"categoryId,omitempty"`, `Children []Category json:"children,omitempty"`
- **엔티티와 입력 분리**: 응답에는 엔티티(`Post`), 요청에는 입력 구조체(`PostInput` 계열) 사용

## 데이터베이스

- **컬럼은 snake_case**: `created_at`, `updated_at`, `parent_id`, `category_id`, `published`
- **테이블은 복수형**: `posts`, `categories`
- **인덱스 명명**: `idx_<table>_<column(s)>` — `idx_posts_published_created_at`, `idx_posts_tags`, `idx_categories_parent_id`
- **마이그레이션 이름**: `<동작>_<대상>` — `create_posts_table`, `add_category_to_posts`, `enable_rls`, `private_posts_rls`

## 라우트

- **공개**: `/api/<resource>` — `GET /api/posts`, `GET /api/posts/{slug}`, `GET /api/categories`
- **동작별**: `/api/posts/{slug}/ask`, `/healthz`, `/docs`
- 경로 파라미터: slug는 `{slug}`, ID는 `{id}` (Go 1.22 문법)

## TODO

- **복수/단수 리소스명**: `posts`(복수) vs `categories`(복수)는 일관되지만, 신규 리소스에 적용할 명시적 규칙은 문서화되어 있지 않다.
