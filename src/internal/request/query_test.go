package request

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestIntParam(t *testing.T) {
	values := url.Values{}
	IntParam(values, "size", 20)
	IntParam(values, "page", 0)
	if values.Get("size") != "20" || values.Get("page") != "" {
		t.Fatalf("unexpected query values: %s", values.Encode())
	}
}

func TestNewHTTPClientIsBoundedAndRejectsRedirects(t *testing.T) {
	client := NewHTTPClient()
	if client.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %s", client.Timeout)
	}
	if err := client.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("unexpected redirect policy: %v", err)
	}
}
