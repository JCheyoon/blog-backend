# 설정 및 시크릿 규칙 (blog-backend)

## 환경 변수 계약 (internal/platform/config.go)

| 변수                      | 필수/선택          | 용도                                                     |
| ------------------------- | ------------------ | -------------------------------------------------------- |
| `DATABASE_URL`            | 필수               | pgx 풀 연결 문자열                                       |
| `JWT_SECRET`              | 필수               | JWT 서명 키 (HS256)                                      |
| `ADMIN_EMAIL`             | 필수               | 로그인 이메일 비교값                                     |
| `ADMIN_PASSWORD_HASH_B64` | 필수               | bcrypt 해시를 **base64 인코딩**한 값 (config에서 디코드) |
| `ANTHROPIC_API_KEY`       | 필수               | Anthropic Messages API 키                                |
| `PORT`                    | 선택 (기본 `8080`) | HTTP 리스닝 포트                                         |

- **필수 검증**: `required := map[string]string{...}` 루프로 빈 값 검사 → 누락 시 `os.Exit(1)` (main에서).
- **선택 값**: `getEnv(key, fallback)` 헬퍼 사용.

## 시크릿 취급 규칙

- 시크릿은 **`.env`에만** 두고 코드/설정 파일에 하드코딩하지 않는다. `.env`는 gitignore되어 있다 (`.gitignore`: `.env*`, `bin/`, `/api`).
- `.env`는 **docker-compose(`env_file: .env`)와 Makefile(`include .env; export`)이 로드**한다. Go 앱은 `os.Getenv`로만 읽는다. 두 도구가 그대로 읽으므로 값 뒤에 주석을 붙이지 않는다.
- `ADMIN_PASSWORD_HASH_B64`에는 평문이 아닌, base64로 감싼 bcrypt 해시를 저장한다. 값은 `go run ./cmd/hashpw "<비밀번호>"` 출력을 그대로 쓴다.
- 변수 목록의 기준은 위 표와 `internal/platform/config.go`다.
  - 변수를 추가하거나 바꾸면 위 표와 README의 로컬 실행 안내를 같은 커밋에서 고친다.
  - `.env.example`은 만들지 않는다. 혼자 쓰는 저장소이기 때문이다(2026-11-23 결정).

## 인증 구현 관례 (internal/auth)

- JWT: HS256, TTL 24시간, `Subject`에 이메일. `ParseToken`은 서명 메서드가 HMAC인지 확인 후 검증 (알고리즘 혼동 공격 방어).
- 비밀번호 비교: `bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))` — 이메일과 해시 비교를 모두 통과해야 토큰 발급.
- 미들웨어: `Authorization: Bearer <token>` 파싱 (`strings.CutPrefix`), 실패 시 401.

## Docker/배포 관련

- `docker-compose.yml`: `api` 서비스 하나, `env_file: .env`, 포트 8080 매핑. Postgres 컨테이너는 없으므로 `DATABASE_URL`은 기존 DB(Supabase 등)를 가리켜야 한다.
- `Dockerfile`: `golang:1.22-alpine` 빌더 → `CGO_ENABLED=0` 정적 빌드 → `alpine:3.20` 런타임.

## TODO

- **배포 환경 (Render) 시크릿**: Render 대시보드의 env 설정 방법은 저장소에 문서화되어 있지 않다.
