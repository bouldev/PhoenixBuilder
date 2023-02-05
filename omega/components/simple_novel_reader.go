package components

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

type SimpleNovelReaderPerPlayerPerBookData struct {
	CurrentPosition int `json:"当前进度"`
	// LineOffset      int            "行内偏移"
	BookMarks []int `json:"书签"`
}

type SimpleNovelReaderData struct {
	PlayerData map[string]map[string]*SimpleNovelReaderPerPlayerPerBookData `json:"玩家数据"`
}

type SimpleNovelReaderOptions struct {
	Head              string `json:"书目页面题头"`
	MaxBooksPerPage   int    `json:"每页最多显示的书目数量"`
	BookEntryFmt      string `json:"书目渲染格式"`
	BooksListEnd      string `json:"书目清单结尾"`
	BooksListStart    string `json:"书目清单开头"`
	BooksListNextPage string `json:"书目清单下一页"`
	MaxWordsPerPage   int    `json:"每页最少显示字符数"`
	BookEnd           string `json:"书内最后一页"`
	BookStart         string `json:"书内第一页"`
	BookNextPage      string `json:"书内"`
	UseActionBar      bool   `json:"使用ActionBar显示内容"`
	HintOnNoBookMark  string `json:"没有书签时的提示"`
	BookMarkFmt       string `json:"书签显示格式"`
	HintOnBookMark    string `json:"有书签时提示"`
	BookMarkOperate   string `json:"书签操作提示"`
}

type SimpleNovelReader struct {
	*defines.BasicComponent
	fileChanged bool
	fileData    *SimpleNovelReaderData
	FileName    string                   `json:"进度记录和书签文件"`
	NovelDirs   string                   `json:"小说文件夹"`
	Triggers    []string                 `json:"触发词"`
	Usage       string                   `json:"提示信息"`
	Render      SimpleNovelReaderOptions `json:"显示选项"`
	bookOrder   []string
	books       map[string][]string
}

func (o *SimpleNovelReader) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
}

// func alterWords(in string) (out string) {
// 	// r := strconv.QuoteToASCII(in)
// 	// if len(r) > 2 {
// 	// 	return r[1 : len(r)-1]
// 	// }

// 	// return ""
// 	// rs := []rune{}
// 	// split := []rune("\x00")
// 	// rrs := []rune(in)
// 	// for i := 0; i < len(rrs); i++ {
// 	// 	rs = append(rs, rrs[i])
// 	// 	rs = append(rs, split...)
// 	// }
// 	// return string(rs)
// 	out = ""
// 	for _, i := range in {
// 		out += i + []rune("\x00")[0]
// 	}
// 	return out
// }

func (o *SimpleNovelReader) bookMark(mark string, pk defines.PlayerKit, book []string, bookData *SimpleNovelReaderPerPlayerPerBookData, tailLine string) {
	availableActions := map[string]func(){}
	availableActions["取消"] = func() {
		go o.showPage(pk, book, bookData, tailLine)
		return
	}
	availableActions["m"] = func() {
		bookData.BookMarks = append(bookData.BookMarks, bookData.CurrentPosition)
		pk.Say("书签已保存")
		o.fileChanged = true
		go o.showPage(pk, book, bookData, tailLine)
		return
	}
	if len(bookData.BookMarks) == 0 {
		pk.Say(o.Render.HintOnNoBookMark)
	} else {
		for _i, _p := range bookData.BookMarks {
			i, p := _i, _p
			I := i + 1
			if p >= len(book) {
				p = len(book) - 1
			}
			if p < 0 {
				p = 0
			}
			l := utils.FormatByReplacingOccurrences(o.Render.BookMarkFmt, map[string]interface{}{
				"[I]":    I,
				"[当前位置]": p,
				"[总长度]":  len(book),
				"[内容预览]": book[p],
			})
			pk.RawSay("§f\n" + l)
			availableActions[fmt.Sprintf("%v", I)] = func() {
				pk.Say(o.Render.BookMarkOperate)
				pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
					if len(chat.Msg) < 1 {
						pk.Say("无法理解的输入")
						go o.bookMark(mark, pk, book, bookData, tailLine)
						return true
					}
					m := chat.Msg[0]
					if m == "d" {
						newBookMarks := make([]int, 0, len(bookData.BookMarks)-1)
						for _i, _p := range bookData.BookMarks {
							if _i != i {
								newBookMarks = append(newBookMarks, _p)
							}
						}
						bookData.BookMarks = newBookMarks
						o.fileChanged = true
						go o.bookMark(mark, pk, book, bookData, tailLine)
						return
					} else {
						bookData.CurrentPosition = p
						go o.showPage(pk, book, bookData, "")
						return true
					}
					return true
				})
			}
		}
		pk.Say(o.Render.HintOnBookMark)
	}
	pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) < 1 {
			pk.Say("无法理解的输入")
			go o.bookMark(mark, pk, book, bookData, tailLine)
			return true
		}
		m := chat.Msg[0]
		if fn, found := availableActions[m]; found {
			fn()
		} else {
			pk.Say("无法理解的输入")
			go o.bookMark(mark, pk, book, bookData, tailLine)
			return true
		}
		return true
	})
}

