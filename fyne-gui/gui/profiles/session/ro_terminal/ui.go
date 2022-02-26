package ro_terminal

import (
	"image/color"
	"math"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

// Config is the state of a terminal, updated upon certain actions or commands.
// Use Terminal.OnConfigure hook to register for changes.
type Config struct {
	Rows, Columns uint
}

// Terminal is a terminal widget that loads a shell and handles input/output.
type Terminal struct {
	widget.BaseWidget
	fyne.ShortcutHandler
	content  *widget.TextGrid
	config   Config
	startDir string

	bell, bright, debug, focused bool
	currentFG, currentBG         color.Color
	cursorRow, cursorCol         int
	savedRow, savedCol           int
	scrollTop, scrollBottom      int

	cursor                   *canvas.Rectangle
	cursorHidden, bufferMode bool // buffer mode is an xterm extension that impacts control keys
	cursorMoved              func()

	onMouseDown, onMouseUp func(int, desktop.Modifier, fyne.Position)
}

// MouseDown handles the down action for desktop mouse events.
func (t *Terminal) MouseDown(ev *desktop.MouseEvent) {
	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseDown(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseDown(2, ev.Modifier, ev.Position)
	}
}

// MouseUp handles the up action for desktop mouse events.
func (t *Terminal) MouseUp(ev *desktop.MouseEvent) {
	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseUp(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseUp(2, ev.Modifier, ev.Position)
	}
}

// Resize is called when this terminal widget has been resized.
// It ensures that the virtual terminal is within the bounds of the widget.
func (t *Terminal) Resize(s fyne.Size) {
	if s.Width == t.Size().Width && s.Height == t.Size().Height {
		return
	}
	if s.Width < 20 { // not sure why we get tiny sizes
		return
	}
	t.BaseWidget.Resize(s)
	t.content.Resize(s)

	cellSize := t.guessCellSize()
	oldRows := int(t.config.Rows)

	t.config.Columns = uint(math.Floor(float64(s.Width) / float64(cellSize.Width)))
	t.config.Rows = uint(math.Floor(float64(s.Height) / float64(cellSize.Height)))
	if t.scrollBottom == 0 || t.scrollBottom == oldRows-1 {
		t.scrollBottom = int(t.config.Rows) - 1
	}
}

// TouchCancel handles the tap action for mobile apps that lose focus during tap.
func (t *Terminal) TouchCancel(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

// TouchDown handles the down action for mobile touch events.
func (t *Terminal) TouchDown(ev *mobile.TouchEvent) {
	if t.onMouseDown != nil {
		t.onMouseDown(1, 0, ev.Position)
	}
}

// TouchUp handles the up action for mobile touch events.
func (t *Terminal) TouchUp(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

// don't call often - should we cache?
func (t *Terminal) guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	min := cell.MinSize()
	return fyne.NewSize(float32(math.Round(float64(min.Width))), float32(math.Round(float64(min.Height))))
}

func (t *Terminal) handleOutputChar(r rune) {
	if t.cursorCol >= int(t.config.Columns) || t.cursorRow >= int(t.config.Rows) {
		return // TODO handle wrap?
	}
	for len(t.content.Rows)-1 < t.cursorRow {
		t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
	}

	cellStyle := &widget.CustomTextGridStyle{FGColor: t.currentFG, BGColor: t.currentBG}
	for len(t.content.Rows[t.cursorRow].Cells)-1 < t.cursorCol {
		newCell := widget.TextGridCell{
			Rune:  ' ',
			Style: cellStyle,
		}
		t.content.Rows[t.cursorRow].Cells = append(t.content.Rows[t.cursorRow].Cells, newCell)
	}

	cell := t.content.Rows[t.cursorRow].Cells[t.cursorCol]
	if cell.Rune != r || cell.Style.TextColor() != cellStyle.FGColor || cell.Style.BackgroundColor() != cellStyle.BGColor {
		cell.Rune = r
		cell.Style = cellStyle
		t.content.SetCell(t.cursorRow, t.cursorCol, cell)
	}
	t.cursorCol++
}

func (t *Terminal) appendRune(r rune) {
	if r == rune("\n"[0]) {
		t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
		return
	}
	if len(t.content.Rows) == 0 {
		t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
	}
	if len(t.content.Rows[len(t.content.Rows)-1].Cells) == int(t.config.Columns) {
		t.content.Rows = append(t.content.Rows, widget.TextGridRow{})
	}
	cellStyle := &widget.CustomTextGridStyle{FGColor: t.currentFG, BGColor: t.currentBG}
	newCell := widget.TextGridCell{
		Rune:  r,
		Style: cellStyle,
	}
	t.content.SetCell(len(t.content.Rows)-1, len(t.content.Rows[len(t.content.Rows)-1].Cells), newCell)
	if r > 128 {
		// maybe chinese
		// Block                                   Range       Comment
		// CJK Unified Ideographs                  4E00-9FFF   Common
		// CJK Unified Ideographs Extension A      3400-4DBF   Rare
		// CJK Unified Ideographs Extension B      20000-2A6DF Rare, historic
		// CJK Unified Ideographs Extension C      2A700–2B73F Rare, historic
		// CJK Unified Ideographs Extension D      2B740–2B81F Uncommon, some in current use
		// CJK Unified Ideographs Extension E      2B820–2CEAF Rare, historic
		// CJK Compatibility Ideographs            F900-FAFF   Duplicates, unifiable variants, corporate characters
		// CJK Compatibility Ideographs Supplement 2F800-2FA1F Unifiable variants
		if r >= 0x4e00 && r <= 0x9fff || r >= 0x3400 && r <= 0x4dbf || r >= 0x20000 && r <= 0x2a6df || r >= 0x2a700 && r <= 0x2b73f || r >= 0x2b740 && r <= 0x2b81f || r >= 0x2b820 && r <= 0x2ceaf || r >= 0xf900 && r <= 0xfaff || r >= 0x2f800 && r <= 0x2fa1f {
			// add a space, because chinese takes two cells
			t.appendRune(rune(" "[0]))
		}
	}
}

func (t *Terminal) AppendString(data string) {
	offset := 0
	for {
		r, len := utf8.DecodeRuneInString(data[offset:])
		if len != 0 {
			t.appendRune(r)
			offset += len
			if len != 1 {

			}
		} else {
			t.Refresh()
			return
		}
	}
}

// New sets up a new terminal instance with the bash shell
func New() *Terminal {
	t := &Terminal{}
	t.ExtendBaseWidget(t)
	t.content = widget.NewTextGrid()

	return t
}
