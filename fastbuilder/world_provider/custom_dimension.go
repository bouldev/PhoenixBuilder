package world_provider

import (
	"time"
)

type CustomDimension struct {}

//func (CustomDimension) Range() cube.Range { return cube.Range{0,256} }
func (CustomDimension) EncodeDimension() int { return 0 }
func (CustomDimension) WaterEvaporates() bool { return false }
func (CustomDimension) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
func (CustomDimension) WeatherCycle() bool { return true }
func (CustomDimension) TimeCycle() bool { return true }
func (CustomDimension) String() string { return "Overworld" }

// This file isn't used currently