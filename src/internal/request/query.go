package request

import (
	"net/url"
	"strconv"
)

// IntParam adds a non-default integer query parameter.
func IntParam(values url.Values, key string, value int) {
	if value != 0 {
		values.Set(key, strconv.Itoa(value))
	}
}
