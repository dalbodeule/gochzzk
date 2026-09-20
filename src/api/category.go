package chzzk

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dalbodeule/gochzzk/src/internal/request"
)

// SearchCategories searches categories whose name contains query.
func (c *Client) SearchCategories(ctx context.Context, query string, size int) ([]Category, error) {
	q := url.Values{"query": []string{query}}
	request.IntParam(q, "size", size)
	var v []Category
	err := c.do(ctx, http.MethodGet, "/open/v1/categories/search", q, nil, authClient, &v)
	return v, err
}
