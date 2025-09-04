# Cardinal USSD Framework

*A minimal yet powerful Go framework for building USSD applications.*

---

## 📞 What is USSD?

**USSD (Unstructured Supplementary Service Data)** is the protocol behind interactive telco menus like:

* `*144#` → check airtime,
* `*111#` → buy data bundles,
* or shortcodes used by banks, fintechs, and mobile money platforms.

Its defining strength is **universality**: it works on every mobile phone, without internet.
This makes it the backbone of **financial inclusion, prepaid ecosystems, and mass-market services** across Africa, Asia, and beyond.

---

## ❓ Why Cardinal?

Building USSD services is deceptively complex. Engineers wrestle with:

* **Session management** → USSD sessions are short (30–60 seconds) and volatile.
* **Menu formatting** → every screen is text, with line breaks and numbered options.
* **Aggregator quirks** → fields like `sessionId`, `msisdn`, `text` vary across gateways.
* **Testing** → simulating a multi-step conversation reliably is non-trivial.

**Cardinal** was created to give developers **clarity and precision**, much like a compass:

* **Minimal Core** — `Request`, `Reply`, `Session`, `Engine`.
* **USSD-native semantics** — `SHOW` for screens, `INPUT` for choices.
* **Menu Builder** — auto-formats options, with conventional `0) Back` and `00) Exit`.
* **Parametric Routes** — `/confirm/data/:id` style paths, accessible via `c.Param("id")`.
* **Pluggable Stores** — in-memory for dev, Redis for production.
* **Simulator Test Kit** — BDD-style session testing, step by step.

Cardinal is deliberately **small, predictable, and composable** — so you build only what you need.

---

## 🚀 Getting Started

### Installation

```bash
go get github.com/grahms/cardinal
```

---

## Step-by-Step Examples

Below, we’ll build a simple airtime app — from **basic screens** to **dynamic menus with parameters** — and show **exact outputs** (`CON` or `END`) that the telco aggregator expects.

---

### 1) Basic Menu

**Code**

```go
r := router.New("/home")

r.SHOW("/home", func(c *router.Ctx) core.Reply {
    return menu.New("/home").
        Title("Welcome").
        Opt("Check Balance", "/balance").
        Opt("Buy Airtime", "/amount").
        Exit("Goodbye").
        Prompt(c)
})
r.INPUT("/home", func(c *router.Ctx) core.Reply {
    return menu.New("/home").
        Opt("Check Balance", "/balance").
        Opt("Buy Airtime", "/amount").
        Exit("Goodbye").
        Handle(c)
})

r.SHOW("/balance", func(c *router.Ctx) core.Reply {
    return menu.New("/balance").
        Title("Balance: 123.45 MZN").
        Back("/home").
        Prompt(c)
})
r.INPUT("/balance", func(c *router.Ctx) core.Reply {
    return menu.New("/balance").Back("/home").Handle(c)
})
```

**Flow and Output**

1. **Session starts** (empty `text`) →

   ```
   CON Welcome
   1) Check Balance
   2) Buy Airtime
   00) Goodbye
   ```

2. **User inputs `1`** →

   ```
   CON Balance: 123.45 MZN
   0) Back
   ```

3. **User inputs `0`** →

   ```
   CON Welcome
   1) Check Balance
   2) Buy Airtime
   00) Goodbye
   ```

4. **User inputs `00`** →

   ```
   END Goodbye
   ```

---

### 2) Capturing Free-Form Input

**Code**

```go
r.SHOW("/amount", func(c *router.Ctx) core.Reply {
    return core.CON("Enter amount (MZN):")
})
r.INPUT("/amount", func(c *router.Ctx) core.Reply {
    in := c.In()
    if !digits(in) {
        return core.CON("Invalid amount. Try again:")
    }
    c.Set("amount", atoi(in))
    c.Redirect("/confirm/airtime")
    return core.CON("")
})
```

**Flow**

* **`text=2` (Buy Airtime)** →

  ```
  CON Enter amount (MZN):
  ```
* **`text=abc`** →

  ```
  CON Invalid amount. Try again:
  ```
