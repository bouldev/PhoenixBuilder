package list_terminal

import (
	"fyne.io/fyne/v2/dialog"
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
	data         []string
	dataCanCopy  []bool
	header       *widget.Entry
	list         *widget.List
	OnPasteFn    func(s string)
}

func (t *Terminal) AppendNewLine(line string, canCopy bool) {
	//t.data=append(t.data,line)
	//t.dataCanCopy=append(t.dataCanCopy,canCopy)
	//offset := 0
	//width := 0
	//for {
	//	_, len := utf8.DecodeRuneInString(line[offset:])
	//	offset += len
	//	if len != 0 {
	//		// maybe chinese
	//		width += 1
	//		if len > 1 {
	//			width += 1
	//		}
	//	} else {
	//		break
	//	}
	//}
	//*t.data = append(*t.data, line)
	//var textObect fyne.CanvasObject
	//// TODO this should automatically be computed
	//if width < 45 {
	//	l := widget.NewLabel(line)
	//	l.Wrapping = fyne.TextTruncate
	//	textObect = l
	//} else {
	//	l := widget.NewLabel(line)
	//	l.Wrapping = fyne.TextTruncate
	//	fulldescription := widget.NewMultiLineEntry()
	//	fulldescription.Wrapping = fyne.TextWrapBreak
	//	fulldescription.Hide()
	//	fulldescription.Text = line
	//	textObect = container.NewBorder(nil, fulldescription, &widget.Button{
	//		Icon:          theme.VisibilityIcon(),
	//		Importance:    widget.LowImportance,
	//		IconPlacement: widget.ButtonIconTrailingText,
	//		OnTapped: func() {
	//			if fulldescription.Hidden {
	//				fulldescription.Show()
	//			} else {
	//				fulldescription.Hide()
	//			}
	//		},
	//	}, nil, l)
	//}
	//if canCopy || width >= 45 {
	//	t.list.Add(container.NewBorder(
	//		nil, nil, //widget.NewLabel(fmt.Sprintf("%d", len(*t.data)+1)),
	//		&widget.Button{
	//			Icon:          theme.ContentCopyIcon(),
	//			Importance:    widget.LowImportance,
	//			IconPlacement: widget.ButtonIconLeadingText,
	//			OnTapped: func() {
	//				// clipboard.Write(clipboard.FmtText, []byte(line))
	//				// fmt.Println(line)
	//				t.masterWindow.Clipboard().SetContent(line)
	//				t.OnPasteFn(line)
	//			},
	//		}, nil, textObect,
	//	))
	//} else {
	//	t.list.Add(textObect)
	//}
	//if len(t.list.Objects) > 100 {
	//	t.list.Objects = t.list.Objects[len(t.list.Objects)-30:]
	//}
	//
	//t.list.Refresh()
	t.data = append(t.data, line)
	t.dataCanCopy = append(t.dataCanCopy, canCopy)
	t.list.ScrollToBottom()
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
	t := &Terminal{
		data:        make([]string, 0),
		dataCanCopy: make([]bool, 0),
		OnPasteFn:   func(s string) {},
	}
	t.header = widget.NewMultiLineEntry()
	t.header.Hide()
	t.list = widget.NewList(func() int {
		return len(t.data)
	}, func() fyne.CanvasObject {
		dataLabel := widget.NewLabel("cmd")
		dataLabel.Wrapping = fyne.TextTruncate
		return container.NewBorder(nil, nil,
			container.NewHBox(
				&widget.Button{
					DisableableWidget: widget.DisableableWidget{},
					Icon:              theme.ContentCopyIcon(),
					Importance:        widget.LowImportance,
					OnTapped: func() {
						t.masterWindow.Clipboard().SetContent(dataLabel.Text)
						t.OnPasteFn(dataLabel.Text)
					},
				}, &widget.Button{
					DisableableWidget: widget.DisableableWidget{},
					Icon:              theme.VisibilityIcon(),
					Importance:        widget.LowImportance,
					OnTapped: func() {
						dialog.ShowCustom("详细信息", "好的",
							// since it is in a dialog, we need to force it's min size
							// it's suck!
							container.NewBorder(widget.NewLabel(""+
								"                                         "),
								&widget.Button{
									DisableableWidget: widget.DisableableWidget{},
									Icon:              theme.ContentCopyIcon(),
									Importance:        widget.HighImportance,
									Text:              "复制信息",
									OnTapped: func() {
										t.masterWindow.Clipboard().SetContent(dataLabel.Text)
										t.OnPasteFn(dataLabel.Text)
									},
								}, nil, nil,
								&widget.Label{
									Text:      dataLabel.Text,
									Wrapping:  fyne.TextWrapWord,
								},
							), t.masterWindow)
						//t.header.SetText(dataLabel.Text)
						//if t.header.Hidden {
						//	t.header.Show()
						//}else{
						//	if t.header.Text==dataLabel.Text{
						//		t.header.Hide()
						//	}
						//}
					},
				},
			), nil, dataLabel,
		)
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		object.(*fyne.Container).Objects[0].(*widget.Label).SetText(t.data[id])
		////(*widget.Label).SetText(g.OperationLogs[id].time.Format("2006-01-02 15:04:05"))
		offset := 0
		width := 0
		for {
			_, len := utf8.DecodeRuneInString(t.data[id][offset:])
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
		if width < 45 {
			object.(*fyne.Container).Objects[1].(*fyne.Container).Objects[1].(*widget.Button).Hide()
		}
		if !(width > 45 || t.dataCanCopy[id]) {
			object.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Button).Hide()
		}

	})
	//t.list = container.NewVBox(
	//	widget.NewLabel("PTY:"),
	//)
	//t.list.Size()
	//t.content = container.NewVScroll(t.list)
	// t.content = container.NewVScroll(list)
	return t
}

func (t *Terminal) GetContent(masterWindow fyne.Window) fyne.CanvasObject {
	t.masterWindow = masterWindow
	return container.NewBorder(
		//&widget.Button{
		//	Text:              "insert",
		//	Icon:              theme.CancelIcon(),
		//	Importance:        widget.MediumImportance,
		//	OnTapped: func() {
		//		g.OperationLogs=append(g.OperationLogs,&LogEntry{time.Now(),"insert"})
		//		LogList.Refresh()
		//	},},
		t.header, nil, nil, nil, t.list)
}
