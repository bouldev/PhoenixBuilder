package item

import (
	"phoenixbuilder/dragonfly/server/entity/effect"
	"phoenixbuilder/dragonfly/server/world"
	"math/rand"
	"time"
)

// RottenFlesh is a food item that can be eaten by the player, at the high risk of inflicting Hunger.
type RottenFlesh struct {
	defaultFood
}

// Consume ...
func (RottenFlesh) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 0.8)
	if rand.Float64() < 0.8 {
		c.AddEffect(effect.New(effect.Hunger{}, 1, 30*time.Second))
	}
	return Stack{}
}

// EncodeItem ...
func (RottenFlesh) EncodeItem() (name string, meta int16) {
	return "minecraft:rotten_flesh", 0
}
