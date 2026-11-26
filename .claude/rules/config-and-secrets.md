# 설정 및 시크릿 규칙 (blog-backend)

## 환경 변수 계약 (internal/platform/config.go)

| 변수                      | 필수/선택          | 용도                                                     |
| ------------------------- | ------------------ | -------------------------------------------------------- |
| `DATABASE_URL`            | 필수               | pgx 풀 연결 문자열                                       |
| `ANTHROPIC_API_KEY`       | 필수               | Anthropic Messages API 키                                |
| `PORT`                    | 선택 (기본 `8080`) | HTTP 리스닝 포트                                         |

- **필수 검증**: `required := map[string]string{...}` 루프로 빈 값 검사 → 누락 시 `os.Exit(1)` (main에서).
- **선택 값**: `getEnv(key, fallback)` 헬퍼 사용.

## 시크릿 취급 규칙

- 시크릿은 **`.env`에만** 두고 코드/설정 파일에 하드코딩하지 않는다. `.env`는 gitignore되어 있다 (`.gitignore`: `.env*`, `bin/`, `/api`).
- `.env`는 **docker-compose(`env_file: .env`)와 Makefile(`include .env; export`)이 로드**한다. Go 앱은 `os.Getenv`로만 읽는다. 두 도구가 그대로 읽으므로 값 뒤에 주석을 붙이지 않는다.
- `JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD_HASH_B64`는 2026-11-25에 인증 코드와 함께 없앴다. `.env`와 Render 환경 변수에 남아 있으면 지워도 된다.
- 변수 목록의 기준은 위 표와 `internal/platform/config.go`다.
  - 변수를 추가하거나 바꾸면 위 표와 README의 로컬 실행 안내를 같은 커밋에서 고친다.
  - `.env.example`은 만들지 않는다. 혼자 쓰는 저장소이기 때문이다(2026-11-23 결정).

## Docker/배포 관련

- `docker-compose.yml`: `api` 서비스 하나, `env_file: .env`, 포트 8080 매핑. Postgres 컨테이너는 없으므로 `DATABASE_URL`은 기존 DB(Supabase 등)를 가리켜야 한다.
- `Dockerfile`: `golang:1.22-alpine` 빌더 → `CGO_ENABLED=0` 정적 빌드 → `alpine:3.20` 런타임.

## TODO

- **배포 환경 (Render) 시크릿**: Render 대시보드의 env 설정 방법은 저장소에 문서화되어 있지 않다.
