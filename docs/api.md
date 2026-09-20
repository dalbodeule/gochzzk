# 치지직 Open API 문서

이 문서는 [치지직 공식 API 문서](https://chzzk.gitbook.io/chzzk)를 기준으로 작성한 Wrapper 사용 안내입니다.

상세 문서: [인증/세션](auth-session.md) · [User](user.md) · [Channel](channel.md) · [Category](category.md) · [Live/Chat/Restriction](live-chat.md) · [Chatbot](chatbot.md)

운영·동시성·보안 점검은 [security.md](security.md)를 참고하세요.

공식 Open API에 포함되지 않는 라이브 상태와 profile-card 조회는 별도 [unofficial.md](unofficial.md)에 분리했습니다.

## 공통 규격

- Base URL: `https://openapi.chzzk.naver.com`
- 성공 응답: `{ "code": 200, "message": null, "content": ... }`
- 실패 응답: `{ "code": integer, "message": string }`
- Client 인증 헤더: `Client-Id`, `Client-Secret`
- Access Token 헤더: `Authorization: Bearer <access-token>`
- 주요 HTTP 오류: `400` 파라미터 오류, `401` 인증 실패, `403` 권한 없음, `404` 없음, `429` quota 초과, `500` 서버 오류

Wrapper는 비정상 HTTP 응답 또는 envelope의 비정상 `code`를 `*chzzk.APIError`로 반환합니다.

## OAuth 인증 흐름

1. `chzzk.AuthorizationURL(clientID, redirectURI, state)`로 동의 URL을 생성합니다.
2. 콜백의 `code`와 검증된 `state`를 `ExchangeCode`에 전달합니다.
3. Access Token은 1일, Refresh Token은 30일 동안 유효합니다. Refresh Token은 일회용이므로 `RefreshToken`이 반환한 새 Refresh Token을 저장해야 합니다.
4. 로그아웃 또는 철회 시 `RevokeToken`을 호출합니다.

Authorization Code URL은 `https://chzzk.naver.com/account-interlock`이며, Token/Revoke API는 Open API 도메인의 `/auth/v1/token` 및 `/auth/v1/token/revoke`입니다.

## API와 필요한 Scope

| Wrapper 메서드 | HTTP API | 인증 | 필요한 Scope |
|---|---|---|---|
| `GetUser` | `GET /open/v1/users/me` | Access Token | 유저 정보 조회 |
| `GetChannels` | `GET /open/v1/channels` | Client | 없음(애플리케이션 등록 및 Client 인증) |
| `GetStreamingRoles` | `GET /open/v1/channels/streaming-roles` | Access Token | 채널 관리자 조회 |
| `GetFollowers` | `GET /open/v1/channels/followers` | Access Token | 채널 팔로워 조회 |
| `GetSubscribers` | `GET /open/v1/channels/subscribers` | Access Token | 채널 구독자 조회 |
| `SearchCategories` | `GET /open/v1/categories/search` | Client | 없음(애플리케이션 등록 및 Client 인증) |
| `GetLives` | `GET /open/v1/lives` | Client | 없음(애플리케이션 등록 및 Client 인증) |
| `GetStreamKey` | `GET /open/v1/streams/key` | Access Token | 방송 스트림키 조회 |
| `GetLiveSetting` | `GET /open/v1/lives/setting` | Access Token | 방송 설정 조회 |
| `UpdateLiveSetting` | `PATCH /open/v1/lives/setting` | Access Token | 방송 설정 변경 |
| `SendChat` | `POST /open/v1/chats/send` | Access Token | 채팅 메시지 쓰기 |
| `RegisterNotice` | `POST /open/v1/chats/notice` | Access Token | 채팅 공지 쓰기 |
| `GetChatSettings` | `GET /open/v1/chats/settings` | Access Token | 채팅 설정 조회 |
| `UpdateChatSettings` | `PUT /open/v1/chats/settings` | Access Token | 채팅 설정 변경 |
| `BlindMessage` | `POST /open/v1/chats/blind-message` | Access Token | 채팅 메시지 쓰기 |
| `AddRestriction` / `RemoveRestriction` | `POST/DELETE /open/v1/restrict-channels` | Access Token | 활동제한 쓰기 |
| `GetRestrictions` | `GET /open/v1/restrict-channels` | Access Token | 활동제한 조회 |
| `ExchangeCode` / `RefreshToken` / `RevokeToken` | `POST /auth/v1/token` / `POST /auth/v1/token/revoke` | Client body credentials | 해당 없음 |

Session API의 이벤트 Scope는 `채팅 메시지 조회`, `후원 조회`, `구독 조회`입니다. `CreateClientSession`/`ListClientSessions`는 Client 인증을, `CreateUserSession`/`ListUserSessions` 및 이벤트 구독은 Access Token을 사용합니다.

| Wrapper 메서드 | HTTP API | 용도 |
|---|---|---|
| `CreateClientSession` / `CreateUserSession` | `GET /open/v1/sessions/auth/client`, `/auth` | Socket.IO 연결 URL 발급 |
| `ListClientSessions` / `ListUserSessions` | `GET /open/v1/sessions/client`, `/sessions` | 세션 이력 및 구독 상태 조회 |
| `SubscribeChat` / `UnsubscribeChat` | `POST /open/v1/sessions/events/.../chat` | 채팅 이벤트 구독/취소 |
| `SubscribeDonation` / `UnsubscribeDonation` | `POST /open/v1/sessions/events/.../donation` | 후원 이벤트 구독/취소 |
| `SubscribeSubscription` / `UnsubscribeSubscription` | `POST /open/v1/sessions/events/.../subscription` | 구독 이벤트 구독/취소 |

문서상 채팅 API에는 `채팅 메시지 조회` Scope도 정의되어 있지만, 현재 공개된 이 Wrapper 메서드 목록에는 메시지 조회 엔드포인트가 없습니다.

## 페이지네이션 및 제약

- `GetChannels`: 한 번에 최대 20개 채널 ID.
- `GetFollowers`, `GetSubscribers`: `page`는 0부터 시작하며 `size`는 1~50입니다. 기본값은 각각 0, 30입니다.
- `GetSubscribers`: `sort`는 `RECENT` 또는 `LONGER`입니다.
- `SearchCategories`: `query`는 필수이며 `size`는 1~50입니다.
- `GetLives`: `size`는 1~20이며, 다음 요청에는 응답의 `Page.Next`를 전달합니다.
- 채팅 메시지와 신규 공지는 최대 100자입니다.
- 라이브/채팅 설정 Patch 구조체는 nil 필드를 생략하므로 특정 값만 변경할 수 있습니다. 빈 태그 목록은 태그 제거를 의미합니다.
- `RefreshToken`은 반환된 새 Refresh Token으로 저장소를 갱신해야 합니다.
- Restriction 목록은 `next` cursor를 다음 `GetRestrictions` 호출에 전달합니다.

## Convenience helper

```go
channel, err := chzzk.GetChzzkUserInfo(ctx, channelID, clientID, clientSecret)
if err != nil { return err }
if channel != nil { fmt.Println(channel.ChannelName, channel.FollowerCount) }
```

`GetChzzkUserInfo`는 공식 `GET /open/v1/channels?channelIds=...`를 호출해 단일 채널을 반환합니다. 채널이 없으면 `nil, nil`을 반환합니다. 여러 채널이 필요하거나 원본 목록이 필요하면 `Client.GetChannels`를 사용하세요.

## Socket.IO 챗봇 구조

치지직 문서는 Socket.IO-client `1.0.0+`와 `2.0.3`까지를 지원한다고 안내하며, WebSocket transport를 사용합니다. 이 Wrapper의 `NewSocketSession`은 API가 반환한 URL에 연결하고 `SYSTEM`, `CHAT`, `DONATION`, `SUBSCRIPTION`을 `SessionEvent`로 정규화합니다. `ChatBot`은 이를 `ChatEvent`, `DonationEvent`, `SubscriptionEvent` callback으로 변환합니다.

REST 세션 API와 소켓 연결은 분리되어 있습니다. 먼저 세션 URL을 발급하고 소켓의 `SYSTEM/connected` 이벤트에서 `sessionKey`를 얻은 다음, 해당 키로 이벤트 구독 API를 호출해야 합니다. 세션당 이벤트는 최대 30개입니다.

## 예시

```go
ctx := context.Background()
client := chzzk.New(
    chzzk.WithClientCredentials(os.Getenv("CHZZK_CLIENT_ID"), os.Getenv("CHZZK_CLIENT_SECRET")),
    chzzk.WithAccessToken(os.Getenv("CHZZK_ACCESS_TOKEN")),
)

page, err := client.GetLives(ctx, 20, "")
if err != nil { return err }
for _, live := range page.Data { fmt.Println(live.ChannelName, live.LiveTitle) }
```

## 공식 문서

- [소개](https://chzzk.gitbook.io/chzzk/introduction/about)
- [Authorization](https://chzzk.gitbook.io/chzzk/chzzk-api/authorization)
- [참고사항 및 공통 인증](https://chzzk.gitbook.io/chzzk/chzzk-api/tips)
- [User](https://chzzk.gitbook.io/chzzk/chzzk-api/user)
- [Channel](https://chzzk.gitbook.io/chzzk/chzzk-api/channel)
- [Category](https://chzzk.gitbook.io/chzzk/chzzk-api/category)
- [Live](https://chzzk.gitbook.io/chzzk/chzzk-api/live)
- [Chat](https://chzzk.gitbook.io/chzzk/chzzk-api/chat)
- [Session](https://chzzk.gitbook.io/chzzk/chzzk-api/session)
- [Restriction](https://chzzk.gitbook.io/chzzk/chzzk-api/restriction)
