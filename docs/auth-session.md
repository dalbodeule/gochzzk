# 인증 및 Session

## 인증 방식

치지직 Open API는 두 인증 방식을 사용합니다.

| 방식 | 헤더 | 사용처 |
|---|---|---|
| Client 인증 | `Client-Id`, `Client-Secret` | 공개 채널/카테고리/라이브 목록, Client Session |
| User 인증 | `Authorization: Bearer <access-token>` | 개인 채널 관리, 채팅, Restriction, User Session |

Base URL은 `https://openapi.chzzk.naver.com`입니다. Authorization Code 동의 URL은 `https://chzzk.naver.com/account-interlock`입니다.

## OAuth 메서드

```go
url := chzzk.AuthorizationURL(clientID, redirectURI, state)
token, err := client.ExchangeCode(ctx, code, state)
token, err = client.RefreshToken(ctx, token.RefreshToken)
err = client.RevokeToken(ctx, token.AccessToken, "access_token")
```

Access Token 유효기간은 공식 문서 기준 1일, Refresh Token은 30일이며 Refresh Token은 일회용입니다. 갱신 응답의 새 Refresh Token을 저장해야 합니다.

## Session REST API

```go
session, err := client.CreateUserSession(ctx)
clientSession, err := client.CreateClientSession(ctx)
userSessions, err := client.ListUserSessions(ctx, "", 20)
clientSessions, err := client.ListClientSessions(ctx, "", 20)
```

세션 발급 결과의 `SessionURL.URL`은 짧은 시간 동안 유효한 Socket.IO 연결 URL입니다. 연결이 끊겼거나 만료되면 새 세션 URL을 발급해야 합니다.

이벤트 Scope는 다음과 같습니다.

- `채팅 메시지 조회`
- `후원 조회`
- `구독 조회`

구독 메서드는 `sessionKey`를 query parameter로 전송합니다.

```go
client.SubscribeChat(ctx, sessionKey)
client.SubscribeDonation(ctx, sessionKey)
client.SubscribeSubscription(ctx, sessionKey)
```

세션당 채팅·후원·구독 이벤트를 합쳐 최대 30개까지 구독할 수 있습니다. Session Socket.IO 연결 상세는 [`chatbot.md`](chatbot.md)를 참고하세요.

유저별 Access Token/Refresh Token 관리와 세 이벤트 자동 구독은 `chatbot.UserChatBot`을 사용하세요. User Session은 Access Token과 같은 사용자 이벤트만 구독하므로 계정마다 별도 인스턴스가 필요합니다.

## 공통 오류

모든 REST 응답은 다음 envelope를 사용합니다.

```json
{"code": 200, "message": null, "content": {}}
```

비정상 HTTP status 또는 envelope code는 `*chzzk.APIError`로 반환됩니다.
