package menu

import (
	"fmt"
	"strings"

	"github.com/grahms/cardinal/core"
	"github.com/grahms/cardinal/router"
)

// Builder composes newline-based option menus with Back/Exit semantics.
type Builder struct {
	path   string
	title  string
	items  []Item
	backTo string
	exitTx string
}

type Item struct {
	Label   string
	Target  string
	Before  func(*router.Ctx) error // optional: run before redirect (e.g., load data)
	EndText string                  // if set, ends session with this text
}

func New(path string) *Builder             { return &Builder{path: path} }
func (b *Builder) Title(s string) *Builder { b.title = s; return b }
func (b *Builder) Opt(label, target string, hooks ...func(*router.Ctx) error) *Builder {
	it := Item{Label: label, Target: target}
	if len(hooks) > 0 {
		it.Before = hooks[0]
	}
	b.items = append(b.items, it)
	return b
}
func (b *Builder) End(label, endText string) *Builder {
	b.items = append(b.items, Item{Label: label, EndText: endText})
	return b
}
func (b *Builder) Back(target string) *Builder { b.backTo = target; return b }
func (b *Builder) Exit(text string) *Builder   { b.exitTx = text; return b }

// Prompt returns a CON reply with the built screen.
func (b *Builder) Prompt(c *router.Ctx) core.Reply {
	var lines []string
	if b.title != "" {
		lines = append(lines, b.title)
	}
	for i, it := range b.items {
		lines = append(lines, fmt.Sprintf("%d) %s", i+1, it.Label))
	}
	if b.backTo != "" {
		lines = append(lines, "0) Voltar")
	}
	if b.exitTx != "" {
		lines = append(lines, "00) Sair")
	}
	return core.CON(strings.Join(lines, "\n"))
}

// Handle consumes c.In() and redirects/ends/re-prompts as needed.
func (b *Builder) Handle(c *router.Ctx) core.Reply {
	in := c.In()
	if in == "" {
		return b.Prompt(c)
	}

	if in == "0" && b.backTo != "" {
		c.Redirect(b.backTo)
		return core.CON("") // engine will SHOW next
	}
	if in == "00" && b.exitTx != "" {
		return core.END(b.exitTx)
	}
	idx, ok := atoi(in)
	if ok && idx >= 1 && idx <= len(b.items) {
		it := b.items[idx-1]
		if it.Before != nil {
			if err := it.Before(c); err != nil {
				return core.END("Serviço indisponível. Tente mais tarde.")
			}
		}
		if it.EndText != "" {
			return core.END(it.EndText)
		}
		if it.Target != "" {
			c.Redirect(it.Target)
			return core.CON("")
		}
	}
	old := b.title
	if old != "" {
		b.title = old + "\n⚠️ Opção inválida."
		defer func() { b.title = old }()
	}
	return b.Prompt(c)
}

func atoi(s string) (int, bool) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}
