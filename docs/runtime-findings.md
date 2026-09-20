# Session 런타임 실측 반영

이 문서는 [fi-xz의 CHZZK Session API 실측 자료](https://gist.github.com/fi-xz/69ce1f35ca1b2318a2b410c0d5757e0f)를 참고해 Wrapper에 반영한 비공식 관측값을 기록합니다. 공식 문서와 달리 언제든 변할 수 있습니다.

## Socket.IO / Engine.IO

- CHZZK Session은 Socket.IO 2.x / Engine.IO revision 3 방식입니다.
- 발급 URL에는 path가 없지만 Engine.IO transport path는 기본 `/socket.io/`를 사용합니다. Socket.IO namespace는 이와 별개인 루트(`/`)입니다.
- 루트 namespace에 별도 `40` CONNECT packet을 보내면 `auth fail`과 disconnect가 발생할 수 있어 전송하지 않습니다.
- 이벤트 packet의 두 번째 인자는 JSON object가 아니라 JSON 문자열로 한 번 더 인코딩되어 관측됩니다. Decoder는 문자열과 object 양쪽을 지원합니다.
- Engine.IO 3 heartbeat는 Client가 `2` ping을 보내고 Server가 `3` pong으로 응답합니다.
- Handshake의 `pingInterval`, `pingTimeout`을 읽어 heartbeat를 수행하며, 누락된 경우 각각 25초와 60초를 사용합니다. 즉 기본값에서는 연결 후 25초에 Client ping을 보내고 최대 60초 동안 Server pong을 기다리므로, 응답이 전혀 없을 때의 heartbeat 감지 구간은 총 85초입니다.
- Handshake의 `upgrades` 목록은 연결 성공 여부 판단에 사용하지 않습니다.
- 발급된 Session URL은 연결 전에 짧은 시간 안에 만료될 수 있습니다. 끊긴 URL로 자동 재연결하지 말고 새 Session URL을 발급해야 합니다.

## REST 구독

- `sessionKey`는 POST body가 아니라 query parameter로 전송합니다. 현재 구현이 이 방식을 사용합니다.
- WebSocket 연결 후 `SYSTEM/connected`를 받아야 구독 REST API를 호출할 수 있습니다.
- 한 세션당 총 30개 이벤트, User Access Token 기준 사용자당 동시 세션 3개 제한을 고려해야 합니다.
- 필요한 Scope가 없으면 Session URL 발급에서 기대와 달리 500 응답이 관측될 수 있습니다.

## Payload 차이

| 항목 | 공식 문서 | 실측 대응 |
|---|---|---|
| `CHAT.userRoleCode` | 최상위 | `profile.userRoleCode`도 수용 |
| `DONATION.payAmount` | String | 숫자/문자열 모두 수용 |
| `eventSentAt` | 없음 | CHAT/DONATION에 optional 문자열 추가 |
| 이벤트 인자 | JSON object | 이중 인코딩 JSON 문자열도 decoding |

`messageTime`은 UTC 절대 시각으로 변환할 수 있는 epoch milliseconds이며 `ChatEvent.MessageTimeUTC()`를 제공합니다. `eventSentAt`은 이와 별개로 오프셋 표기가 없는 KST이고 소수점 이하 9자리까지 관측되므로 `time.Time` 자동 decoding 대신 원문 문자열로 보존합니다. 필요하면 `api.ParseEventSentAt`을 사용하세요.

구독 이벤트(`SUBSCRIPTION`)는 실측 자료에서도 실제 payload 수신이 검증되지 않았으므로 현재 구조체는 공식 문서 기반이며 `eventSentAt`만 optional로 둡니다.
