package mux_pumper

import (
	"phoenixbuilder/fastbuilder/lib/utils/sync_wrapper"

	"github.com/google/uuid"
)

type InputPumperMux struct {
	pumpers *sync_wrapper.SyncMap[chan string]
}

func NewInputPumperMux() *InputPumperMux {
	return &InputPumperMux{
		pumpers: sync_wrapper.NewInstanceMap[chan string](),
	}
}

func (i *InputPumperMux) PumpInput(input string) {
	currentPumper := i.pumpers
	i.pumpers = sync_wrapper.NewInstanceMap[chan string]()
	currentPumper.Iter(func(k string, listener chan string) (continueInter bool) {
		select {
		case listener <- input:
		default:
		}
		return true
	})
}

func (i *InputPumperMux) NewListener() chan string {
	listener := make(chan string)
	i.pumpers.Set(uuid.New().String(), listener)
	return listener
}
