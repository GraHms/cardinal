package core

import (
	"context"
	"errors"
	"time"
)

// Request is the normalized inbound USSD request from an aggregator/MNO.
type Request struct {
	SessionID   string
	Msisdn      string
	ServiceCode string
	Text        string            // raw text e.g. "1*200"
	Meta        map[string]string // optional vendor-specific metadata
}

// Reply is the outbound USSD response.
type Reply struct {
	Continue bool   // true => "CON", false => "END"
	Message  string // screen body
}

func CON(msg string) Reply { return Reply{Continue: true, Message: msg} }
func END(msg string) Reply { return Reply{Continue: false, Message: msg} }

// Session holds per-session key/value state.
type Session struct {
	id   string
	data map[string]any
}

func (s *Session) ID() string               { return s.id }
func (s *Session) Data() map[string]any     { return s.data }
func (s *Session) Get(k string) (any, bool) { v, ok := s.data[k]; return v, ok }
func (s *Session) Set(k string, v any)      { s.data[k] = v }
func (s *Session) Del(k string)             { delete(s.data, k) }
func (s *Session) MustString(k string) string {
	if v, ok := s.data[k].(string); ok {
		return v
	}
	return ""
}
func (s *Session) MustInt(k string) int {
	if v, ok := s.data[k].(int); ok {
		return v
	}
	return 0
}

// Store is a pluggable session store (e.g., in-memory, Redis).
type Store interface {
	Get(ctx context.Context, sid string) (map[string]any, error)
	Put(ctx context.Context, sid string, data map[string]any, ttl time.Duration) error
	Del(ctx context.Context, sid string) error
}

// App is implemented by your application. Cardinal calls Handle for each step.
type App interface {
	Handle(ctx context.Context, s *Session, req Request) (Reply, error)
}

// Config for the Engine.
type Config struct {
	Store      Store
	SessionTTL time.Duration // default 60s if zero
}

// Engine coordinates session state and calls the App.
type Engine struct {
	cfg Config
	app App
}

func New(app App, cfg Config) *Engine {
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 60 * time.Second
	}
	return &Engine{cfg: cfg, app: app}
}

// Handle processes a single USSD step. It loads the session, delegates to the app,
// and persists or deletes the session depending on the reply.
func (e *Engine) Handle(ctx context.Context, req Request) (Reply, error) {
	if req.SessionID == "" {
		return END("Invalid session"), errors.New("missing session id")
	}

	data, _ := e.cfg.Store.Get(ctx, req.SessionID)
	if data == nil {
		data = map[string]any{}
	}
	s := &Session{id: req.SessionID, data: data}

	reply, err := e.app.Handle(ctx, s, req)

	if err != nil || !reply.Continue {
		_ = e.cfg.Store.Del(ctx, req.SessionID)
		return reply, err
	}
	_ = e.cfg.Store.Put(ctx, req.SessionID, s.Data(), e.cfg.SessionTTL)
	return reply, nil
}
