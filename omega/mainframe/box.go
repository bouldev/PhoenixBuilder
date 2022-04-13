package mainframe

import (
	"fmt"
	"phoenixbuilder/omega/defines"
)

type Box struct {
	*Omega
	ComponentName  string
	backendDisplay defines.LineDst
}

func NewBox(o *Omega, Name string) *Box {
	b := &Box{
		Omega:         o,
		ComponentName: Name,
		backendDisplay: &BackEndLogger{loggers: []defines.LineDst{
			o.GetBackendDisplay(),
			o.GetLogger("component[" + Name + "]backend.log"),
		}},
	}
	return b
}

func (b *Box) GetBackendDisplay() defines.LineDst {
	return b.backendDisplay
}

func (b *Box) FatalError(err string) {
	b.backendDisplay.Write(fmt.Sprintf("%v Trigger an Fetal Error %v", b.ComponentName, err))
	b.Omega.Stop()
}
