package mainframe

import (
	"fmt"
	"phoenixbuilder/omega/defines"
)

type Box struct {
	*Omega
	ComponentName  string
	backendDisplay defines.LineDst
	NameSpace      string
}

func NewBox(o *Omega, Name string, nameSpace string) *Box {
	b := &Box{
		Omega:         o,
		ComponentName: Name,
		backendDisplay: &BackEndLogger{loggers: []defines.LineDst{
			o.GetBackendDisplay(),
			o.GetLogger("component[" + Name + "]backend.log"),
		}},
		NameSpace: nameSpace,
	}
	return b
}

func (b *Box) GetContext(key string) (entry interface{}, found bool) {
	if entry := b.GetGlobalContext(b.NameSpace + key); entry != nil {
		return entry, true
	} else if b.GetGlobalContext(key); entry != nil {
		return entry, true
	}
	return nil, false
}

func (b *Box) SetContext(key string, entry interface{}) {
	b.SetGlobalContext(b.NameSpace+key, entry)
}

func (b *Box) GetBackendDisplay() defines.LineDst {
	return b.backendDisplay
}

func (b *Box) FatalError(err string) {
	b.backendDisplay.Write(fmt.Sprintf("%v Trigger an Fetal Error %v", b.ComponentName, err))
	b.Omega.Stop()
}
