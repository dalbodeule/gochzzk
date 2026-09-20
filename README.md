# gochzzk

Go `1.26.6` 이상에서 사용할 수 있는 치지직 Open API Wrapper입니다. 공식 REST API와 Socket.IO 세션 API를 한 패키지에서 사용할 수 있도록 구성했습니다.

## 제공 범위

- OAuth 인증 코드 URL, Access Token 발급·갱신·폐기
- User, Channel, Category, Live, Chat API
- Restriction(활동 제한) API
- Client/User Session API
- Socket.IO 기반 채팅·후원·구독 이벤트 수신
- 챗봇 callback 구조체
- 채널 단일 조회 및 라이브 상태 조회 Helper

공식 API의 인증 방식·Scope·응답 구조는 [`docs/api.md`](docs/api.md)에 요약되어 있고, 유형별 상세 문서는 다음과 같이 나뉘어 있습니다.

| 문서 | 내용 |
|---|---|
| [`docs/auth-session.md`](docs/auth-session.md) | OAuth와 Session API |
| [`docs/user.md`](docs/user.md) | User API |
| [`docs/channel.md`](docs/channel.md) | Channel API |
| [`docs/category.md`](docs/category.md) | Category API |
| [`docs/live-chat.md`](docs/live-chat.md) | Live/Chat/Restriction API |
| [`docs/chatbot.md`](docs/chatbot.md) | Socket.IO와 챗봇 구성 |
| [`docs/security.md`](docs/security.md) | 동시성·운영·보안 점검 |
| [`docs/unofficial.md`](docs/unofficial.md) | 비공식 라이브 상태·팔로우 날짜 API |
| [`docs/runtime-findings.md`](docs/runtime-findings.md) | Session 서버 실측 차이와 대응 |

## 설치

```bash
go get github.com/dalbodeule/gochzzk
```

`go.mod`의 Go 버전은 보안 수정이 포함된 `1.26.6`으로 선언되어 있습니다.

## Client 생성

```go
import (
    chzzk "github.com/dalbodeule/gochzzk/src/api"
    "github.com/dalbodeule/gochzzk/src/chatbot"
)

client := chzzk.New(
    chzzk.WithClientCredentials(
        os.Getenv("CHZZK_CLIENT_ID"),
        os.Getenv("CHZZK_CLIENT_SECRET"),
    ),
    chzzk.WithAccessToken(os.Getenv("CHZZK_ACCESS_TOKEN")),
    chzzk.WithUserAgent("my-chzzk-bot/1.0"),
)
```

Client 인증이 필요한 메서드에는 `Client-Id`, `Client-Secret`이 사용되고, 사용자 권한이 필요한 메서드에는 `Authorization: Bearer <token>`이 사용됩니다. 하나의 Client에 두 인증 정보를 모두 넣어도 됩니다.

## 인증 예시

```go
authorizationURL := chzzk.AuthorizationURL(
    clientID,
    "https://example.com/chzzk/callback",
    state,
)
// authorizationURL로 사용자를 이동시킵니다.

token, err := client.ExchangeCode(ctx, codeFromCallback, state)
if err != nil { return err }
client = chzzk.New(
    chzzk.WithClientCredentials(clientID, clientSecret),
    chzzk.WithAccessToken(token.AccessToken),
)
```

Refresh Token은 일회용이므로 `RefreshToken` 호출 결과의 새 Refresh Token을 반드시 저장해야 합니다.

## API 사용 예시

```go
ctx := context.Background()

// Client 인증: 현재 라이브 목록
lives, err := client.GetLives(ctx, 20, "")
if err != nil { return err }
for _, live := range lives.Data {
    fmt.Println(live.ChannelName, live.LiveTitle, live.LiveThumbnailImageURL)
}

// 단일 채널 조회 Helper
channel, err := chzzk.GetChzzkUserInfo(ctx, channelID, clientID, clientSecret)
if err != nil { return err }
if channel != nil { fmt.Println(channel.ChannelName, channel.FollowerCount) }

// 해당 채널의 현재 라이브 여부와 상세 정보
status, err := client.GetChannelLiveStatus(ctx, channelID)
if err != nil { return err }
if status.IsLive && status.Live != nil {
    fmt.Println(status.Live.LiveTitle, status.ThumbnailURL)
}
```

