package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/emulator"
	"github.com/grahms/cardinal/menu"
	"github.com/grahms/cardinal/router"
	"github.com/grahms/cardinal/store"
	"github.com/grahms/cardinal/transport"
)

/*
Mobile Wallet demo:
- 3 accounts (MSISDNs) linked to the same user
- View balances
- Transfer: (auto source from caller) -> destination -> amount -> confirm
- Mini-statement (last N entries)

NOTES:
- This is an in-memory mock; replace WalletSvc with your real services.
- Keep handlers fast (<800ms).
*/

func main() {
	// ----- Mock data/service
	svc := NewWalletSvc(
		[]string{"+258840000001", "+258840000002", "+258840000003"},
		[]int64{50000, 75000, 25000}, // minor units (centavos): 500.00, 750.00, 250.00
	)

	// ----- Router and middlewares
	r := router.New("/home")
	// Example: add your own middlewares (Recover/Logging/RateLimit/etc.) if desired
	// r.Use(middleware.Recover(), middleware.Logging(log.Default()))

	// ----- Home
	r.SHOW("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Title("Mobile Wallet").
			Opt("My Wallet", "/wallet").
			Opt("Mini-statement", "/wallet/history").
			Exit("Goodbye.").
			Prompt(c)
	})
	r.INPUT("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Opt("My Wallet", "/wallet").
			Opt("Mini-statement", "/wallet/history").
			Exit("Goodbye.").
			Handle(c)
	})

	// ----- Wallet section
	r.SHOW("/wallet", func(c *router.Ctx) core.Reply {
		return menu.New("/wallet").
			Title("Wallet").
			Opt("Balances", "/wallet/balances").
			Opt("Transfer", "/wallet/transfer/start"). // smart start uses caller MSISDN
			Back("/home").
			Prompt(c)
	})
	r.INPUT("/wallet", func(c *router.Ctx) core.Reply {
		return menu.New("/wallet").
			Opt("Balances", "/wallet/balances").
			Opt("Transfer", "/wallet/transfer/start").
			Back("/home").
			Handle(c)
	})

	// ----- Balances (shows 3 MSISDNs; tag caller as (you))
	r.SHOW("/wallet/balances", func(c *router.Ctx) core.Reply {
		lines := []string{"Balances"}
		for _, acc := range svc.Accounts() {
			tag := ""
			if acc == strings.TrimSpace(c.Req.Msisdn) {
				tag = " (you)"
			}
			lines = append(lines, fmt.Sprintf("%s: %s%s", acc, fmtAmount(svc.Balance(acc)), tag))
		}
		return core.CON(strings.Join(lines, "\n") + "\n0) Back")
	})
	r.INPUT("/wallet/balances", func(c *router.Ctx) core.Reply {
		// any input -> back
		return menu.New("/wallet/balances").Back("/wallet").Handle(c)
	})

	// ----- Transfer flow: smart start -> destination -> amount -> confirm

	// Smart start: if caller MSISDN is one of the accounts, use it as source; else fall back.
	r.SHOW("/wallet/transfer/start", func(c *router.Ctx) core.Reply {
		caller := strings.TrimSpace(c.Req.Msisdn)
		if svc.Has(caller) {
			c.Set("_xfer_src", caller)
			c.Redirect("/wallet/transfer/dest/" + caller)
			return core.CON("") // engine will SHOW next
		}
		c.Redirect("/wallet/transfer/source")
		return core.CON("")
	})
	r.INPUT("/wallet/transfer/start", func(c *router.Ctx) core.Reply {
		caller := strings.TrimSpace(c.Req.Msisdn)
		if svc.Has(caller) {
			c.Set("_xfer_src", caller)
			c.Redirect("/wallet/transfer/dest/" + caller)
			return core.CON("")
		}
		c.Redirect("/wallet/transfer/source")
		return core.CON("")
	})

	// Manual source picker (used only if caller MSISDN isn't an owned account)
	r.SHOW("/wallet/transfer/source", func(c *router.Ctx) core.Reply {
		b := menu.New("/wallet/transfer/source").Title("Choose source")
		for _, acc := range svc.Accounts() {
			b.Opt(acc, "/wallet/transfer/dest/"+acc)
		}
		b.Back("/wallet")
		return b.Prompt(c)
	})
	r.INPUT("/wallet/transfer/source", func(c *router.Ctx) core.Reply {
		b := menu.New("/wallet/transfer/source")
		for _, acc := range svc.Accounts() {
			b.Opt(acc, "/wallet/transfer/dest/"+acc)
		}
		b.Back("/wallet")
		return b.Handle(c)
	})

	// Choose destination (exclude source)
	r.SHOW("/wallet/transfer/dest/:src", func(c *router.Ctx) core.Reply {
		src := c.Param("src")
		b := menu.New("/wallet/transfer/dest/:src").Title("Destination")
		for _, acc := range svc.Accounts() {
			if acc == src {
				continue
			}
			b.Opt(acc, "/wallet/transfer/amount/"+src+"/"+acc)
		}
		b.Back("/wallet/transfer/source")
		return b.Prompt(c)
	})
	r.INPUT("/wallet/transfer/dest/:src", func(c *router.Ctx) core.Reply {
		src := c.Param("src")
		b := menu.New("/wallet/transfer/dest/:src")
		for _, acc := range svc.Accounts() {
			if acc == src {
				continue
			}
			b.Opt(acc, "/wallet/transfer/amount/"+src+"/"+acc)
		}
		b.Back("/wallet/transfer/source")
		return b.Handle(c)
	})

	// Enter amount (free form)
	r.SHOW("/wallet/transfer/amount/:src/:dst", func(c *router.Ctx) core.Reply {
		c.Set("_xfer_src", c.Param("src"))
		c.Set("_xfer_dst", c.Param("dst"))
		return core.CON("Enter amount (MZN):")
	})
	r.INPUT("/wallet/transfer/amount/:src/:dst", func(c *router.Ctx) core.Reply {
		in := strings.TrimSpace(c.In())
		if !digits(in) || in == "0" {
			return core.CON("Invalid amount. Try again:")
		}
		// store minor units
		amt := parseMajorToMinor(in) // "100" -> 10000
		c.Set("_xfer_amt", amt)
		c.Redirect("/wallet/transfer/confirm")
		return core.CON("")
	})

	// Confirm transfer
	r.SHOW("/wallet/transfer/confirm", func(c *router.Ctx) core.Reply {
		src := mustString(c, "_xfer_src")
		dst := mustString(c, "_xfer_dst")
		amt := mustInt64(c, "_xfer_amt")
		title := fmt.Sprintf("Transfer %s\nFrom %s\nTo   %s ?", fmtMinor(amt), src, dst)
		return menu.New("/wallet/transfer/confirm").
			Title(title).
			End("Confirm", "Processing...").
			Back("/wallet/transfer/source").
			Prompt(c)
	})
	r.INPUT("/wallet/transfer/confirm", func(c *router.Ctx) core.Reply {
		// Idempotent confirm: only apply once per session.
		if applied, _ := c.Get("_xfer_done"); applied == true {
			return core.END("Already processed.")
		}
		src := mustString(c, "_xfer_src")
		dst := mustString(c, "_xfer_dst")
		amt := mustInt64(c, "_xfer_amt")

		// Basic checks
		if src == "" || dst == "" || amt <= 0 {
			return core.END("Invalid transfer request.")
		}
		// Apply
		err := svc.Transfer(src, dst, amt, "USSD")
		if err != nil {
			return core.END("Transfer failed: " + err.Error())
		}
		c.Set("_xfer_done", true)
		return core.END("Transfer complete.")
	})

	// ----- Mini-statement
	r.SHOW("/wallet/history", func(c *router.Ctx) core.Reply {
		// Merge last 5 entries from all accounts and show the newest first
		entries := svc.LastEntries(5)
		if len(entries) == 0 {
			return core.CON("No recent activity.\n0) Back")
		}
		lines := []string{"Recent activity"}
		for _, e := range entries {
			sign := "+"
			if e.Type == "DEBIT" {
				sign = "-"
			}
			lines = append(lines,
				fmt.Sprintf("%s %s %s → %s (%s)",
					e.Time.Format("15:04"),
					sign+fmtMinor(e.Amount),
					e.From, e.To, e.Ref,
				),
			)
		}
		return core.CON(strings.Join(lines, "\n") + "\n0) Back")
	})
	r.INPUT("/wallet/history", func(c *router.Ctx) core.Reply {
		return menu.New("/wallet/history").Back("/home").Handle(c)
	})

	// ----- Wire engine + emulator
	st := store.NewInMemoryStore(60 * time.Second)
	eng := core.New(r.Mount(), core.Config{Store: st, SessionTTL: 60 * time.Second})

	mux := http.NewServeMux()
	mux.Handle("/ussd", transport.HTTPHandler(eng))
	emulator.Attach(mux, eng) // /emu and /emu/send

	log.Println("Wallet demo on :8080 — USSD: /ussd  | Emulator: /emu")
	_ = http.ListenAndServe(":8080", mux)
}

