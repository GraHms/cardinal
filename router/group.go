package router

// Group represents a sub-router with a fixed path prefix and extra middlewares.
type Group struct {
	rt     *Router
	prefix string
	mws    []Middleware // group-level middleware (applied after global Use(...) and before route-level)
}

// Group creates a new group from the root router.
func (rt *Router) Group(prefix string, mw ...Middleware) *Group {
	return &Group{rt: rt, prefix: cleanPrefix(prefix), mws: mw}
}

// Group creates a subgroup from another group (supports nesting).
func (g *Group) Group(prefix string, mw ...Middleware) *Group {
	return &Group{
		rt:     g.rt,
		prefix: join(g.prefix, cleanPrefix(prefix)),
		mws:    append([]Middleware{}, append(g.mws, mw...)...), // inherit + add
	}
}

/* ---------- Group route registration ---------- */

// SHOW registers a SHOW handler under the group's prefix (globals + group mws).
func (g *Group) SHOW(path string, h Handler) {
	fullPath := join(g.prefix, cleanPrefix(path))
	fullMws := append([]Middleware{}, g.rt.mws...) // globals
	fullMws = append(fullMws, g.mws...)            // group-level
	g.rt.addCore(fullPath, wrap(h, fullMws), true)
}

// INPUT registers an INPUT handler under the group's prefix (globals + group mws).
func (g *Group) INPUT(path string, h Handler) {
	fullPath := join(g.prefix, cleanPrefix(path))
	fullMws := append([]Middleware{}, g.rt.mws...)
	fullMws = append(fullMws, g.mws...)
	g.rt.addCore(fullPath, wrap(h, fullMws), false)
}

// SHOWWith adds per-route middleware as well (globals + group + route mws).
func (g *Group) SHOWWith(path string, h Handler, mw ...Middleware) {
	fullPath := join(g.prefix, cleanPrefix(path))
	fullMws := append([]Middleware{}, g.rt.mws...) // globals
	fullMws = append(fullMws, g.mws...)            // group
	fullMws = append(fullMws, mw...)               // route
	g.rt.addCore(fullPath, wrap(h, fullMws), true)
}

// INPUTWith adds per-route middleware as well (globals + group + route mws).
func (g *Group) INPUTWith(path string, h Handler, mw ...Middleware) {
	fullPath := join(g.prefix, cleanPrefix(path))
	fullMws := append([]Middleware{}, g.rt.mws...)
	fullMws = append(fullMws, g.mws...)
	fullMws = append(fullMws, mw...)
	g.rt.addCore(fullPath, wrap(h, fullMws), false)
}

/* ---------- small helpers ---------- */

func cleanPrefix(p string) string {
	if p == "" || p == "/" {
		return ""
	}
	if p[0] != '/' {
		p = "/" + p
	}
	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}
func join(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if a[len(a)-1] == '/' {
		a = a[:len(a)-1]
	}
	if b[0] != '/' {
		b = "/" + b
	}
	return a + b
}
