package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/menu"
	"github.com/grahms/cardinal/router"
	"github.com/grahms/cardinal/store"
	"github.com/grahms/cardinal/transport"
)

func main() {
	r := router.New("/home")

	// Home
	r.SHOW("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Title("Bem-vindo").
			Opt("Consultar saldo", "/balance").
			Opt("Comprar airtime", "/amount").
			Exit("Até já.").
			Prompt(c)
	})
	r.INPUT("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Opt("Consultar saldo", "/balance").
			Opt("Comprar airtime", "/amount").
			Exit("Até já.").
			Handle(c)
	})

	// Balance
	r.SHOW("/balance", func(c *router.Ctx) core.Reply {
		// pretend to fetch balance quickly
		c.Set("bal", 123.45)
		return menu.New("/balance").
			Title("Saldo: 123.45 MZN").
			Back("/home").
			Prompt(c)
	})
	r.INPUT("/balance", func(c *router.Ctx) core.Reply {
		return menu.New("/balance").Back("/home").Handle(c)
	})

	// Amount capture (free-form)
	r.SHOW("/amount", func(c *router.Ctx) core.Reply {
		return core.CON("Introduza o montante (MZN):")
	})
	r.INPUT("/amount", func(c *router.Ctx) core.Reply {
		in := c.In()
		if !digits(in) {
			return core.CON("Montante inválido. Tente novamente:")
		}
		c.Set("amount", atoi(in))
		c.Redirect("/confirm/airtime")
		return core.CON("")
	})

	// Confirm (no params)
	r.SHOW("/confirm/airtime", func(c *router.Ctx) core.Reply {
		amt, _ := c.Get("amount")
		return menu.New("/confirm/airtime").
			Title(fmt.Sprintf("Confirmar %v MZN?", amt)).
			End("Sim", "Compra efetuada. Obrigado.").
			Back("/home").
			Prompt(c)
	})
	r.INPUT("/confirm/airtime", func(c *router.Ctx) core.Reply {
		return menu.New("/confirm/airtime").
			End("Sim", "Compra efetuada. Obrigado.").
			Back("/home").
			Handle(c)
	})

	st := store.NewInMemoryStore(60 * time.Second)
	eng := core.New(r.Mount(), core.Config{Store: st, SessionTTL: 60 * time.Second})

	h := transport.HTTPHandler(eng)
	mux := http.NewServeMux()
	mux.Handle("/ussd", h)

	log.Println("Cardinal example listening on :8080  (POST form: sessionId, phoneNumber, text)")
	_ = http.ListenAndServe(":8080", mux)
}

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
func atoi(s string) int {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n
}
