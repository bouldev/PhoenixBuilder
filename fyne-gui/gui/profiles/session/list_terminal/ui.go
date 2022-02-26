package list_terminal

import (
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	// "golang.design/x/clipboard"
	// "golang.design/x/clipboard"
)

// func init() {
// 	clipboard.Init()
// }

type Terminal struct {
	masterWindow fyne.Window
	data         *[]string
	header       *[]string
	list         *fyne.Container
	content      *container.Scroll
	OnPasteFn    func(s string)
}

func (t *Terminal) AppendNewLine(line string, canCopy bool) {
	offset := 0
	width := 0
	for {
		_, len := utf8.DecodeRuneInString(line[offset:])
		offset += len
		if len != 0 {
			// maybe chinese
			width += 1
			if len > 1 {
				width += 1
			}
		} else {
			break
		}
	}
	*t.data = append(*t.data, line)
	var textObect fyne.CanvasObject
	// TODO this should automatically be computed
	if width < 45 {
		l := widget.NewLabel(line)
		l.Wrapping = fyne.TextTruncate
		textObect = l
	} else {
		l := widget.NewLabel(line)
		l.Wrapping = fyne.TextTruncate
		fulldescription := widget.NewMultiLineEntry()
		fulldescription.Wrapping = fyne.TextWrapBreak
		fulldescription.Hide()
		fulldescription.Text = line
		textObect = container.NewBorder(nil, fulldescription, &widget.Button{
			Icon:          theme.VisibilityIcon(),
			Importance:    widget.LowImportance,
			IconPlacement: widget.ButtonIconTrailingText,
			OnTapped: func() {
				if fulldescription.Hidden {
					fulldescription.Show()
				} else {
					fulldescription.Hide()
				}
			},
		}, nil, l)
	}
	if canCopy || width >= 45 {
		t.list.Add(container.NewBorder(
			nil, nil, //widget.NewLabel(fmt.Sprintf("%d", len(*t.data)+1)),
			&widget.Button{
				Icon:          theme.ContentCopyIcon(),
				Importance:    widget.LowImportance,
				IconPlacement: widget.ButtonIconLeadingText,
				OnTapped: func() {
					// clipboard.Write(clipboard.FmtText, []byte(line))
					// fmt.Println(line)
					t.masterWindow.Clipboard().SetContent(line)
					t.OnPasteFn(line)
				},
			}, nil, textObect,
		))
	} else {
		t.list.Add(textObect)
	}
	if len(t.list.Objects) > 100 {
		t.list.Objects = t.list.Objects[len(t.list.Objects)-30:]
	}

	t.list.Refresh()
	t.content.ScrollToBottom()
	// t.content.Refresh()
	// t.list.ScrollToBottom()
}

// func (t *Terminal) AppendStringToLastLine(s string) {
// 	(*t.data)[len(*t.data)-1] += s
// 	t.content.ScrollToBottom()
// 	t.list.Refresh()
// 	// t.content.Refresh()
// 	// t.list.ScrollToBottom()
// }

func New() *Terminal {
	datas := make([]string, 0)
	header := make([]string, 0)
	t := &Terminal{
		data:      &datas,
		header:    &header,
		OnPasteFn: func(s string) {},
	}
	// list := widget.NewList(
	// 	func() int {
	// 		return len(*t.data)
	// 	},
	// 	func() fyne.CanvasObject {
	// 		l := widget.NewLabel("empty")
	// 		// l.Wrapping = fyne.TextWrapBreak
	// 		return l
	// 	},
	// 	func(i widget.ListItemID, co fyne.CanvasObject) {
	// 		label := co.(*widget.Label)
	// 		label.SetText((*t.data)[i])
	// 	},
	// )
	t.list = container.NewVBox(
		widget.NewLabel("PTY:"),
	)
	t.list.Size()
	t.content = container.NewVScroll(t.list)
	// t.content = container.NewVScroll(list)
	return t
}

func (t *Terminal) GetContent(masterWindow fyne.Window) fyne.CanvasObject {
	t.masterWindow = masterWindow
	return t.content
}