치지직 공개 API에는 채널 단일 라이브 조회 endpoint가 별도로 제공되지 않으므로 `GetChannelLiveStatus`는 `GET /open/v1/lives`의 cursor 목록을 순회합니다. 호출량과 quota를 고려해 적절히 캐시하세요.

## 챗봇 예시

운영 환경에서는 사용자별 Access Token/Refresh Token을 분리하고 사용자마다 `chatbot.UserChatBot`을 하나씩 실행하는 구성을 권장합니다. `UserChatBot`은 token refresh, User Session 발급, 채팅·후원·구독 이벤트 구독을 묶어서 처리합니다. 자세한 예시는 [`docs/chatbot.md`](docs/chatbot.md)를 참고하세요.

```go
session, err := client.CreateUserSession(ctx)
if err != nil { return err }

bot := &chatbot.ChatBot{
    Session: chatbot.NewSocketSession(session.URL),
    OnSystem: func(event chzzk.SystemEvent) error {
        if event.Type == "connected" {
            // 세션당 최대 30개의 이벤트를 구독할 수 있습니다.
            if err := client.SubscribeChat(ctx, event.Data.SessionKey); err != nil { return err }
            if err := client.SubscribeDonation(ctx, event.Data.SessionKey); err != nil { return err }
            return client.SubscribeSubscription(ctx, event.Data.SessionKey)
        }
        return nil
    },
    OnChat: func(event chzzk.ChatEvent) error {
        fmt.Printf("%s: %s\n", event.Profile.Nickname, event.Content)
        return nil
    },
    OnDonation: func(event chzzk.DonationEvent) error {
        fmt.Printf("donation=%s text=%s\n", event.PayAmount, event.DonationText)
        return nil
    },
}

return bot.Run(ctx)
```

Socket.IO 연결은 REST 세션 발급과 분리되어 있습니다. 먼저 세션 URL을 발급하고 연결 완료(`SYSTEM/connected`) 이벤트의 `sessionKey`로 구독 API를 호출합니다.

## 오류 처리

API 오류는 `*chzzk.APIError`로 반환됩니다.

```go
var apiErr *chzzk.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

주요 상태 코드는 `400` 파라미터 오류, `401` 인증 실패, `403` 권한/Scope 부족, `404` 없음, `429` quota 초과, `500` 서버 오류입니다.

## 소스 구조

공개 `src/api` 패키지의 exported `Client` 메서드는 유형별 파일로 분리했습니다. `src/chatbot`은 API 모델을 사용하지만 REST Client를 import하지 않는 독립적인 Socket.IO 계층입니다. `src/internal/request`는 외부에서 import할 수 없는 공통 HTTP 요청 처리 계층입니다.

```text
src/api/auth.go          OAuth 인증
src/api/session.go       Session REST 메서드
src/api/user.go          User 메서드
src/api/channel.go       Channel 메서드
src/api/category.go      Category 메서드
src/api/live.go          Live 메서드
src/api/chat.go          Chat 메서드
src/api/restriction.go   Restriction 메서드
src/api/types.go         API 모델
src/chatbot/socketio.go  Socket.IO transport와 ChatBot
src/chatbot/user_bot.go  사용자별 token refresh와 자동 이벤트 구독
src/internal/request/    내부 HTTP 요청·인증·envelope 처리
src/unofficial/          비공식 live-status/profile-card API
```

## 테스트

외부 치지직 서버에 의존하지 않는 fake HTTP transport와 Socket.IO payload decoder 테스트를 포함합니다.

```bash
go test ./...
go vet ./...
go test -race ./...
```

## 공식 문서

- [CHZZK Developers](https://chzzk.gitbook.io/chzzk)
- [Authorization](https://chzzk.gitbook.io/chzzk/chzzk-api/authorization)
- [Session](https://chzzk.gitbook.io/chzzk/chzzk-api/session)
- [Restriction](https://chzzk.gitbook.io/chzzk/chzzk-api/restriction)

## License

MIT License. 자세한 내용은 [`license.md`](license.md)를 참고하세요.