func (o *SimpleNovelReader) showPage(pk defines.PlayerKit, book []string, bookData *SimpleNovelReaderPerPlayerPerBookData, tailLine string) {
	p := bookData.CurrentPosition
	content := []string{}
	finished := false
	start := false

	if p >= len(book) {
		content = append(content, "全书完")
		finished = true
		p = len(book)
	} else if p < 0 {
		p = 0
		start = true
	}
	if bookData.CurrentPosition != p {
		bookData.CurrentPosition = p
		o.fileChanged = true
	}

	previousPos := p
	nextPos := p
	counter := 0
	firstLine := ""
	for _i := p; _i < len(book); _i++ {
		firstLine = book[_i]
		nextPos = _i
		content = append(content, book[_i])
		l := []rune(book[_i])
		counter += len(l)
		if counter > o.Render.MaxWordsPerPage {
			break
		}
	}
	nextPos++
	for _i := p - 1; _i >= 0; _i-- {
		previousPos = _i
		l := []rune(book[_i])
		counter += len(l)
		if counter > o.Render.MaxWordsPerPage {
			break
		}
	}
	infoToDisplay := tailLine
	for _, l := range content {
		infoToDisplay += l + "\n"
	}
	tail := ""
	if start {
		tail = o.Render.BookStart
	} else if finished {
		tail = o.Render.BookEnd
	} else {
		tail = o.Render.BookNextPage
	}

	s := 0
	e := 0
	lastB := 0
	for e = range infoToDisplay {
		if infoToDisplay[e] == '\n' {
			lastB = e
		}
		if (e-s) > 300 && (infoToDisplay[e] == '\n' || (e-lastB) > 400) {
			// pterm.Info.Println(e-s, strings.TrimSpace(infoToDisplay[s:e]))
			pk.RawSay("§r\n" + strings.TrimSpace(infoToDisplay[s:e]))
			s = e
		}
	}
	nextTailLine := infoToDisplay[s:]
	// pterm.Warning.Println(nextTailLine)

	pk.Say(utils.FormatByReplacingOccurrences(tail, map[string]interface{}{
		"[当前位置]": bookData.CurrentPosition,
		"[总长度]":  len(book),
	}))
	// if o.Render.UseActionBar {
	// 	pk.ActionBar(infoToDisplay)
	// } else {

	// }
	pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) < 1 {
			pk.Say("无法理解的输入")
			go o.showPage(pk, book, bookData, tailLine)
			return false
		}
		m := chat.Msg[0]
		if m == "退出" {
			pk.Say("已保存")
			return true
		}
		if m == "p" && !start {
			bookData.CurrentPosition = previousPos
			go o.showPage(pk, book, bookData, "")
			return true
		}
		if m == "n" && !finished {
			bookData.CurrentPosition = nextPos
			go o.showPage(pk, book, bookData, nextTailLine)
			return true
		}
		if m == "m" {
			o.bookMark(firstLine, pk, book, bookData, nextTailLine)
			return true
		} else {
			pk.Say("无法理解的输入")
			go o.showPage(pk, book, bookData, tailLine)
			return false
		}
	})
}

func (o *SimpleNovelReader) readBook(chat *defines.GameChat, bookName string) {
	playerData := o.fileData.PlayerData[chat.Name]
	if playerData == nil {
		playerData = make(map[string]*SimpleNovelReaderPerPlayerPerBookData)
		o.fileData.PlayerData[chat.Name] = playerData
		o.fileChanged = true
	}
	bookData := playerData[bookName]
	if bookData == nil {
		o.fileData.PlayerData[chat.Name][bookName] = &SimpleNovelReaderPerPlayerPerBookData{
			CurrentPosition: 0,
			BookMarks:       []int{},
		}
		o.fileChanged = true
		bookData = playerData[bookName]
	}
	book := o.books[bookName]
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	o.showPage(pk, book, bookData, "")
}

