package transport

import (
	"net"
	"net/http"
)

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// best-effort client IP (for logs/Meta)
func clientIP(r *http.Request) string {
	// if behind proxies, extend this (X-Forwarded-For, etc.)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
