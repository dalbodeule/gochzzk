# Category API

## `SearchCategories`

카테고리 이름을 포함 검색합니다.

```go
categories, err := client.SearchCategories(ctx, "게임", 20)
```

- HTTP: `GET /open/v1/categories/search`
- 인증: Client
- `query`: 필수
- `size`: 기본 20, 최대 50

`Category`는 `CategoryType`, `CategoryID`, `CategoryValue`, `PosterImageURL`을 제공합니다. CategoryType은 `GAME`, `SPORTS`, `ETC` 중 하나입니다.