func (o *SimpleNovelReader) onTrigger(chat *defines.GameChat, currentI int) {
	pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
	infoToDisplay := []string{}
	// pk.Say(o.Render.Head)
	infoToDisplay = append(infoToDisplay, o.Render.Head+"§r")
	i := currentI
	availableAction := map[string]string{}
	for i = currentI; i < len(o.bookOrder) && i < currentI+o.Render.MaxBooksPerPage; i++ {
		l := utils.FormatByReplacingOccurrences(o.Render.BookEntryFmt, map[string]interface{}{
			"[I]":  i + 1,
			"[书名]": o.bookOrder[i],
		})
		availableAction[fmt.Sprintf("%v", i+1)] = o.bookOrder[i]
		availableAction[o.bookOrder[i]] = o.bookOrder[i]
		infoToDisplay = append(infoToDisplay, l+"§r")
		// pk.Say(l)
	}
	if i == len(o.bookOrder) {
		infoToDisplay = append(infoToDisplay, o.Render.BooksListEnd+"§r")
		// pk.Say(o.Render.BooksListEnd)
	} else if currentI == 0 {
		infoToDisplay = append(infoToDisplay, o.Render.BooksListStart+"§r")
		// pk.Say(o.Render.BooksListStart)
	} else {
		infoToDisplay = append(infoToDisplay, o.Render.BooksListNextPage+"§r")
		// pk.Say(o.Render.BooksListNextPage)
	}
	pk.RawSay("->\n" + strings.Join(infoToDisplay, "\n"))
	pk.SetOnParamMsg(func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) < 1 {
			pk.Say("无法理解的输入")
			return true
		}
		m := chat.Msg[0]
		if m == "取消" {
			pk.Say("已取消")
			return true
		}
		if m == "p" && currentI != 0 {
			newI := currentI - o.Render.MaxBooksPerPage
			if newI < 0 {
				newI = 0
			}
			go o.onTrigger(chat, newI)
			return true
		}
		if m == "n" && i != len(o.bookOrder) {
			go o.onTrigger(chat, currentI+o.Render.MaxBooksPerPage)
			return true
		}
		if bookName, hasK := availableAction[m]; hasK {
			o.readBook(chat, bookName)
			return true
		} else {
			pk.Say("无法理解的输入")
			return true
		}
	})
	return
}

func (o *SimpleNovelReader) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.fileData = &SimpleNovelReaderData{
		PlayerData: make(map[string]map[string]*SimpleNovelReaderPerPlayerPerBookData),
	}
	o.Frame.GetJsonData(o.FileName, &o.fileData)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: func(chat *defines.GameChat) (stop bool) {
			o.onTrigger(chat, 0)
			return true
		},
	})
	absDir := o.Frame.GetRelativeFileName(o.NovelDirs)
	o.books = make(map[string][]string)
	o.bookOrder = make([]string, 0)
	if err := filepath.Walk(absDir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(path, ".txt") {
			if data, err := os.ReadFile(path); err != nil {
				return err
			} else {
				data, err = utils.AutoConvertTextToUtf8(data)
				if err != nil {
					pterm.Error.Println("无法自动修正编码格式")
					err = nil
				}
				bookName := strings.ReplaceAll(info.Name(), ".txt", "")
				o.bookOrder = append(o.bookOrder, bookName)
				cleanUpData := []string{}
				for _, l := range strings.Split(string(data), "\n") {
					if l == "" {
						continue
					} else {
						cleanUpData = append(cleanUpData, l)
					}
				}
				o.books[bookName] = cleanUpData
			}
		}
		return nil
	}); err != nil {
		panic(fmt.Errorf("读取小说时出现问题: %v", err))
	}
	if len(o.books) == 0 {
		os.MkdirAll(absDir, 0755)
		panic(fmt.Errorf("读取小说时出现问题: %v 下一本以 .txt 文件结尾的小说都没有", absDir))
	} else {
		pterm.Info.Printfln("共找到%v本小说", len(o.books))
	}
}

func (o *SimpleNovelReader) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChanged {
			o.fileChanged = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.fileData)
		}
	}
	return nil
}

func (o *SimpleNovelReader) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.fileData)
}
