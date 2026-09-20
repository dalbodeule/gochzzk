# gochzzk 기여 가이드

1. 이슈를 확인하거나 변경 목적을 설명하는 이슈를 작성합니다.
2. Go 1.26.6 이상을 사용하고 변경 범위에 맞는 테스트와 문서를 추가합니다.
3. 아래 명령을 통과시킨 뒤 PR을 작성합니다.

```bash
gofmt -w ./src
go test ./...
go test -race ./...
go vet ./...
```

공식 API와 실측 동작이 다르면 둘을 구분해 문서화하고, 비공식 endpoint는 `src/unofficial`에만 추가해 주세요. Access Token, Refresh Token, Client Secret과 Session URL은 이슈, 테스트 fixture, 로그에 포함하면 안 됩니다.
