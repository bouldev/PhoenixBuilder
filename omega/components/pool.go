package components

import "phoenixbuilder/omega/defines"

type BasicComponent struct {
	cfg      *defines.ComponentConfig
	frame    defines.MainFrame
	ctrl     defines.GameControl
	listener defines.GameListener
}

func (bc *BasicComponent) Init(cfg *defines.ComponentConfig) {
	bc.cfg = cfg
}

func (bc *BasicComponent) Inject(frame defines.MainFrame) {
	bc.frame = frame
	bc.listener = frame.GetGameListener()
}

func (bc *BasicComponent) Activate() {
	bc.ctrl = bc.frame.GetGameControl()
}

func (bc *BasicComponent) Stop() error {
	return nil
}

func GetComponentsPool() map[string]func() defines.Component {
	return map[string]func() defines.Component{
		"Bonjour": func() defines.Component {
			return &Bonjour{BasicComponent: &BasicComponent{}}
		},
		"ChatLogger": func() defines.Component {
			return &ChatLogger{BasicComponent: &BasicComponent{}}
		},
		"Banner": func() defines.Component {
			return &Banner{BasicComponent: &BasicComponent{}}
		},
		"FeedBack": func() defines.Component {
			return &FeedBack{BasicComponent: &BasicComponent{}}
		},
		"Memo": func() defines.Component {
			return &Memo{BasicComponent: &BasicComponent{}}
		},
		"PlayerTP": func() defines.Component {
			return &PlayerTP{BasicComponent: &BasicComponent{}}
		},
		"BackToHQ": func() defines.Component {
			return &BackToHQ{BasicComponent: &BasicComponent{}}
		},
		"SetSpawnPoint": func() defines.Component {
			return &SetSpawnPoint{BasicComponent: &BasicComponent{}}
		},
		"Respawn": func() defines.Component {
			return &Respawn{BasicComponent: &BasicComponent{}}
		},
		"AboutMe": func() defines.Component {
			return &AboutMe{BasicComponent: &BasicComponent{}}
		},
	}
}
