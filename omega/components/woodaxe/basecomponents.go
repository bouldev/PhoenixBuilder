package woodaxe

import (
	"phoenixbuilder/omega/defines"
)

type BasicComponent struct {
	Config   *defines.ComponentConfig
	Frame    defines.MainFrame
	Ctrl     defines.GameControl
	Listener defines.GameListener
}

func (bc *BasicComponent) Init(cfg *defines.ComponentConfig) {
	bc.Config = cfg
}

func (bc *BasicComponent) Inject(frame defines.MainFrame) {
	bc.Frame = frame
	bc.Listener = frame.GetGameListener()
}

func (bc *BasicComponent) Activate() {
	bc.Ctrl = bc.Frame.GetGameControl()
}

func (bc *BasicComponent) Stop() error {
	return nil
}

func (bc *BasicComponent) Signal(signal int) error {
	return nil
}
