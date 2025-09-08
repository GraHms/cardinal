package main

import (
	"log"
	"net/http"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/emulator"
	"github.com/grahms/cardinal/menu"
	"github.com/grahms/cardinal/router"
	"github.com/grahms/cardinal/store"
	"github.com/grahms/cardinal/transport"
)

func main() {
	// Minimal routes
	r := router.New("/home")
	r.SHOW("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Title("Hello from Cardinal").
			Opt("Balance", "/balance").
			Exit("Goodbye").Prompt(c)
	})
	r.INPUT("/home", func(c *router.Ctx) core.Reply {
		return menu.New("/home").
			Opt("Balance", "/balance").
			Exit("Goodbye").Handle(c)
	})
	r.SHOW("/balance", func(c *router.Ctx) core.Reply {
		return menu.New("/balance").Title("Balance: 123.45 MZN").Back("/home").Prompt(c)
	})
	r.INPUT("/balance", func(c *router.Ctx) core.Reply {
		return menu.New("/balance").Back("/home").Handle(c)
	})

	// Engine + store
	eng := core.New(r.Mount(), core.Config{Store: store.NewInMemoryStore(60 * time.Second)})

	// Mux with multiple endpoints illustrating vendor adapters:
	mux := http.NewServeMux()

	// Generic/previous handler (form fields: sessionId, phoneNumber, text)
	mux.Handle("/ussd", transport.HTTPHandler(eng)) // if you already have this in transport

	// Africa's Talking (form)
	mux.Handle("/ussd/at", transport.AfricaTalkingHandler(eng))

	// Vodacom (JSON)
	mux.Handle("/ussd/voda", transport.VodacomHandler(eng))

	// Infobip (form; uppercase keys by default)
	mux.Handle("/ussd/infobip", transport.InfobipFormHandler(eng))

	// Generic JSON (custom keys)
	mux.Handle("/ussd/json", transport.JSONGenericHandler(
		eng,
		transport.JSONMap{InSessionID: "sid", InMsisdn: "from", InText: "input"},
		transport.JSONMap{OutTextKey: "message", OutWrapperKey: "type", OutWrapperVal: "Response"},
	))

	// Emulator (nice for local testing)
	emulator.Attach(mux, eng)

	log.Println("Adapters demo on :8080")
	log.Println("  - Emulator:        GET /emu")
	log.Println("  - Generic default: POST /ussd (form: sessionId, phoneNumber, text)")
	log.Println("  - Africa's Talking:POST /ussd/at (form: sessionId, phoneNumber, text)")
	log.Println("  - Vodacom JSON:    POST /ussd/voda (json: sessionId, msisdn, userInput)")
	log.Println("  - Infobip form:    POST /ussd/infobip (form: SESSION_ID, MSISDN, INPUT)")
	log.Println("  - JSON generic:    POST /ussd/json (json: sid, from, input)")
	_ = http.ListenAndServe(":8080", mux)
}
