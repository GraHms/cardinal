package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grahms/cardinal/core"
)

// JSONGenericHandler lets you adapt arbitrary JSON payloads by configuring key names.
// Example:
//
//	mux.Handle("/ussd", transport.JSONGenericHandler(eng,
//	    JSONMap{InSessionID: "sid", InMsisdn: "from", InText: "payload"},
//	    JSONMap{OutTextKey: "msg", OutWrapperKey: "kind", OutWrapperVal: "Response"},
//	))
func JSONGenericHandler(eng *core.Engine, in JSONMap, out JSONMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		req := core.Request{
			SessionID: asString(body[in.InSessionID]),
			Msisdn:    strings.TrimSpace(asString(body[in.InMsisdn])),
			Text:      strings.TrimSpace(asString(body[in.InText])),
			Meta: map[string]string{
				"vendor": "generic-json",
				"ip":     clientIP(r),
			},
		}
		rep, _ := eng.Handle(r.Context(), req)

		outDoc := map[string]any{}
		if out.OutWrapperKey != "" {
			outDoc[out.OutWrapperKey] = out.OutWrapperVal
		}
		if rep.Continue {
			outDoc[out.OutTextKey] = "CON " + rep.Message
		} else {
			outDoc[out.OutTextKey] = "END " + rep.Message
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(outDoc)
	})
}

type JSONMap struct {
	InSessionID   string
	InMsisdn      string
	InText        string
	OutTextKey    string
	OutWrapperKey string
	OutWrapperVal string
}
