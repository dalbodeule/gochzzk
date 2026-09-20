# Channel API

## 공개 채널 조회

```go
channels, err := client.GetChannels(ctx, []string{"channel-id"})
```

- HTTP: `GET /open/v1/channels`
- 인증: Client
- 최대 20개 channel ID

반환되는 `Channel`에는 채널명, 이미지 URL, 팔로워 수, 인증 마크가 포함됩니다.

단일 채널을 간단히 조회하려면 다음 Helper를 사용할 수 있습니다.

```go
channel, err := chzzk.GetChzzkUserInfo(ctx, channelID, clientID, clientSecret)
```

## 관리자·팔로워·구독자

| 메서드 | Endpoint | Scope |
|---|---|---|
| `GetStreamingRoles` | `GET /open/v1/channels/streaming-roles` | 채널 관리자 조회 |
| `GetFollowers` | `GET /open/v1/channels/followers` | 채널 팔로워 조회 |
| `GetSubscribers` | `GET /open/v1/channels/subscribers` | 채널 구독자 조회 |

팔로워/구독자는 `page`가 0부터 시작하고 `size` 기본값은 30, 최대 50입니다. 구독자 정렬은 `RECENT` 또는 `LONGER`입니다.
