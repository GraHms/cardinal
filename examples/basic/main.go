package main

import (
	"log"
	"net/http"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/menu"
	"github.com/grahms/cardinal/middleware"
	"github.com/grahms/cardinal/router"
	"github.com/grahms/cardinal/store"
	"github.com/grahms/cardinal/transport"
)

func main() {
	r := router.New("/home")

	// Attach middlewares (order matters: first is outermost)
	r.Use(
		middleware.Recover(),
		middleware.Logging(log.Default()),
		middleware.RateLimitPerMSISDN(5, 5*time.Second), // 5 req / 5s per MSISDN
	)

	r.SHOW("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Title("Welcome").
			Opt("Check Balance", "/balance").
			Opt("Buy Airtime", "/amount").
			Exit("Goodbye.").
			Prompt(c)
	})
	r.INPUT("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Opt("Check Balance", "/balance").
			Opt("Buy Airtime", "/amount").
			Exit("Goodbye.").
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

	st := store.NewInMemoryStore(60 * time.Second)
	eng := core.New(r.Mount(), core.Config{Store: st})

	mux := http.NewServeMux()
	mux.Handle("/ussd", transport.HTTPHandler(eng))
	log.Println("Cardinal with middleware on :8080")
	_ = http.ListenAndServe(":8080", mux)
}
