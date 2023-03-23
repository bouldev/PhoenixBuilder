package uqholder

import "fastbuilder-core/lib/minecraft/neomega/omega"

func init() {
	if false {
		func(uq omega.PlayerUQ) {}(&PlayerUQ{})
	}
}

type PlayerUQ struct {
}

func (p PlayerUQ) IsBot() bool {
	//TODO implement me
	panic("implement me")
}

func (p PlayerUQ) GetPlayerName() string {
	//TODO implement me
	panic("implement me")
}
