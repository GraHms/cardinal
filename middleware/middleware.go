package middleware

import (
	"log"
	"sync"
	"time"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/router"
)

// Recover: guards against panics and returns a safe END message.
func Recover() router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(c *router.Ctx) (rep core.Reply) {
			defer func() {
				if r := recover(); r != nil {
					// never let a panic leak to the transport
					rep = core.END("Service unavailable.")
				}
			}()
			return next(c)
		}
	}
}

// Logging: minimal structured logging using stdlib log.Logger.
func Logging(l *log.Logger) router.Middleware {
	return func(next router.Handler) router.Handler {
		return func(c *router.Ctx) core.Reply {
			start := time.Now()
			rep := next(c)
			l.Printf("sid=%s msisdn=%s path=%s in=%q continue=%t latency_ms=%d",
				c.Session.ID(),
				c.Req.Msisdn,
				c.Path(),
				c.In(),
				rep.Continue,
				time.Since(start).Milliseconds(),
			)
			return rep
		}
	}
}

// RateLimitPerMSISDN: simple token-bucket per phone number.
// allow N requests per window (rough approximation, reset on window).
func RateLimitPerMSISDN(limit int, window time.Duration) router.Middleware {
	b := &bucket{
		limit:  limit,
		window: window,
		m:      map[string]*entry{},
	}
	return func(next router.Handler) router.Handler {
		return func(c *router.Ctx) core.Reply {
			if !b.allow(c.Req.Msisdn) {
				return core.END("Busy. Please try again.")
			}
			return next(c)
		}
	}
}

type entry struct {
	cnt   int
	reset time.Time
}
type bucket struct {
	mu     sync.Mutex
	m      map[string]*entry
	limit  int
	window time.Duration
}

func (b *bucket) allow(key string) bool {
	now := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.m[key]
	if !ok || now.After(e.reset) {
		b.m[key] = &entry{cnt: 1, reset: now.Add(b.window)}
		return true
	}
	if e.cnt < b.limit {
		e.cnt++
		return true
	}
	return false
}
