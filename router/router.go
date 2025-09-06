package router

import (
	"context"
	"strings"

	"github.com/grahms/cardinal/core"
)

type Ctx struct {
	context.Context
	Session *core.Session
	Req     core.Request
	path    string
	in      string
	next    string
	params  map[string]string
}

func (c *Ctx) Path() string             { return c.path }
func (c *Ctx) In() string               { return c.in }
func (c *Ctx) Redirect(p string)        { c.next = p }
func (c *Ctx) Next() string             { return c.next }
func (c *Ctx) Set(k string, v any)      { c.Session.Set(k, v) }
func (c *Ctx) Get(k string) (any, bool) { return c.Session.Get(k) }
func (c *Ctx) Param(k string) string    { return c.params[k] }

type Handler func(*Ctx) core.Reply

type Middleware func(Handler) Handler

type route struct {
	pattern string
	show    Handler
	input   Handler
}

type Router struct {
	start string
	exact map[string]route
	param []route

	mws []Middleware
}

func New(start string) *Router {
	return &Router{start: start, exact: map[string]route{}, param: []route{}}
}

func (rt *Router) Use(mw ...Middleware) { rt.mws = append(rt.mws, mw...) }

func (rt *Router) SHOW(path string, h Handler)  { rt.add(path, h, true) }
func (rt *Router) INPUT(path string, h Handler) { rt.add(path, h, false) }

func (rt *Router) add(path string, h Handler, isShow bool) {
	h = wrap(h, rt.mws) // only globals
	rt.addCore(path, h, isShow)
}

func (rt *Router) Mount() core.App { return &app{rt: rt} }

type app struct{ rt *Router }

func (a *app) Handle(ctx context.Context, s *core.Session, req core.Request) (core.Reply, error) {
	path := mustString(s, "_p")
	if path == "" {
		path = a.rt.start
		s.Set("_p", path)
		return a.execSHOW(ctx, s, req, path), nil
	}

	input := lastToken(req.Text)
	if input == "" {
		return a.execSHOW(ctx, s, req, path), nil
	}

	reply := a.execINPUT(ctx, s, req, path, input)
	if !reply.Continue {
		return reply, nil
	}
	if next := mustString(s, "_next"); next != "" {
		s.Set("_p", next)
		s.Set("_next", "")
		return a.execSHOW(ctx, s, req, next), nil
	}
	return a.execSHOW(ctx, s, req, path), nil
}

// SHOWWith/INPUTWith attach per-route middleware (after globals).
func (rt *Router) SHOWWith(path string, h Handler, mw ...Middleware) {
	rt.addWith(path, h, true, mw...)
}
func (rt *Router) INPUTWith(path string, h Handler, mw ...Middleware) {
	rt.addWith(path, h, false, mw...)
}

func (rt *Router) addWith(path string, h Handler, isShow bool, mw ...Middleware) {
	// global first, then per-route (outermost first via wrap)
	full := append([]Middleware{}, rt.mws...) // copy
	full = append(full, mw...)                // route-level
	h = wrap(h, full)
	rt.addCore(path, h, isShow)
}

func (rt *Router) addCore(path string, h Handler, isShow bool) {
	if strings.Contains(path, ":") {
		for i := range rt.param {
			if rt.param[i].pattern == path {
				if isShow {
					rt.param[i].show = h
				} else {
					rt.param[i].input = h
				}
				return
			}
		}
		r := route{pattern: path}
		if isShow {
			r.show = h
		} else {
			r.input = h
		}
		rt.param = append(rt.param, r)
		return
	}
	r := rt.exact[path]
	r.pattern = path
	if isShow {
		r.show = h
	} else {
		r.input = h
	}
	rt.exact[path] = r
}

func (a *app) execSHOW(ctx context.Context, s *core.Session, req core.Request, path string) core.Reply {
	h, params := a.match(path, true)
	if h == nil {
		return core.END("Service unavailable.")
	}
	cc := &Ctx{Context: ctx, Session: s, Req: req, path: path, params: params}
	return h(cc)
}
func (a *app) execINPUT(ctx context.Context, s *core.Session, req core.Request, path, in string) core.Reply {
	h, params := a.match(path, false)
	if h == nil {
		return core.END("Service unavailable.")
	}
	cc := &Ctx{Context: ctx, Session: s, Req: req, path: path, in: in, params: params}
	reply := h(cc)
	if cc.next != "" {
		s.Set("_next", cc.next)
	}
	return reply
}

func (a *app) match(path string, wantSHOW bool) (Handler, map[string]string) {
	if r, ok := a.rt.exact[path]; ok {
		if wantSHOW {
			return r.show, nil
		}
		return r.input, nil
	}
	for _, r := range a.rt.param {
		if params, ok := matchParams(path, r.pattern); ok {
			if wantSHOW {
				return r.show, params
			}
			return r.input, params
		}
	}
	return nil, nil
}

func wrap(h Handler, mws []Middleware) Handler {
	// apply in reverse so first Use is outermost (Use(A,B) => A(B(h)))
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func matchParams(path, pattern string) (map[string]string, bool) {
	ps := splitClean(path)
	pp := splitClean(pattern)
	if len(ps) != len(pp) {
		return nil, false
	}
	out := map[string]string{}
	for i := range ps {
		if strings.HasPrefix(pp[i], ":") {
			out[pp[i][1:]] = ps[i]
			continue
		}
		if pp[i] != ps[i] {
			return nil, false
		}
	}
	return out, true
}

func splitClean(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if s[0] == '/' {
		s = s[1:]
	}
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "/")
}
func mustString(s *core.Session, k string) string {
	if v, ok := s.Get(k); ok {
		if x, ok := v.(string); ok {
			return x
		}
	}
	return ""
}
func lastToken(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return ""
	}
	parts := strings.Split(t, "*")
	return parts[len(parts)-1]
}
