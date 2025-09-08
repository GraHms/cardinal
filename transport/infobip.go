package transport

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/grahms/cardinal/core"
)

// InfobipFormHandler handles Infobip-style USSD callbacks (form-encoded).
// Defaults:
//
//	SESSION_ID   -> core.Request.SessionID
//	MSISDN       -> core.Request.Msisdn
//	USSD_STRING  -> core.Request.Text         (primary, per docs)
//
// Fallbacks accepted for Text if USSD_STRING is empty: INPUT, text
//
// Response is plain text prefixed with "CON " or "END " as required by Infobip.
func InfobipFormHandler(eng *core.Engine, opts ...IBOption) http.Handler {
	cfg := ibConfig{
		FieldSessionID: "SESSION_ID",
		FieldMsisdn:    "MSISDN",
		FieldText:      "USSD_STRING",
		TextFallbacks:  []string{"INPUT", "text"},
	}
	for _, o := range opts {
		o(&cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// Resolve text with fallbacks: USSD_STRING -> INPUT -> text
		txt := strings.TrimSpace(r.FormValue(cfg.FieldText))
		if txt == "" {
			for _, k := range cfg.TextFallbacks {
				if v := strings.TrimSpace(r.FormValue(k)); v != "" {
					txt = v
					break
				}
			}
		}

		req := core.Request{
			SessionID: r.FormValue(cfg.FieldSessionID),
			Msisdn:    strings.TrimSpace(r.FormValue(cfg.FieldMsisdn)),
			Text:      txt,
			Meta: map[string]string{
				"vendor": "infobip",
				"ip":     clientIP(r),
			},
		}

		rep, _ := eng.Handle(r.Context(), req)

		prefix := "CON "
		if !rep.Continue {
			prefix = "END "
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprint(w, prefix+rep.Message)
	})
}

type ibConfig struct {
	FieldSessionID string
	FieldMsisdn    string
	FieldText      string   // primary key for user input (default: USSD_STRING)
	TextFallbacks  []string // additional keys to try if primary is empty
}

type IBOption func(*ibConfig)

// IBFields overrides the default form field names.
// Example: IBFields("SESSION", "MSISDN", "USSD") or IBFields("", "", "INPUT") to change only text.
func IBFields(sessionID, msisdn, text string) IBOption {
	return func(c *ibConfig) {
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

// IBTextFallbacks overrides or augments the fallback keys for text resolution.
func IBTextFallbacks(keys ...string) IBOption {
	return func(c *ibConfig) {
		if len(keys) > 0 {
			c.TextFallbacks = append([]string{}, keys...)
		}
	}
}
