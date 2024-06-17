package describe

import "strings"

type BaseWithNameSpace [2]string

func (b BaseWithNameSpace) BaseName() string {
	return b[1]
}

func (b BaseWithNameSpace) LongName() string {
	return b[0] + ":" + b[1]
}

func (b BaseWithNameSpace) NameSpace() string {
	return b[0]
}

func BlockNameForSearch(name string) BaseWithNameSpace {
	name = strings.TrimSpace(name)
	frags := strings.Split(name, ":")
	if len(frags) == 1 {
		return BaseWithNameSpace{"minecraft", name}
	} else {
		if len(frags) != 2 {
			panic(frags)
		}
		return BaseWithNameSpace{frags[0], frags[1]}
	}
}
