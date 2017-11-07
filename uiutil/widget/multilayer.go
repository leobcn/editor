package widget

import (
	"image"
)

// First child is bottom layer.
type MultiLayer struct {
	ContainerEmbedNode
}

// Level 1 nodes are nodes immediatly after Level 0 (bottom layer) so they stay always below other upper layers (like menus or floatboxes)
func (ml *MultiLayer) InsertLevel1(n Node) {
	child := ml.FirstChild().Embed()
	if child == nil || child.NextInAll() == nil {
		ml.Append(n)
	} else {
		ml.InsertBefore(n, child.NextInAll())
	}
}

func (ml *MultiLayer) MarkChildNeedsPaint(child Node, r *image.Rectangle) {
	ml.ContainerEmbedNode.MarkChildNeedsPaint(child, r)
	for _, c := range ml.Childs() {
		if c == child {
			continue
		}
		ml.visit(c, r)
	}
}
func (ml *MultiLayer) visit(n Node, r *image.Rectangle) {
	if n.Marks().NeedsPaint() {
		return
	}
	if !n.Bounds().Overlaps(*r) {
		return
	}
	if n.Bounds().Eq(*r) {
		n.MarkNeedsPaint() // highly recursive from here
		return
	}

	// overlap

	// if the childs union doesn't contain the rectangle, this node needs paint
	var u image.Rectangle
	for _, c := range n.Childs() {
		u = u.Union(c.Bounds())
	}
	if !r.In(u) {
		n.MarkNeedsPaint() // highly recursive from here
		return
	}

	// visit each child to see which ones contain or partially contain the rectangle
	for _, c := range n.Childs() {
		ml.visit(c, r)
	}
}

func (ml *MultiLayer) Measure(hint image.Point) image.Point {
	panic("calling measure on multilayer")
}
func (ml *MultiLayer) CalcChildsBounds() {
	u := ml.Bounds()
	for _, n := range ml.Childs() {
		// all childs get full bounds
		n.SetBounds(&u)

		n.CalcChildsBounds()
	}
}
