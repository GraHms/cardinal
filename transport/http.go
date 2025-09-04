package transport

import (
	"net/http"
	"net/url"

	"github.com/grahms/cardinal/core"
)

// HTTPHandler returns a generic handler that understands common aggregator keys.
// Accepted keys: sessionId, phoneNumber, serviceCode, text (case/alias tolerant)
func HTTPHandler(e *core.Engine) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f := func(k string) string { return pick(r.Form, k) }
		req := core.Request{
			SessionID:   f("sessionId"),
			Msisdn:      f("phoneNumber"),
			ServiceCode: f("serviceCode"),
			Text:        f("text"),
			Meta:        map[string]string{},
		}
		reply, _ := e.Handle(r.Context(), req)
		prefix := "CON "
		if !reply.Continue {
			prefix = "END "
		}
		_, _ = w.Write([]byte(prefix + reply.Message))
	})
}

func pick(v url.Values, key string) string {
	if x := v.Get(key); x != "" {
		return x
	}
	aliases := map[string][]string{
		"sessionId":   {"sessionid", "session_id", "sid"},
		"phoneNumber": {"msisdn", "phone", "from"},
		"serviceCode": {"code", "service", "shortcode"},
		"text":        {"message", "input"},
	}
	for k, al := range aliases {
		if k == key {
			for _, a := range al {
				if x := v.Get(a); x != "" {
					return x
				}
			}
		}
	}
	return ""
}
