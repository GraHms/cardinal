package menu

import (
	"fmt"
	"strconv"
)

// Paginator renders long lists into numbered pages using only Builder.Opt(...) and Builder.Back(...).
// - Prev/Next are appended as normal numbered options.
// - Back remains "0) Voltar" (handled by Builder.Back).
// - You can customize the title and Prev/Next labels; keys are the next numbers.
type Paginator struct {
	BasePath  string   // e.g. "/bundles"
	Items     []string // display labels for items
	Size      int      // items per page (default 5)
	Title     string   // screen title (optional)
	PrevLabel string   // default: "Prev"
	NextLabel string   // default: "Next"
	BackTo    string   // optional: where "0) Voltar" goes
}

// NewPaginator creates a paginator with sane defaults.
func NewPaginator(path string, items []string, size int) *Paginator {
	if size <= 0 {
		size = 5
	}
	return &Paginator{
		BasePath:  path,
		Items:     items,
		Size:      size,
		Title:     "Menu",
		PrevLabel: "Prev",
		NextLabel: "Next",
	}
}

func (p *Paginator) WithTitle(title string) *Paginator {
	if title != "" {
		p.Title = title
	}
	return p
}

func (p *Paginator) WithNavLabels(prev, next string) *Paginator {
	if prev != "" {
		p.PrevLabel = prev
	}
	if next != "" {
		p.NextLabel = next
	}
	return p
}

func (p *Paginator) WithBack(target string) *Paginator {
	p.BackTo = target
	return p
}

// Render builds a menu.Builder for the given 1-based page.
// Item targets are "/<base>/item/<absIndex>" so your route can read which item was chosen.
func (p *Paginator) Render(page int) *Builder {
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * p.Size
	if start >= len(p.Items) {
		start = 0
		page = 1
	}
	end := start + p.Size
	if end > len(p.Items) {
		end = len(p.Items)
	}

	b := New(p.BasePath + "/" + strconv.Itoa(page)).Title(p.Title)

	// Page items (1..N)
	for i, label := range p.Items[start:end] {
		abs := start + i // 0-based absolute index
		b.Opt(label, fmt.Sprintf("%s/item/%d", p.BasePath, abs))
	}

	// Prev/Next as numbered options (N+1, N+2)
	hasPrev := start > 0
	hasNext := end < len(p.Items)
	if hasPrev {
		b.Opt(p.PrevLabel, fmt.Sprintf("%s/%d", p.BasePath, page-1))
	}
	if hasNext {
		b.Opt(p.NextLabel, fmt.Sprintf("%s/%d", p.BasePath, page+1))
	}

	// Back (0) if requested
	if p.BackTo != "" {
		b.Back(p.BackTo)
	}
	return b
}
