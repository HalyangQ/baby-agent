package shared

import "encoding/json"

func Ptr[T any](v T) *T {
	return &v
}

func JsonString(v any) string {
	msg, _ := json.Marshal(v)
	return string(msg)
}
