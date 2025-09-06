# Cardinal USSD Framework

*A minimal yet powerful Go framework for building USSD applications in Go.*

---

## 📞 What is USSD?

**USSD (Unstructured Supplementary Service Data)** powers interactive telco menus like:

* `*144#` → check airtime
* `*111#` → buy data bundles
* bank, fintech, and mobile money shortcodes

Its strength is **universality**: works on every mobile phone, without internet.
It remains the backbone of **financial inclusion, prepaid ecosystems, and mass-market services** across Africa, Asia, and beyond.

---

## ❓ Why Cardinal?

Building USSD services is deceptively complex:

* **Session management** → sessions are short (30–60s) and volatile.
* **Menu formatting** → every screen is text with numbered options.
* **Aggregator quirks** → different payloads (`sessionId`, `msisdn`, `text`).
* **Testing** → simulating multi-step conversations is non-trivial.

**Cardinal** gives developers a compass 🧭 for clarity:

* **Minimal Core** — `Request`, `Reply`, `Session`, `Engine`.
* **USSD-native semantics** — `SHOW` for screens, `INPUT` for choices.
* **Menu Builder** — auto-formats options, with `0) Back` and `00) Exit`.
* **Parametric Routes** — e.g. `/confirm/data/:id` via `c.Param("id")`.
* **Pluggable Stores** — in-memory for dev, Redis (planned) for prod.
* **Middleware Chain** — global, per-route, and group-level.
* **Router Groups** — prefix + shared middleware for subtrees.
* **Simulator Test Kit** — BDD-style flow testing.
* **Web Emulator** — try flows in your browser at `/emu`.

Cardinal is **small, predictable, composable** — you only build what you need.

---

## 🚀 Getting Started

### Install

```bash
go get github.com/grahms/cardinal
```

---

## 📋 Step-by-Step Examples

We’ll build a simple airtime app, from **basic menu** to **parameters**.

(… keep your existing *Examples 1–4* sections here unchanged …)

---

## 🔗 Middleware

Middleware wraps handlers — like `net/http` but USSD-native.

```go
r.Use(
    middleware.Recover(),
    middleware.Logging(log.Default()),
    middleware.RateLimitPerMSISDN(5, 5*time.Second),
)
```

Per-route middleware:

```go
r.SHOWWith("/balance", balanceShow,
    middleware.HMAC("super-secret"),
    middleware.TightRouteLimit(),
)
```

Execution order: **globals → group(s) → route → handler**

---

## 🗂 Router Groups

Organize flows with a prefix and shared middlewares:

```go
secure := r.Group("/secure",
    middleware.HMAC("super-secret"),
    middleware.RateLimitPerMSISDN(5, 5*time.Second),
)

secure.SHOW("/balance", balanceShow)
secure.INPUT("/balance", balanceInput)

admin := secure.Group("/admin", middleware.Audit("admin"))
admin.SHOW("/dashboard", adminShow)
```

---

## 🖥 Emulator

Cardinal ships with a lightweight emulator for dev/test.

Mount it next to your engine:

```go
mux := http.NewServeMux()
mux.Handle("/ussd", transport.HTTPHandler(eng))
emulator.Attach(mux, eng) // /emu and /emu/send
```

Run your app:

```bash
go run ./examples/basic
```

Then open: [http://localhost:8080/emu](http://localhost:8080/emu)
You’ll see a phone-like screen to test sessions interactively.

---

## 🧪 Testing with Simulator

(keep your existing testkit example)

---

## 🗂 Package Structure

```
cardinal/
 ├─ core/        # Engine, Session, Request, Reply
 ├─ router/      # SHOW/INPUT router, params, groups
 ├─ menu/        # Menu builder
 ├─ store/       # InMemory, Redis (planned)
 ├─ middleware/  # Recover, Logging, RateLimit, etc.
 ├─ transport/   # HTTP adapter for aggregator callbacks
 ├─ testkit/     # Simulator for BDD tests
 ├─ emulator/    # Web-based emulator
 ├─ examples/    # Example flows (wallet, airtime, etc.)
 └─ go.mod
```

---

## 📜 Design Philosophy

* **Austere core** — no bloat, just USSD primitives.
* **USSD-native semantics** — `SHOW` / `INPUT` instead of HTTP verbs.
* **Composable** — build FSMs, REST-like routes, or simple switch/case.
* **Predictable** — consistent numbering, back/exit.
* **Testable** — simulate flows deterministically.

---

## ⚖️ License

MIT — free to use, fork, adapt.

