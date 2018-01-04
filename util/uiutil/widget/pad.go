package widget

import (
	"image"
	"image/color"

	"github.com/jmigpin/editor/util/imageutil"
)

// Can be used as border. If no color is set, it won't paint.
type Pad struct {
	EmbedNode
	Top, Right, Bottom, Left int
	Color                    *color.Color
	ctx                      Context
}

func NewPad(ctx Context, child Node) *Pad {
	p := &Pad{ctx: ctx}
	p.SetWrapper(p)
	p.Append(child)
	return p
}
func (p *Pad) Set(v int) {
	p.Top = v
	p.Right = v
	p.Bottom = v
	p.Left = v
}
func (p *Pad) Measure(hint image.Point) image.Point {
	h := hint
	h.X -= p.Right + p.Left
	h.Y -= p.Top + p.Bottom
	h = MaxPoint(h, image.Point{0, 0})
	m := p.EmbedNode.Measure(h)
	m.X += p.Right + p.Left
	m.Y += p.Top + p.Bottom
	m = MinPoint(m, hint)
	return m
}
func (p *Pad) CalcChildsBounds() {
	u := p.Bounds
	u.Min = u.Min.Add(image.Point{p.Left, p.Top})
	u.Max = u.Max.Sub(image.Point{p.Right, p.Bottom})
	u = u.Intersect(p.Bounds)
	p.IterChilds(func(c Node) {
		c.Embed().Bounds = u
		c.CalcChildsBounds()
	})
}
func (p *Pad) Paint() {
	if p.Color == nil {
		return
	}
	b := p.Bounds
	// top
	u := b
	u.Max.Y = u.Min.Y + p.Top
	u = u.Intersect(b)
	imageutil.FillRectangle(p.ctx.Image(), &u, *p.Color)
	// bottom
	u = b
	u.Min.Y = u.Max.Y - p.Bottom
	u = u.Intersect(b)
	imageutil.FillRectangle(p.ctx.Image(), &u, *p.Color)
	// right
	u = b
	u.Min.X = u.Max.X - p.Right
	u = u.Intersect(b)
	imageutil.FillRectangle(p.ctx.Image(), &u, *p.Color)
	// left
	u = b
	u.Max.X = u.Min.X + p.Left
	u = u.Intersect(b)
	imageutil.FillRectangle(p.ctx.Image(), &u, *p.Color)
}