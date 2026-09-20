# 비공식 CHZZK Web API

`src/unofficial`은 공식 Open API에 없는 채널 단일 라이브 상태와 profile-card 조회를 제공합니다.

> 이 endpoint들은 치지직 웹 서비스 내부용이며 공식 Open API가 아닙니다. 사전 공지 없이 URL·필드·인증·rate limit이 변경되거나 차단될 수 있습니다. 결제, 권한 판정, 영구 데이터처럼 중요한 로직의 유일한 근거로 사용하지 마세요.

## 특정 채널 라이브 상태와 chatChannelId

```go
web := unofficial.New(unofficial.WithUserAgent("my-bot/1.0"))
status, err := web.GetLiveStatus(ctx, channelID)
if err != nil { return err }
if status != nil && status.Status == "OPEN" {
    fmt.Println(status.LiveTitle, status.ChatChannelID)
}

chatChannelID, err := web.GetChatChannelID(ctx, channelID)
```

호출 endpoint:

```text
GET https://api.chzzk.naver.com/polling/v3/channels/{channelId}/live-status
    ?includePlayerRecommendContent=false
```

`LiveStatus`는 제목, 상태, 현재/누적 시청자 수, 성인·한국 제한 여부, 시작/종료 시각, clip 활성화 여부와 chatChannelId를 제공합니다.

## 특정 채팅 채널에서 사용자의 팔로우 날짜

```go
profile, err := web.GetFollowProfile(ctx, chatChannelID, userID)
followDate, err := web.GetFollowDate(ctx, chatChannelID, userID)
if followDate != nil { fmt.Println(*followDate) }
```

호출 endpoint:

```text
GET https://comm-api.game.naver.com/nng_main/v1/chats/{chatChannelId}/users/{userId}/profile-card
    ?chatType=STREAMING
```

라이브 상태에서 얻은 `chatChannelId`와 조회 대상 사용자의 channel/user ID가 필요합니다. 팔로우 정보가 없거나 응답 content가 `null`이면 `GetFollowDate`는 `nil, nil`을 반환합니다.

비공식 endpoint에 token이나 Client Secret을 보내지 않으며, 식별자는 URL path escaping 후 전송합니다. 응답 크기는 4 MiB로 제한합니다.
