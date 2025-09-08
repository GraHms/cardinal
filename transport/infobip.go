package transport

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/grahms/cardinal/core"
)

// Infobip often posts form fields, but shapes vary by region.
// Common fields seen: "MSISDN", "SESSION_ID", "INPUT" (case varies).
func InfobipFormHandler(eng *core.Engine, opts ...IBOption) http.Handler {
	cfg := ibConfig{
		FieldSessionID: "SESSION_ID",
		FieldMsisdn:    "MSISDN",
		FieldText:      "INPUT",
	}
	for _, o := range opts {
		o(&cfg)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		req := core.Request{
			SessionID: r.FormValue(cfg.FieldSessionID),
			Msisdn:    strings.TrimSpace(r.FormValue(cfg.FieldMsisdn)),
			Text:      strings.TrimSpace(r.FormValue(cfg.FieldText)),
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
	FieldText      string
}
type IBOption func(*ibConfig)

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
