# 운영·동시성·보안 점검

## 동시성

- `http.Client`는 Go 표준 구현 기준 concurrent-safe이며, Client는 생성 후 설정을 변경하지 않는 방식으로 사용합니다.
- Client 설정 필드는 외부에 노출하지 않고 `With...` 옵션으로만 설정하도록 했습니다. 실행 중 인증 토큰·HTTP client·Base URL을 바꾸는 race를 방지합니다.
- `SocketSession`의 URL과 WebSocket dialer도 생성 시 고정됩니다.
- `ChatBot.Run`의 callback은 소켓 수신 루프에서 순차적으로 실행됩니다. callback에서 오래 걸리는 작업이나 외부 I/O를 직접 수행하지 말고 worker/channel로 넘기는 것이 안전합니다.
- 여러 goroutine이 같은 Client를 사용해 API를 호출하는 것은 가능하지만, callback이 공유 상태를 수정한다면 애플리케이션 쪽에서 mutex 또는 channel을 사용해야 합니다.
- `UserChatBot`은 같은 인스턴스의 중복 `Run`을 거부합니다. 여러 프로세스에서 같은 사용자의 일회용 Refresh Token을 동시에 사용하지 않도록 저장소 또는 분산 lock으로 직렬화해야 합니다.

검증 명령:

```bash
go test -race ./...
```

## 보안

- Access Token, Refresh Token, Client Secret은 로그·에러 메시지·소스 코드에 기록하지 말고 환경변수 또는 Secret Manager를 사용합니다.
- Refresh Token은 일회용이므로 갱신 응답으로 반환된 새 값을 원자적으로 저장합니다.
- 사용자별 token pair를 섞지 마세요. Session 구독과 채팅 전송의 주체는 Access Token 사용자입니다.
- Session URL에는 연결 인증 정보가 포함될 수 있으므로 장기간 저장하거나 로그에 남기지 않습니다.
- `NewSocketSession`은 제공된 Session URL에 연결하므로 신뢰할 수 있는 치지직 Session API 응답만 전달해야 합니다.
- 사용자 입력은 `SendChat` 전에 100 rune 및 valid UTF-8 검사를 수행합니다. JSON은 표준 `encoding/json`으로 직렬화합니다.
- 기본 API endpoint는 HTTPS이며, 운영 환경에서 `WithBaseURL`로 임의의 평문 또는 제3자 endpoint를 지정하지 마세요.
- 기본 HTTP client는 30초 timeout을 사용하고 redirect를 따르지 않아 인증 헤더가 다른 endpoint로 전달되지 않도록 합니다. `WithHTTPClient` 사용 시 동일한 정책을 직접 적용하세요.
- `UserChatBot`은 만료 예정 또는 `INVALID_TOKEN` 응답 시 토큰을 자동 갱신하여 `TokenStore`에 저장합니다. 운영 환경에서는 암호화된 영구 저장소와 사용자별 분산 lock을 제공하고, 그 밖의 Client 사용에서는 애플리케이션이 토큰 수명과 저장·폐기를 관리해야 합니다.
