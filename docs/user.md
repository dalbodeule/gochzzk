# User API

## `GetUser`

로그인한 Access Token 소유자의 채널을 조회합니다.

```go
user, err := client.GetUser(ctx)
fmt.Println(user.ChannelID, user.ChannelName)
```

- HTTP: `GET /open/v1/users/me`
- 인증: User Access Token
- Scope: `유저 정보 조회`

`User`는 `ChannelID`, `ChannelName`을 제공합니다. 다른 채널의 정보는 [`GetChannels`](channel.md)를 사용합니다.
