package structure

import (
	"errors"
	"phoenixbuilder/mirror/define"
)

type IOBlock struct {
	Pos      define.CubePos
	RTID     uint32
	NBT      map[string]interface{}
	Hit      bool
	Expand16 bool
}

var ErrImportFormateNotSupport = errors.New("formate not support")
