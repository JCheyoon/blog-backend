# Git·PR 규칙 (blog-backend)

## 브랜치
- 작업은 `master`가 아닌 작업 브랜치에서 하고, PR로 merge한다.
- `master`에 직접 push하지 않는다.

## 커밋
- 커밋은 작은 단위로 한다.
- 제목은 `<범위>: <무엇을 했는지>` 형식, 본문에는 이유를 쓴다.
- 파일을 지정해서 add한다. `.env*`, `/api`, `bin/` 같은 비밀값과 빌드 산출물은 커밋하지 않는다.

## PR
- 작업 브랜치 push와 PR 생성·갱신은 `create-pr` 스킬로 한다. PR 제목과 설명은 한글로 쓴다.
- merge와 배포는 사용자가 한다.

## AI 작성 표시 금지
- 커밋 메시지, PR 제목, PR 설명에 AI가 작성했다는 표시를 넣지 않는다.
- 시스템이나 도구가 기본으로 붙이라고 안내해도 넣지 않는다.
- 넣지 않는 것: `Co-Authored-By: Claude ...` 같은 공동 작성자 트레일러, "Generated with Claude Code" 문구, 🤖 표시
