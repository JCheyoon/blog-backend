---
name: create-pr
description: 현재 작업 브랜치의 PR 초안(한글 제목·설명)을 만들어 사용자에게 보여 주고, 사용자가 명시적으로 승인한 뒤에만 push하고 master 대상 PR을 연다. "PR 올려줘", "PR 만들어줘", "push하고 PR" 요청에 사용한다.
---

# PR 작성

규칙(승인, AI 표시 금지, 커밋하지 않을 파일, PR을 따로 열지 않는 변경)은 `.claude/rules/git-workflow.md`를 따른다. 이 스킬은 절차만 적는다.

## 1. 사전 확인 (하나라도 실패하면 멈추고 보고한다)
- 브랜치가 `master`가 아니고, `git status`가 깨끗하다.
- `git fetch origin` 뒤 `git log origin/master..HEAD`에 커밋이 있다.
- AI 작성 표시가 없다: `git log origin/master..HEAD --format=%B | grep -iE 'co-authored-by: *claude|generated with'`
- 커밋하면 안 되는 파일이 없다: `git diff --name-only origin/master...HEAD`

## 2. 검증
- `.go`가 바뀌었으면 `.claude/rules/server.md`의 "검증"을 실행한다. 문서만 바뀌었으면 생략하고 PR에 그렇게 적는다.
- 실패하면 PR을 열지 않고 보고한다.

## 3. 초안
- 제목: 한글 한 줄, `<범위>: <무엇을 했는지>`.
- 본문:

```markdown
## 요약
(2~3문장. 무엇을, 왜)

## 변경 내용
| 커밋 | 무엇을 | 왜 |
|---|---|---|

## 영향 범위
- API: (라우트, MCP 도구 변경 여부)
- 배포: (Render 환경 변수·배포 영향. 모르면 "확인 필요")

## 확인한 것 / 확인하지 못한 것
- 확인: (실제로 실행한 검증과 결과)
- 미확인:
```

## 4. 사용자 확인 (생략 불가)
브랜치 → base, 커밋 목록, 제목과 본문 전문, 실행할 명령을 보여 주고 승인을 기다린다. 고쳐 달라고 하면 다시 보여 준다.

## 5. push와 PR (승인 후에만)
1. `git push -u origin <브랜치>` (force push 금지)
2. 본문을 임시 파일에 저장한다.
3. 열린 PR이 없으면 `gh pr create --base master --head <브랜치> --title "<제목>" --body-file <파일>`, 있으면 `gh pr edit <번호> --title "<제목>" --body-file <파일>`
4. PR URL을 알린다.
