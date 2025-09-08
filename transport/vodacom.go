package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grahms/cardinal/core"
)

// Example JSON shape (customize via options if your instance differs):
// IN:  {"sessionId":"VDC-123", "msisdn":"+25884...", "userInput":"1"}
// OUT: {"type":"Response", "text":"CON <message>"}
func VodacomHandler(eng *core.Engine, opts ...VodaOption) http.Handler {
	cfg := vodaConfig{
		FieldSessionID: "sessionId",
		FieldMsisdn:    "msisdn",
		FieldText:      "userInput",
		RespTypeKey:    "type",
		RespTextKey:    "text",
		RespTypeValue:  "Response",
	}
	for _, o := range opts {
		o(&cfg)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		req := core.Request{
			SessionID: asString(in[cfg.FieldSessionID]),
			Msisdn:    strings.TrimSpace(asString(in[cfg.FieldMsisdn])),
			Text:      strings.TrimSpace(asString(in[cfg.FieldText])),
			Meta: map[string]string{
				"vendor": "vodacom",
				"ip":     clientIP(r),
			},
		}
		rep, _ := eng.Handle(r.Context(), req)

		out := map[string]any{
			cfg.RespTypeKey: cfg.RespTypeValue,
		}
		if rep.Continue {
			out[cfg.RespTextKey] = "CON " + rep.Message
		} else {
			out[cfg.RespTextKey] = "END " + rep.Message
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	})
}

type vodaConfig struct {
	FieldSessionID string
	FieldMsisdn    string
	FieldText      string

	RespTypeKey   string
	RespTextKey   string
	RespTypeValue string
}
type VodaOption func(*vodaConfig)

func VodaFields(sessionID, msisdn, text string) VodaOption {
	return func(c *vodaConfig) {
		if sessionID != "" {
			c.FieldSessionID = sessionID
		}
		if msisdn != "" {
			c.FieldMsisdn = msisdn
		}
		if text != "" {
			c.FieldText = text
		}
	}
}
func VodaRespKeys(typeKey, textKey, typeVal string) VodaOption {
	return func(c *vodaConfig) {
		if typeKey != "" {
			c.RespTypeKey = typeKey
		}
		if textKey != "" {
			c.RespTextKey = textKey
		}
		if typeVal != "" {
			c.RespTypeValue = typeVal
		}
	}
}
