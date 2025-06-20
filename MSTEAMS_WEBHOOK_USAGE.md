# Microsoft Teams Webhook 사용법

Atlantis에서 Microsoft Teams로 Terraform apply 결과를 알림으로 받을 수 있습니다.

## 설정 방법

### 1. MS Teams에서 Incoming Webhook 설정

1. MS Teams에서 알림을 받을 채널로 이동
2. 채널 설정 → 커넥터 → Incoming Webhook 추가
3. Webhook URL 복사 (예: `https://outlook.office.com/webhook/...`)

### 2. Atlantis 서버 설정

#### YAML 설정 파일 사용

```yaml
webhooks:
  - event: apply
    workspace-regex: ".*"
    branch-regex: ".*"
    kind: msteams
    url: "https://outlook.office.com/webhook/your-webhook-url"
```

#### 환경변수 사용

```bash
export ATLANTIS_WEBHOOKS='[{"event":"apply","workspace-regex":".*","branch-regex":".*","kind":"msteams","url":"https://outlook.office.com/webhook/your-webhook-url"}]'
```

#### 명령행 인수 사용

```bash
atlantis server \
  --webhooks='[{"event":"apply","workspace-regex":".*","branch-regex":".*","kind":"msteams","url":"https://outlook.office.com/webhook/your-webhook-url"}]'
```

## 설정 옵션

- `event`: 현재 `apply`만 지원
- `workspace-regex`: 알림을 보낼 워크스페이스 정규식 (예: `production.*`)
- `branch-regex`: 알림을 보낼 브랜치 정규식 (예: `main|master`)
- `kind`: `msteams`로 설정
- `url`: MS Teams Incoming Webhook URL

## 메시지 형식

MS Teams로 전송되는 메시지에는 다음 정보가 포함됩니다:

- **성공/실패 상태**: 색상으로 구분 (성공: 초록색, 실패: 빨간색)
- **Repository**: 저장소 이름
- **Workspace**: Terraform 워크스페이스
- **Branch**: 브랜치 이름
- **User**: 실행한 사용자
- **Directory**: Terraform 파일이 있는 디렉토리
- **Project**: 프로젝트 이름 (설정된 경우)
- **Pull Request**: PR 링크

## 예시 설정

### 프로덕션 워크스페이스만 알림

```yaml
webhooks:
  - event: apply
    workspace-regex: "production"
    branch-regex: ".*"
    kind: msteams
    url: "https://outlook.office.com/webhook/your-webhook-url"
```

### 메인 브랜치만 알림

```yaml
webhooks:
  - event: apply
    workspace-regex: ".*"
    branch-regex: "main|master"
    kind: msteams
    url: "https://outlook.office.com/webhook/your-webhook-url"
```

### 여러 채널에 알림

```yaml
webhooks:
  - event: apply
    workspace-regex: "production"
    branch-regex: ".*"
    kind: msteams
    url: "https://outlook.office.com/webhook/production-webhook-url"
  - event: apply
    workspace-regex: "staging"
    branch-regex: ".*"
    kind: msteams
    url: "https://outlook.office.com/webhook/staging-webhook-url"
```

## 문제 해결

### 메시지가 전송되지 않는 경우

1. Webhook URL이 올바른지 확인
2. `workspace-regex`와 `branch-regex`가 현재 워크스페이스/브랜치와 매치되는지 확인
3. Atlantis 로그에서 오류 메시지 확인

### 테스트 방법

Atlantis 서버 로그에서 다음과 같은 메시지를 확인할 수 있습니다:

```
[INFO] Sending webhook to MS Teams: https://outlook.office.com/webhook/...
```

오류가 발생한 경우:

```
[WARN] error sending webhook: ...
```

## 지원되는 다른 Webhook 타입

- `slack`: Slack 채널로 알림
- `http`: 일반 HTTP 엔드포인트로 JSON 전송
- `msteams`: Microsoft Teams로 알림 (새로 추가됨)
