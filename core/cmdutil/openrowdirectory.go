package cmdutil

import (
	"os"
	"path"

	"github.com/jmigpin/editor/ui"
)

func OpenRowDirectory(ed Editorer) {
	erow, ok := ed.ActiveERower()
	if !ok {
		return
	}

	fp := erow.Filename()
	p := path.Dir(fp) // if fp=="", dir returns "."

	fi, err := os.Stat(p)
	if err != nil {
		ed.Error(err)
		return
	}
	if !fi.IsDir() {
		ed.Errorf("not a directory: %v", p)
		return
	}

	next, _ := erow.Row().NextRow()
	erow2 := OpenDirectoryRow(ed, p, erow.Row().Col, next)
	erow2.Row().Flash()
}
func OpenDirectoryRow(ed Editorer, path string, col *ui.Column, next *ui.Row) ERower {
	erow, ok := ed.FindERower(path)
	if !ok {
		erow = ed.NewERowerBeforeRow(path, col, next)
		err := erow.LoadContentClear()
		if err != nil {
			erow.Ed().Error(err)
		}
	}
	return erow
}
