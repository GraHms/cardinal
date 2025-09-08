package transport

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/grahms/cardinal/core"
)

// AfricaTalkingHandler handles AT's form-encoded USSD POSTs.
// Docs (common shape):
//   - sessionId: string
//   - serviceCode: string (ignored here)
//   - phoneNumber: +<cc><msisdn>
//   - text: accumulated input "1*100*1" or last token (we forward as-is; engine uses last token)
func AfricaTalkingHandler(eng *core.Engine, opts ...ATOption) http.Handler {
	cfg := atConfig{
		FieldSessionID: "sessionId",
		FieldMsisdn:    "phoneNumber",
		FieldText:      "text",
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
				"vendor": "africastalking",
				"ip":     clientIP(r),
			},
		}
		rep, _ := eng.Handle(r.Context(), req)
		prefix := "CON "
		if !rep.Continue {
			prefix = "END "
		}
		// AT expects plain text response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprint(w, prefix+rep.Message)
	})
}

type atConfig struct {
	FieldSessionID string
	FieldMsisdn    string
	FieldText      string
}
type ATOption func(*atConfig)

func ATFields(sessionID, msisdn, text string) ATOption {
	return func(c *atConfig) {
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