* **`text=100`** → redirect to confirm:

  ```
  CON Confirm 100 MZN?
  1) Yes
  0) Back
  ```

---

### 3) Confirmation Screen

**Code**

```go
r.SHOW("/confirm/airtime", func(c *router.Ctx) core.Reply {
    amt, _ := c.Get("amount")
    return menu.New("/confirm/airtime").
        Title(fmt.Sprintf("Confirm %v MZN?", amt)).
        End("Yes", "Top-up successful.").
        Back("/home").
        Prompt(c)
})
r.INPUT("/confirm/airtime", func(c *router.Ctx) core.Reply {
    return menu.New("/confirm/airtime").
        End("Yes", "Top-up successful.").
        Back("/home").
        Handle(c)
})
```

**Flow**

* **`text=100`** (from `/amount`) →

  ```
  CON Confirm 100 MZN?
  1) Yes
  0) Back
  ```
* **`text=1`** →

  ```
  END Top-up successful.
  ```

---

### 4) Sub-Paths with Parameters

**Code**

```go
r.SHOW("/products/data", func(c *router.Ctx) core.Reply {
    return menu.New("/products/data").
        Title("Data Bundles").
        Opt("1GB / 100 MZN", "/confirm/data/1").
        Opt("2GB / 180 MZN", "/confirm/data/2").
        Back("/home").
        Prompt(c)
})
r.INPUT("/products/data", func(c *router.Ctx) core.Reply {
    return menu.New("/products/data").
        Opt("1GB / 100 MZN", "/confirm/data/1").
        Opt("2GB / 180 MZN", "/confirm/data/2").
        Back("/home").
        Handle(c)
})

r.SHOW("/confirm/data/:id", func(c *router.Ctx) core.Reply {
    return menu.New("/confirm/data/:id").
        Title("Confirm bundle " + c.Param("id") + "?").
        End("Yes", "Purchase complete").
        Back("/products/data").
        Prompt(c)
})
r.INPUT("/confirm/data/:id", func(c *router.Ctx) core.Reply {
    return menu.New("/confirm/data/:id").
        End("Yes", "Purchase complete").
        Back("/products/data").
        Handle(c)
})
```

**Flow**

1. **`text=2`** →

   ```
   CON Data Bundles
   1) 1GB / 100 MZN
   2) 2GB / 180 MZN
   0) Back
   ```
2. **`text=2`** →

   ```
   CON Confirm bundle 2?
   1) Yes
   0) Back
   ```
3. **`text=1`** →

   ```
   END Purchase complete
   ```

---

## 🧪 Testing with the Simulator

Cardinal ships with a **test kit** to simulate full USSD flows.

**Code**

```go
func TestTopupFlow(t *testing.T) {
    eng := BuildEngineForTests()
    sim := testkit.New(t, eng).Start("+25884xxxxxx")

    sim.Expect("Welcome").
        Send("2").             // Buy Airtime
        Expect("Enter amount").
        Send("100").
        Expect("Confirm 100").
        Send("1").
        ExpectEndsWith("Top-up successful")
}
```

---

## 🗂 Package Structure

```
cardinal/
 ├─ core/        # Engine, Session, Request, Reply
 ├─ router/      # SHOW/INPUT router with param support
 ├─ menu/        # Menu builder abstraction
 ├─ store/       # InMemory, Redis (planned)
 ├─ transport/   # HTTP adapter for aggregator callbacks
 ├─ testkit/     # Simulator for BDD-style tests
 ├─ examples/    # End-to-end flows
 └─ go.mod
```

---

## 📜 Design Philosophy

* **Austere core** — no bloat, just the primitives USSD needs.
* **USSD-native semantics** — screens (`SHOW`) and inputs (`INPUT`) instead of HTTP verbs.
* **Composable** — you can build FSMs, REST-like routes, or plain switch/case.
* **Predictable** — menus always render with consistent numbering and conventions.
* **Testable** — full flows simulated with deterministic assertions.

---

## ⚖️ License

MIT — free to use, fork, and adapt.
