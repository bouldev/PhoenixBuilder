package lang

import (
	_ "embed"
	"encoding/json"
	ofmt "fmt"
)

//go:embed lang.json
var LangJson []byte

type DictsRecord struct {
	Languages []string            `json:"languages"`
	Dicts     map[string][]string `json:"dicts"`
}

var allDicts = map[string]*map[string]string{}
var dict = &map[string]string{}

func UpdateDict(records *DictsRecord) {
	for _, lang := range records.Languages {
		if allDicts[lang] == nil {
			allDicts[lang] = &map[string]string{}
		}
	}
	for key, values := range records.Dicts {
		for i, value := range values {
			(*allDicts[records.Languages[i]])[key] = value
		}
	}
}

func AllLanguageName() []string {
	names := make([]string, len(allDicts))
	i := 0
	for name := range allDicts {
		names[i] = name
		if (*(allDicts[name]))["lang.name"] != "" {
			names[i] = (*(allDicts[name]))["lang.name"]
		}
		i++
	}
	return names
}

func SetLanguage(lang string) (err error) {
	if allDicts[lang] != nil {
		dict = allDicts[lang]
	} else {
		for name := range allDicts {
			if (*(allDicts[name]))["lang.name"] == lang {
				dict = allDicts[name]
				return
			}
		}
	}
	return Fmt.Errorf("language %v not found", lang)
}

func init() {
	records := &DictsRecord{}
	err := json.Unmarshal(LangJson, &records)
	if err != nil {
		panic(err)
	}
	UpdateDict(records)
	dict = allDicts["simplified_chinese"]
}

func Trans(frags []any) []any {
	if len(*dict) == 0 {
		return frags
	}
	translatedFrags := make([]any, len(frags))
	copy(translatedFrags, frags)

	for i, frag := range frags {
		if sfrag, ok := frag.(string); ok {
			if translatedFrag, found := (*dict)[sfrag]; found && translatedFrag != "" {
				translatedFrags[i] = translatedFrag
			}
		}
	}
	return translatedFrags
}

func Transf(fmt string, frags []any) (string, []any) {
	if len(*dict) == 0 {
		return fmt, frags
	}
	translatedFrags := Trans(frags)
	tfmt := fmt
	if translatedFmt, found := (*dict)[fmt]; found && translatedFmt != "" {
		tfmt = translatedFmt
	}
	return tfmt, translatedFrags
}

type BasicPrinter struct{}

func (p *BasicPrinter) Println(frags ...any) {
	ofmt.Println(Trans(frags)...)
}

func (p *BasicPrinter) Sprintf(fmts string, frags ...any) string {
	tfmts, tfrags := Transf(fmts, frags)
	return ofmt.Sprintf(tfmts, tfrags...)
}

func (p *BasicPrinter) Printf(fmts string, frags ...any) {
	tfmts, tfrags := Transf(fmts, frags)
	ofmt.Printf(tfmts, tfrags...)
}

func (p *BasicPrinter) Errorf(fmts string, frags ...any) error {
	tfmts, tfrags := Transf(fmts, frags)
	return ofmt.Errorf(tfmts, tfrags...)
}

func (p *BasicPrinter) PrintError(err error) {
	if err != nil {
		p.Println(ofmt.Sprint(err))
	}
}

var Fmt = &BasicPrinter{}

func Printf(fmts string, frags ...any) {
	Fmt.Printf(fmts, frags...)
}

func Println(frags ...any) {
	Fmt.Println(frags...)
}

func Sprintf(fmts string, frags ...any) string {
	return Fmt.Sprintf(fmts, frags...)
}

func Errorf(fmts string, frags ...any) error {
	return Fmt.Errorf(fmts, frags...)
}

func PrintError(err error) {
	Fmt.PrintError(err)
}