/* ---------------- Wallet mock service ---------------- */

type Tx struct {
	Time   time.Time
	From   string
	To     string
	Amount int64  // minor units
	Type   string // DEBIT/CREDIT
	Ref    string
}
type WalletSvc struct {
	accts []string
	bal   map[string]int64
	logs  []Tx
}

func NewWalletSvc(accts []string, start []int64) *WalletSvc {
	b := make(map[string]int64)
	for i, a := range accts {
		b[a] = start[i]
	}
	return &WalletSvc{accts: accts, bal: b, logs: []Tx{}}
}

func (w *WalletSvc) Accounts() []string {
	cp := append([]string{}, w.accts...)
	return cp
}
func (w *WalletSvc) Balance(acc string) int64 {
	return w.bal[acc]
}

// Has reports whether msisdn belongs to this wallet.
func (w *WalletSvc) Has(msisdn string) bool {
	_, ok := w.bal[msisdn]
	return ok
}
func (w *WalletSvc) Transfer(from, to string, amt int64, ref string) error {
	if from == to {
		return fmt.Errorf("same account")
	}
	if _, ok := w.bal[from]; !ok {
		return fmt.Errorf("unknown source")
	}
	if _, ok := w.bal[to]; !ok {
		return fmt.Errorf("unknown destination")
	}
	if amt <= 0 {
		return fmt.Errorf("invalid amount")
	}
	if w.bal[from] < amt {
		return fmt.Errorf("insufficient funds")
	}
	// apply
	w.bal[from] -= amt
	w.bal[to] += amt
	now := time.Now()
	w.logs = append(w.logs,
		Tx{Time: now, From: from, To: to, Amount: amt, Type: "DEBIT", Ref: ref},
		Tx{Time: now, From: from, To: to, Amount: amt, Type: "CREDIT", Ref: ref},
	)
	return nil
}

// LastEntries merges newest-first across all logs, capped to n
func (w *WalletSvc) LastEntries(n int) []Tx {
	cp := append([]Tx{}, w.logs...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Time.After(cp[j].Time) })
	if len(cp) > n {
		cp = cp[:n]
	}
	return cp
}

/* ---------------- helpers ---------------- */

func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parse "100" major to 10000 minor (two decimals)
func parseMajorToMinor(s string) int64 {
	// only integers in demo; adapt to decimals if needed
	n := int64(0)
	for _, r := range s {
		n = n*10 + int64(r-'0')
	}
	return n * 100
}
func fmtAmount(minor int64) string { return fmt.Sprintf("%d.%02d MZN", minor/100, minor%100) }
func fmtMinor(minor int64) string  { return fmt.Sprintf("%d.%02d", minor/100, minor%100) }

func mustString(c *router.Ctx, k string) string {
	if v, ok := c.Get(k); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
func mustInt64(c *router.Ctx, k string) int64 {
	if v, ok := c.Get(k); ok {
		switch x := v.(type) {
		case int64:
			return x
		case int:
			return int64(x)
		case string:
			if n, err := strconv.ParseInt(x, 10, 64); err == nil {
				return n
			}
		}
	}
	return 0
}
