package collaborate

import "strings"

const (
	INTERFACE_POSSIBLE_NAME = "GET_POSSIBLE_NAME"
)

type TYPE_NameEntry struct {
	CurrentName    string   `json:"current_name"`
	LastUpdateTime string   `json:"last_update_time"`
	NameRecord     []string `json:"history"`
}

type TYPE_PossibleNames struct {
	Entry            *TYPE_NameEntry
	SearchableString string
}

func (o *TYPE_PossibleNames) GenSearchAbleString() {
	o.SearchableString = o.Entry.CurrentName
	for i, h := range o.Entry.NameRecord {
		if i == 0 {
			continue
		}
		ss := strings.Split(h, ";")
		if len(ss) == 2 {
			o.SearchableString += " " + ss[0]
		}
	}
}

func (o *TYPE_PossibleNames) GenReadAbleStringPair() (currentName string, historyHint string) {
	historyNames := []string{}
	for i, h := range o.Entry.NameRecord {
		if i == 0 {
			continue
		}
		if i == 3 {
			break
		}
		ss := strings.Split(h, ";")
		if len(ss) == 2 {
			historyNames = append(historyNames, ss[0])
		}
	}
	if len(historyNames) == 0 {
		return o.Entry.CurrentName, ""
	} else {
		return o.Entry.CurrentName, "(曾用名: " + strings.Join(historyNames, ", ") + " )"
	}
}

type FUNC_GetPossibleName func(name string, maxC int) []*TYPE_PossibleNames
