# Socket.IO 챗봇 구조

치지직 Session 문서는 Socket.IO-client 1.0.0+ 및 2.0.3까지와 WebSocket transport를 안내합니다. Wrapper는 `gorilla/websocket`으로 해당 연결을 수행합니다.

실제 서버 동작은 공식 문서와 일부 차이가 있어 [커뮤니티 실측 자료](https://gist.github.com/fi-xz/69ce1f35ca1b2318a2b410c0d5757e0f)를 방어적으로 반영했습니다. 이 내용은 공식 호환성 보장이 아니라 관측 결과이므로 [`runtime-findings.md`](runtime-findings.md)에 별도로 정리했습니다.

## 권장 구조: 사용자당 하나의 UserChatBot

치지직 User Session은 세션 발급에 사용한 Access Token의 사용자 이벤트만 구독합니다. 따라서 여러 스트리머 계정을 운영하는 애플리케이션은 사용자별 Access Token/Refresh Token과 `UserChatBot` 인스턴스를 각각 유지해야 합니다. 채팅 전송 역시 사용한 Access Token 소유자의 채널 명의로 수행됩니다.

```go
store := chatbot.NewMemoryTokenStore()
_ = store.Save(ctx, userID, chatbot.NewStoredToken(tokenFromOAuth, time.Now()))

bot := &chatbot.UserChatBot{
    UserID: userID,
    ClientID: clientID,
    ClientSecret: clientSecret,
    Tokens: store,
    OnChat: func(event api.ChatEvent) error {
        fmt.Println(event.Profile.Nickname, event.Content)
        return nil
    },
}
return bot.Run(ctx)
```

`UserChatBot.Run`은 다음을 수행합니다.

1. `TokenStore`에서 해당 사용자의 token pair를 읽습니다.
2. 만료 임박 시 Refresh Token으로 갱신하고 새 Access/Refresh Token을 함께 저장합니다.
3. Access Token으로 User Session URL을 발급합니다.
4. `SYSTEM/connected`에서 받은 `sessionKey`로 `CHAT`, `DONATION`, `SUBSCRIPTION`을 모두 구독합니다.
5. REST 호출이 `401 INVALID_TOKEN`이면 한 번 갱신한 뒤 새 세션으로 다시 연결합니다.

Refresh Token은 일회용입니다. 같은 사용자에 대해 여러 `UserChatBot`을 동시에 실행하지 마세요. `MemoryTokenStore`는 단일 프로세스 테스트용이며, 운영 환경에서는 암호화된 저장소와 사용자별 분산 lock을 지원하는 `TokenStore` 구현을 권장합니다.

## 연결 순서

1. `CreateUserSession`으로 연결 URL을 발급합니다.
2. `NewSocketSession(url).Run` 또는 `ChatBot.Run`으로 연결합니다.
3. `SYSTEM/connected` 이벤트의 `sessionKey`를 확인합니다.
4. `SubscribeChat`, `SubscribeDonation`, `SubscribeSubscription`을 호출합니다.
5. callback에서 typed event를 처리합니다.

```go
import (
    api "github.com/dalbodeule/gochzzk/src/api"
    "github.com/dalbodeule/gochzzk/src/chatbot"
)

session, err := client.CreateUserSession(ctx)
if err != nil { return err }

bot := &chatbot.ChatBot{
    Session: chatbot.NewSocketSession(session.URL),
    OnSystem: func(event api.SystemEvent) error {
        if event.Type != "connected" { return nil }
        return client.SubscribeChat(ctx, event.Data.SessionKey)
    },
    OnChat: func(event api.ChatEvent) error {
        fmt.Println(event.Profile.Nickname, event.Content)
        return nil
    },
}
return bot.Run(ctx)
```

## 이벤트 모델

- `SYSTEM`: `connected`, `subscribed`, `unsubscribed`, `revoked`
- `CHAT`: 작성자, role, 메시지 내용, 이모티콘, message time
- `DONATION`: 후원 유형, 금액, 후원자, 메시지, 이모티콘
- `SUBSCRIPTION`: 구독자, tier, tier name, month

각 이벤트는 `ChatEvent`, `DonationEvent`, `SubscriptionEvent`, `SystemEvent`로 역직렬화됩니다. 저수준 이벤트가 필요하면 `SocketSession.Run`의 `EventHandler`를 사용하고, `DecodeEvent[T]`로 직접 타입을 지정할 수 있습니다.

공식 명세와 대조한 필드는 다음과 같습니다.

| 이벤트 | 필드 |
|---|---|
| `CHAT` | channelId, senderChannelId, chatChannelId, profile(실측 role 포함), userRoleCode(문서 위치), content, emojis, messageTime, eventSentAt(실측) |
| `DONATION` | donationType, channelId, donatorChannelId, donatorNickname, payAmount, donationText, emojis, eventSentAt(실측) |
| `SUBSCRIPTION` | channelId, subscriberChannelId, subscriberNickname, tierNo, tierName, month, eventSentAt(optional) |

공식 문서는 `profile.badges`를 Object 배열로만 정의하고 내부 필드 타입은 공개하지 않습니다. 따라서 `ChatBadge.BadgeNo`는 `any`로 보존하고 알려진 나머지 필드는 optional로 처리해 명세 확장 때문에 전체 이벤트 decoding이 실패하는 상황을 줄였습니다.

실측상 `CHAT.userRoleCode`는 `profile.userRoleCode`에 있으므로 두 위치를 모두 받고 `EffectiveUserRoleCode()`를 제공합니다. `DONATION.payAmount`는 문서상 문자열이지만 실측상 숫자이므로 `DonationAmount`가 두 JSON 타입을 모두 처리합니다. `eventSentAt`은 오프셋 없는 KST 문자열로 보존하며 `ParseEventSentAt`으로 해석할 수 있습니다.

## 운영 팁

- 세션 URL은 장기간 저장하지 말고 연결 직전에 발급합니다.
- 연결이 끊기면 기존 URL을 재사용하지 말고 Session URL 발급부터 다시 시작합니다.
- Access Token과 Refresh Token은 환경변수나 Secret Manager에 보관합니다.
- 연결 종료 시 context를 취소하고 새 세션 URL을 발급해 재연결합니다.
- callback에서 오래 걸리는 작업은 별도 worker로 넘겨 소켓 수신 루프를 막지 않습니다.
