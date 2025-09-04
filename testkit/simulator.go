package testkit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/grahms/cardinal/core"
)

// Step is optional if you want to predefine scripted flows.
type Step struct {
	Input  string // what the user types (last token)
	Expect string // substring expected in the next screen
	End    bool   // whether we expect the session to end
}

// Simulator drives the Engine with fake requests and lets you assert responses.
type Simulator struct {
	t       *testing.T
	eng     *core.Engine
	session string
	msisdn  string
	service string
	last    core.Reply
	ctx     context.Context
}

// New builds a simulator around an Engine.
func New(t *testing.T, eng *core.Engine) *Simulator {
	return &Simulator{
		t:   t,
		eng: eng,
		ctx: context.Background(),
	}
}

// Start begins a session with an MSISDN (and optional service code).
func (s *Simulator) Start(msisdn string, service ...string) *Simulator {
	s.session = fmt.Sprintf("sess-%d", time.Now().UnixNano())
	s.msisdn = msisdn
	if len(service) > 0 {
		s.service = service[0]
	}

	// First call: empty text triggers SHOW of start path.
	rep, err := s.eng.Handle(s.ctx, core.Request{
		SessionID:   s.session,
		Msisdn:      s.msisdn,
		ServiceCode: s.service,
		Text:        "",
	})
	if err != nil {
		s.t.Fatalf("start: %v", err)
	}
	s.last = rep
	return s
}

// Send simulates user input (usually just the last token).
func (s *Simulator) Send(token string) *Simulator {
	rep, err := s.eng.Handle(s.ctx, core.Request{
		SessionID:   s.session,
		Msisdn:      s.msisdn,
		ServiceCode: s.service,
		Text:        token,
	})
	if err != nil {
		s.t.Fatalf("send(%q): %v", token, err)
	}
	s.last = rep
	return s
}

// Expect asserts that the current screen (CON) contains a substring.
func (s *Simulator) Expect(substr string) *Simulator {
	if !s.last.Continue {
		s.t.Fatalf("expected CON with substring %q, but got END: %q", substr, s.last.Message)
	}
	if !strings.Contains(s.last.Message, substr) {
		s.t.Fatalf("expected substring %q in %q", substr, s.last.Message)
	}
	return s
}

// ExpectEndsWith asserts that the session ended with a message containing substring.
func (s *Simulator) ExpectEndsWith(substr string) *Simulator {
	if s.last.Continue {
		s.t.Fatalf("expected END, but got CON: %q", s.last.Message)
	}
	if !strings.Contains(s.last.Message, substr) {
		s.t.Fatalf("expected end message to contain %q, got %q", substr, s.last.Message)
	}
	return s
}
