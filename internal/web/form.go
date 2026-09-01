package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func formStr(r *http.Request, key string) string {
	return strings.TrimSpace(r.FormValue(key))
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func formInt(r *http.Request, key string) int {
	n, _ := strconv.Atoi(formStr(r, key))
	return n
}

func formFloat(r *http.Request, key string) float64 {
	f, _ := strconv.ParseFloat(formStr(r, key), 64)
	return f
}

// formID64Ptr reads an optional foreign-key field: empty string -> nil.
func formID64Ptr(r *http.Request, key string) *int64 {
	v := formStr(r, key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

// formDatePtr parses an optional yyyy-mm-dd date field.
func formDatePtr(r *http.Request, key string) *time.Time {
	v := formStr(r, key)
	if v == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil
	}
	return &t
}

// formDate parses a required yyyy-mm-dd date field, defaulting to today.
func formDate(r *http.Request, key string) time.Time {
	if t := formDatePtr(r, key); t != nil {
		return *t
	}
	return time.Now().Truncate(24 * time.Hour)
}
