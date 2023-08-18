package ResourcesControl

import (
	GameInterface "phoenixbuilder/game_control/game_interface"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/game_control/websocket_api/interfaces"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

/*
Func List

func (*commandRequestWithResponse).LoadResponseAndDelete(key uuid.UUID) ResourcesControl.CommandRespond
func (*commandRequestWithResponse).WriteRequest(key uuid.UUID, options CommandRequestOptions) error
*/

type Command_LRAD struct {
	Key uuid.UUID `json:"key"`
}

type Command_LRAD_R struct {
	Respond   packet.CommandOutput `json:"respond"`
	Error     string               `json:"error"`
	ErrorType uint8                `json:"error_type"`
}

func (c *Command_LRAD) AutoMarshal(io interfaces.IO) {
	io.UUID(&c.Key)
}

func (c *Command_LRAD_R) AutoMarshal(io interfaces.IO) {
	io.Uint32(&c.Respond.CommandOrigin.Origin)
	io.UUID(&c.Respond.CommandOrigin.UUID)
	io.String(&c.Respond.CommandOrigin.RequestID)
	io.Int64(&c.Respond.CommandOrigin.PlayerUniqueID)
	io.Uint8(&c.Respond.OutputType)
	io.Uint32(&c.Respond.SuccessCount)
	io.CommandOutputMessageSlice(&c.Respond.OutputMessages)
	io.String(&c.Respond.DataSet)
	io.String(&c.Error)
	io.Uint8(&c.ErrorType)
}

func (c *Command_LRAD) Run(env *GameInterface.GameInterface) interfaces.Return {
	resp := env.Resources.Command.LoadResponseAndDelete(c.Key)
	return &Command_LRAD_R{
		Respond:   resp.Respond,
		Error:     resp.Error.Error(),
		ErrorType: resp.ErrorType,
	}
}

type Command_WR struct {
	Key     uuid.UUID                              `json:"key"`
	Options ResourcesControl.CommandRequestOptions `json:"options"`
}

type Command_WR_R struct {
	Error string `json:"error"`
}

func (c *Command_WR) AutoMarshal(io interfaces.IO) {
	io.UUID(&c.Key)
	io.Int64((*int64)(&c.Options.TimeOut))
}

func (c *Command_WR_R) AutoMarshal(io interfaces.IO) {
	io.String(&c.Error)
}

func (c *Command_WR) Run(env *GameInterface.GameInterface) interfaces.Return {
	resp := env.Resources.Command.LoadResponseAndDelete(c.Key)
	return &Command_WR_R{
		Error: resp.Error.Error(),
	}
}
