package cmdutil

import (
	"fmt"

	"github.com/jmigpin/editor/edit/toolbardata"
	"github.com/jmigpin/editor/ui/tautil"
)

func Replace(erow ERower, part *toolbardata.Part) {
	a := part.Args[1:]
	if len(a) != 2 {
		err := fmt.Errorf("replace: expecting 2 arguments")
		erow.Ed().Error(err)
		return
	}
	old, new := a[0].Trim(), a[1].Trim()
	tautil.Replace(erow.Row().TextArea, old, new)
}
