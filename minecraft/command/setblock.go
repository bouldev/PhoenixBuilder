package command

import (
	"fmt"
	"gophertunnel/minecraft/mctype"
)


func SetBlockRequest(module mctype.Module, config mctype.MainConfig) string {
	Block := module.Block
	Point := module.Point
	Method := config.Method
	return fmt.Sprintf("setblock %v %v %v %v %v %v",Point.X, Point.Y, Point.Z, Block.Name, Block.Data, Method)
}

type SetBlock struct {
	Position mctype.Position `json:"position"`
	StatusMessage string `json:"statusMessage"`
	StatusCode int `json:"statusCode"`
}

