# Live, Chat, Restriction

## Live

```go
page, err := client.GetLives(ctx, 20, "")
next := page.Page.Next
nextPage, err := client.GetLives(ctx, 20, next)
```

- `GET /open/v1/lives`: Client 인증, 현재 라이브 목록
- `GET /open/v1/streams/key`: User 인증, 방송 스트림키
- `GET /open/v1/lives/setting`: User 인증, 방송 설정 조회
- `PATCH /open/v1/lives/setting`: User 인증, 방송 설정 부분 변경

`Live`에는 제목, 현재 시청자 수, 카테고리, 태그, 채널 정보, `LiveThumbnailImageURL`이 포함됩니다.

채널별 현재 상태는 다음 Helper로 확인할 수 있습니다.

```go
status, err := client.GetChannelLiveStatus(ctx, channelID)
if status.IsLive && status.Live != nil {
    fmt.Println(status.Live.LiveTitle, status.ThumbnailURL)
}
```

공식 공개 endpoint는 전체 라이브 목록이므로 Helper는 cursor를 따라 목록을 조회합니다.

## Chat

| 메서드 | Endpoint | Scope |
|---|---|---|
| `SendChat` | `POST /open/v1/chats/send` | 채팅 메시지 쓰기 |
| `RegisterNotice` | `POST /open/v1/chats/notice` | 채팅 공지 쓰기 |
| `GetChatSettings` | `GET /open/v1/chats/settings` | 채팅 설정 조회 |
| `UpdateChatSettings` | `PUT /open/v1/chats/settings` | 채팅 설정 변경 |
| `BlindMessage` | `POST /open/v1/chats/blind-message` | 채팅 메시지 쓰기 |

메시지와 신규 공지는 최대 100자입니다. `LiveSettingPatch`, `ChatSettingsPatch`는 nil 필드를 JSON에서 생략합니다.

`SendChat`과 신규 메시지 방식의 `RegisterNotice`는 요청 전에 Unicode rune 수를 검증합니다. 초과하면 `*ChatMessageLengthError`를 반환하며 네트워크 요청을 보내지 않습니다.

```go
safe := chzzk.TruncateChatMessage(message)
_, err := client.SendChat(ctx, chzzk.SendChatRequest{Message: safe})
```

`TruncateChatMessage`는 한글·영어·이모지 등 UTF-8 문자를 중간에서 자르지 않고 최대 100 rune으로 줄입니다.

## Restriction

```go
err := client.AddRestriction(ctx, targetChannelID)
err = client.RemoveRestriction(ctx, targetChannelID)
page, err := client.GetRestrictions(ctx, 30, "")
```

- `POST /open/v1/restrict-channels`: 활동제한 쓰기
- `DELETE /open/v1/restrict-channels`: 활동제한 쓰기
- `GET /open/v1/restrict-channels`: 활동제한 조회

목록은 `next` cursor를 다음 호출에 전달합니다.
