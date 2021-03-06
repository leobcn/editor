package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/jmigpin/editor/core/toolbarparser"
	"github.com/jmigpin/editor/ui"
)

type Sessions struct {
	Sessions []*Session
}

func NewSessions(filename string) (*Sessions, error) {
	// read file
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// empty sessions if it doesn't exist
			return &Sessions{}, nil
		}
		return nil, err
	}
	ss := Sessions{}
	// decode
	dec := json.NewDecoder(f)
	err = dec.Decode(&ss)
	if err != nil {
		return nil, err
	}
	return &ss, err
}
func (ss *Sessions) save(filename string) error {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	f, err := os.OpenFile(filename, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(&ss)
}

//----------

func sessionsFilename() string {
	home := os.Getenv("HOME")
	return path.Join(home, ".editor_sessions.json")
}

//----------

type Session struct {
	Name      string
	RootTbStr string
	Columns   []*ColumnState

	// DEPRECATED: backward compatible
	//LayoutTbStr string
}

func NewSessionFromEditor(ed *Editor) *Session {
	s := &Session{
		RootTbStr: ed.UI.Root.Toolbar.Str(),
	}
	for _, c := range ed.UI.Root.Cols.Columns() {
		cstate := NewColumnState(ed, c)
		s.Columns = append(s.Columns, cstate)
	}
	return s
}
func (s *Session) restore(ed *Editor) {
	uicols := ed.UI.Root.Cols

	// layout toolbar
	tbStr := s.RootTbStr

	//// DEPRECATED: backward compatible
	//if s.LayoutTbStr != "" {
	//	tbStr = s.LayoutTbStr
	//}

	ed.UI.Root.Toolbar.SetStrClearHistory(tbStr)

	// close all current columns
	for _, c := range uicols.Columns() {
		c.Close()
	}

	// open n new columns
	// allow other columns to exist already (ex: on close, the editor could be ensuring one column)
	for len(uicols.Columns()) < len(s.Columns) {
		_ = ed.NewColumn()
	}

	// setup columns sizes (end percents)
	uicolumns := uicols.Columns()
	for i, c := range s.Columns {
		sp := c.StartPercent

		//// backward compatible
		//if i > 0 && s.Columns[i-1].EndPercent != 0 {
		//	sp = s.Columns[i-1].EndPercent
		//}

		uicols.ColsLayout.Spl.SetRawStartPercent(uicolumns[i], sp)
	}

	// create rows
	m := make(map[*RowState]*ERow)
	for i, c := range s.Columns {
		uicol := uicolumns[i]

		for _, rs := range c.Rows {
			rowPos := &ui.RowPos{Column: uicol}
			erow, ok, err := rs.OpenERow(ed, rowPos)
			if err != nil {
				ed.Error(err)
			}
			if ok {
				m[rs] = erow

				// setup row size
				sp := rs.StartPercent

				//// backward compatible
				//if j > 0 && c.Rows[j-1].EndPercent != 0 {
				//	sp = c.Rows[j-1].EndPercent
				//}

				uicol.RowsLayout.Spl.SetRawStartPercent(erow.Row, sp)
			}
		}
	}

	// restore positions after all rows have been created
	for rs, erow := range m {
		rs.RestorePos(erow)
	}
}

//----------

type ColumnState struct {
	StartPercent float64
	Rows         []*RowState

	// DEPRECATED: backward compatible
	//EndPercent   float64
}

func NewColumnState(ed *Editor, col *ui.Column) *ColumnState {
	cstate := &ColumnState{
		StartPercent: col.Cols.ColsLayout.Spl.RawStartPercent(col),
	}
	for _, row := range col.Rows() {
		rstate := NewRowState(row)
		cstate.Rows = append(cstate.Rows, rstate)
	}
	return cstate
}

//----------

// Used in sessions and reopenrow.
type RowState struct {
	TbStr         string
	TbCursorIndex int
	TaCursorIndex int
	TaOffsetIndex int
	StartPercent  float64

	// DEPRECATED: backward compatible
	//EndPercent    float64
}

func NewRowState(row *ui.Row) *RowState {
	rs := &RowState{
		TbStr:         row.Toolbar.Str(),
		TbCursorIndex: row.Toolbar.TextCursor.Index(),
		TaCursorIndex: row.TextArea.TextCursor.Index(),
		TaOffsetIndex: row.TextArea.OffsetIndex(),
	}

	// check row.col in case the row has been removed from columns (reopenrow?)
	if row.Col != nil {
		rs.StartPercent = row.Col.RowsLayout.Spl.RawStartPercent(row)
	}

	return rs
}

func (state *RowState) OpenERow(ed *Editor, rowPos *ui.RowPos) (*ERow, bool, error) {
	data := toolbarparser.Parse(state.TbStr)
	arg0, ok := data.Part0Arg0()
	if !ok {
		return nil, false, fmt.Errorf("missing toolbar arg 0: %s", state.TbStr)
	}

	name := ed.HomeVars.Decode(arg0.Str())
	info := ed.ReadERowInfo(name)

	// create erow, event if it will have errors
	erow, err := info.NewERowCreateOnErr(rowPos)
	if err != nil {
		ed.Error(err)
		// just reporting error, continue
	}

	// setup toolbar even if erow had errors
	w := data.Str[arg0.End:]
	if strings.TrimSpace(w) != "" {
		erow.ToolbarSetStrAfterNameClearHistory(w)
	}

	if err != nil {
		return erow, ok, err
	}

	return erow, true, nil
}

func (state *RowState) RestorePos(erow *ERow) {
	erow.Row.Toolbar.TextCursor.SetIndex(state.TbCursorIndex)
	erow.Row.TextArea.TextCursor.SetIndex(state.TaCursorIndex)
	erow.Row.TextArea.SetOffsetIndex(state.TaOffsetIndex)
}

//----------

func SaveSession(ed *Editor, part *toolbarparser.Part) {
	err := saveSession(ed, part, sessionsFilename())
	if err != nil {
		ed.Error(err)
	}
}
func saveSession(ed *Editor, part *toolbarparser.Part, filename string) error {
	if len(part.Args) != 2 {
		return fmt.Errorf("savesession: missing session name")
	}
	sessionName := part.Args[1].Str()

	s1 := NewSessionFromEditor(ed)
	s1.Name = sessionName

	ss, err := NewSessions(filename)
	if err != nil {
		return err
	}
	// replace session already stored
	replaced := false
	for i, s := range ss.Sessions {
		if s.Name == sessionName {
			ss.Sessions[i] = s1
			replaced = true
			break
		}
	}
	// append if a new session
	if !replaced {
		ss.Sessions = append(ss.Sessions, s1)
	}
	// save to file
	err = ss.save(filename)
	if err != nil {
		return err
	}
	return nil
}

//----------

func ListSessions(ed *Editor) {
	ss, err := NewSessions(sessionsFilename())
	if err != nil {
		ed.Error(err)
		return
	}

	// sort sessions names
	var u []string
	for _, session := range ss.Sessions {
		u = append(u, session.Name)
	}
	sort.Strings(u)

	// concat opensession lines
	buf := &bytes.Buffer{}
	for _, sname := range u {
		fmt.Fprintf(buf, "OpenSession %v\n", sname)
	}

	erow, _ := ed.ExistingOrNewERow("+Sessions")
	erow.Row.TextArea.SetBytesClearPos(buf.Bytes())
	erow.Flash()
}

//----------

func OpenSession(ed *Editor, part *toolbarparser.Part) {
	if len(part.Args) != 2 {
		ed.Errorf("missing session name")
		return
	}
	sessionName := part.Args[1].Str()
	OpenSessionFromString(ed, sessionName)
}

func OpenSessionFromString(ed *Editor, sessionName string) {
	ss, err := NewSessions(sessionsFilename())
	if err != nil {
		return
	}
	for _, s := range ss.Sessions {
		if s.Name == sessionName {
			s.restore(ed)
			return
		}
	}
	ed.Errorf("session not found: %v", sessionName)
}

//----------

func DeleteSession(ed *Editor, part *toolbarparser.Part) {
	err := deleteSession(ed, part)
	if err != nil {
		ed.Error(err)
	}
}
func deleteSession(ed *Editor, part *toolbarparser.Part) error {
	if len(part.Args) != 2 {
		return fmt.Errorf("deletesession: missing session name")
	}
	sessionName := part.Args[1].Str()
	ss, err := NewSessions(sessionsFilename())
	if err != nil {
		return err
	}
	found := false
	for i, s := range ss.Sessions {
		if s.Name == sessionName {
			found = true
			u := ss.Sessions
			ss.Sessions = append(u[:i], u[i+1:]...)
			break
		}
	}
	if !found {
		return fmt.Errorf("deletesession: session not found: %v", sessionName)
	}
	return ss.save(sessionsFilename())
}
