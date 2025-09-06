package emulator

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grahms/cardinal/core"
)

// Attach mounts the emulator UI and API onto your mux.
// GET  /emu        -> HTML UI
// POST /emu/send   -> {sessionId, msisdn, text, append} -> {raw, continue, message}
func Attach(mux *http.ServeMux, eng *core.Engine) {
	mux.HandleFunc("/emu", func(w http.ResponseWriter, r *http.Request) {
		_ = pageTmpl.Execute(w, map[string]any{
			"RandSID":    "sess-" + strconv.FormatInt(time.Now().UnixNano(), 36),
			"RandMSISDN": "+25884" + strconv.Itoa(100000+rand.Intn(899999)),
		})
	})

	mux.HandleFunc("/emu/send", func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			SessionID string `json:"sessionId"`
			Msisdn    string `json:"msisdn"`
			Text      string `json:"text"`
			Append    bool   `json:"append"` // if true, emulate gateways that accumulate: "1*100*1"
		}
		type resp struct {
			Raw      string `json:"raw"`      // "CON ..." or "END ..."
			Continue bool   `json:"continue"` // true if CON
			Message  string `json:"message"`  // body after prefix
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body reqBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		// Optional: emulate accumulated text
		text := strings.TrimSpace(body.Text)

		rep, _ := eng.Handle(r.Context(), core.Request{
			SessionID: body.SessionID,
			Msisdn:    body.Msisdn,
			Text:      text,
			Meta:      map[string]string{"emu": "true"},
		})
		prefix := "CON "
		if !rep.Continue {
			prefix = "END "
		}

		_ = json.NewEncoder(w).Encode(resp{
			Raw:      prefix + rep.Message,
			Continue: rep.Continue,
			Message:  rep.Message,
		})
	})
}

var pageTmpl = template.Must(template.New("emu").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>Cardinal — USSD Emulator</title>
<style>
  :root { --bg:#0f172a; --fg:#e2e8f0; --mut:#94a3b8; --acc:#22c55e; --err:#ef4444; --card:#111827; }
  body{background:var(--bg);color:var(--fg);font-family:ui-sans-serif,system-ui,-apple-system; margin:0;padding:24px;}
  h1{font-size:18px;margin:0 0 12px}
  .grid{display:grid;grid-template-columns:320px 1fr;gap:16px;align-items:start}
  .card{background:var(--card);border:1px solid #1f2937;border-radius:14px;padding:16px}
  label{display:block;font-size:12px;color:var(--mut);margin-bottom:6px}
  input[type=text]{width:100%;padding:10px;border-radius:10px;border:1px solid #1f2937;background:#0b1220;color:var(--fg)}
  button{padding:10px 14px;border-radius:10px;border:0;background:var(--acc);color:#052e15;font-weight:700;cursor:pointer}
  button:disabled{opacity:.5;cursor:not-allowed}
  .row{display:flex;gap:8px;align-items:center}
  .stack{display:grid;gap:10px}
  .kbd{font:12px/1.6 ui-monospace, SFMono-Regular, Menlo; background:#0b1220;color:#a3e635;padding:6px 8px;border-radius:8px}
  .log{max-height:70vh;overflow:auto;font:12px/1.5 ui-monospace, Menlo, Consolas;background:#0b1220;border:1px solid #1f2937;border-radius:12px;padding:12px;white-space:pre-wrap}
  .line{margin:0 0 8px}
  .pill{display:inline-block;padding:2px 8px;border-radius:9999px;font-size:11px;margin-left:6px}
  .pill.con{background:#064e3b;color:#a7f3d0}
  .pill.end{background:#4c0519;color:#fecdd3}
  .small{font-size:12px;color:var(--mut)}
  .hint{color:var(--mut);font-size:12px}
</style>
</head>
<body>
  <h1>Cardinal — USSD Emulator <span class="small">(test mode)</span></h1>
  <div class="grid">
    <div class="card">
      <div class="stack">
        <div>
          <label>Session ID</label>
          <input id="sid" type="text" value="{{.RandSID}}" />
        </div>
        <div>
          <label>MSISDN</label>
          <input id="msisdn" type="text" value="{{.RandMSISDN}}" />
        </div>
        <div>
          <label>Input <span class="hint">(enter the next choice or value — e.g., <span class="kbd">1</span>, <span class="kbd">100</span>, <span class="kbd">00</span>)</span></label>
          <input id="text" type="text" placeholder="e.g. 1 or 100" />
        </div>
        <div class="row">
          <button id="start">Start Session</button>
          <button id="send" disabled>Send</button>
          <button id="reset">Reset</button>
        </div>
        <div class="hint">Tip: press <span class="kbd">Enter</span> in the input to send.</div>
      </div>
    </div>

    <div class="card">
      <div class="row" style="justify-content:space-between;align-items:baseline">
        <div>
          <div class="small">Emulator Log</div>
        </div>
        <div class="small">Shows exact raw replies (<span class="kbd">CON</span>/<span class="kbd">END</span>).</div>
      </div>
      <div id="log" class="log"></div>
    </div>
  </div>

<script>
const $ = id => document.getElementById(id);
const sid = $('sid'), msisdn = $('msisdn'), text = $('text');
const btnStart = $('start'), btnSend = $('send'), btnReset = $('reset');
const log = $('log');

function row(raw) {
  const isCon = raw.startsWith('CON ');
  const pill = isCon ? '<span class="pill con">CON</span>' : '<span class="pill end">END</span>';
  return '<div class="line">' + pill + ' ' + raw.replace(/^CON\\s|^END\\s/,'') + '</div>';
}
function append(raw) { log.insertAdjacentHTML('beforeend', row(raw)); log.scrollTop = log.scrollHeight; }
function setSending(v){ btnStart.disabled=v; btnSend.disabled=!v; }

async function call(textVal) {
  const body = { sessionId: sid.value.trim(), msisdn: msisdn.value.trim(), text: textVal };
  const res = await fetch('/emu/send', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body) });
  const j = await res.json();
  append(j.raw);
  if (!j.continue) { setSending(false); }
}

btnStart.addEventListener('click', async () => {
  log.innerHTML=''; setSending(true);
  await call('');  // first call with empty text shows start screen
  text.focus();
});

btnSend.addEventListener('click', async () => {
  if (!text.value.trim()) return;
  await call(text.value.trim());
  text.value='';
  text.focus();
});

text.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') { e.preventDefault(); btnSend.click(); }
});

btnReset.addEventListener('click', () => {
  log.innerHTML=''; text.value=''; setSending(false);
});

</script>
</body></html>`))
