package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/emulator"
	"github.com/grahms/cardinal/menu"
	"github.com/grahms/cardinal/router"
	"github.com/grahms/cardinal/store"
	"github.com/grahms/cardinal/transport"
)

func main() {
	// --- Sample catalog (12 bundles)
	bundles := []string{
		"Daily 100MB / 10 MZN",
		"Daily 500MB / 20 MZN",
		"Daily 1GB / 30 MZN",
		"Weekly 1GB / 50 MZN",
		"Weekly 2GB / 90 MZN",
		"Weekly 5GB / 200 MZN",
		"Monthly 1GB / 100 MZN",
		"Monthly 5GB / 400 MZN",
		"Monthly 10GB / 700 MZN",
		"Night 1GB / 15 MZN",
		"Night 3GB / 40 MZN",
		"Night 10GB / 100 MZN",
	}

	// --- Router
	r := router.New("/home")

	// Home: entry point → leads to the paginated bundles list
	r.SHOW("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Title("Bem-vindo").
			Opt("Pacotes de Dados", "/bundles/1").
			Exit("Até logo").
			Prompt(c)
	})
	r.INPUT("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Opt("Pacotes de Dados", "/bundles/1").
			Exit("Até logo").
			Handle(c)
	})

	// --- Paginator: 5 items per page, custom labels, back to /home
	p := menu.NewPaginator("/bundles", bundles, 5).
		WithTitle("Pacotes de Dados").
		WithNavLabels("Anterior", "Seguinte").
		WithBack("/home")

	// Paginated list: /bundles/:page
	r.SHOW("/bundles/:page", func(c *router.Ctx) core.Reply {
		page := mustAtoi(c.Param("page"), 1)
		return p.Render(page).Prompt(c)
	})
	r.INPUT("/bundles/:page", func(c *router.Ctx) core.Reply {
		page := mustAtoi(c.Param("page"), 1)
		return p.Render(page).Handle(c)
	})

	// Item confirmation: /bundles/item/:idx  (idx is absolute index in the catalog)
	r.SHOW("/bundles/item/:idx", func(c *router.Ctx) core.Reply {
		idx := mustAtoi(c.Param("idx"), -1)
		if idx < 0 || idx >= len(bundles) {
			return core.CON("Item inválido.\n0) Voltar")
		}
		chosen := bundles[idx]
		return menu.New("/bundles/item/:idx").
			Title("Confirmar\n"+chosen+"?").
			End("Sim", "Compra concluída.").
			Back("/bundles/1").
			Prompt(c)
	})
	r.INPUT("/bundles/item/:idx", func(c *router.Ctx) core.Reply {
		return menu.New("/bundles/item/:idx").
			End("Sim", "Compra concluída.").
			Back("/bundles/1").
			Handle(c)
	})

	// --- Engine, Store, Transport + Emulator
	st := store.NewInMemoryStore(60 * time.Second)
	eng := core.New(r.Mount(), core.Config{Store: st, SessionTTL: 60 * time.Second})

	mux := http.NewServeMux()
	mux.Handle("/ussd", transport.HTTPHandler(eng))
	emulator.Attach(mux, eng) // /emu and /emu/send

	log.Println("Bundles demo on :8080 — USSD: /ussd  | Emulator: /emu")
	_ = http.ListenAndServe(":8080", mux)
}

// mustAtoi parses s as int, returns def on error.
func mustAtoi(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
