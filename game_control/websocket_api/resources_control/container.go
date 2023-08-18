package ResourcesControl

import (
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/game_control/websocket_api/interfaces"
	"phoenixbuilder/minecraft/protocol/packet"
)

/*
Func List

func (*container).AwaitChangesAfterSendingPacket()
func (*container).AwaitChangesBeforeSendingPacket()
func (*container).GetContainerClosingData() *packet.ContainerClose
func (*container).GetContainerOpeningData() *packet.ContainerOpen
*/

type Container_ACFSP struct{}
type Container_ACFSP_R struct{}

func (c *Container_ACFSP) AutoMarshal(io interfaces.IO) {}

func (c *Container_ACFSP_R) AutoMarshal(io interfaces.IO) {}

func (c *Container_ACFSP) Run(env *GameInterface.GameInterface) interfaces.Return {
	env.Resources.Container.AwaitChangesAfterSendingPacket()
	return &Container_ACFSP_R{}
}

type Container_ACBSP struct{}
type Container_ACBSP_R struct{}

func (c *Container_ACBSP) AutoMarshal(io interfaces.IO) {}

func (c *Container_ACBSP_R) AutoMarshal(io interfaces.IO) {}

func (c *Container_ACBSP) Run(env *GameInterface.GameInterface) interfaces.Return {
	env.Resources.Container.AwaitChangesBeforeSendingPacket()
	return &Container_ACBSP_R{}
}

type Container_GCCD struct{}
type Container_GCCD_R struct {
	PacketContainerClose packet.ContainerClose `json:"packet_container_close"`
	ReturnValueIsNil     bool                  `json:"return_value_is_nil"`
}

func (c *Container_GCCD) AutoMarshal(io interfaces.IO) {}

func (c *Container_GCCD_R) AutoMarshal(io interfaces.IO) {
	io.Uint8(&c.PacketContainerClose.WindowID)
	io.Bool(&c.PacketContainerClose.ServerSide)
	io.Bool(&c.ReturnValueIsNil)
}

func (c *Container_GCCD) Run(env *GameInterface.GameInterface) interfaces.Return {
	resp := env.Resources.Container.GetContainerClosingData()
	if resp == nil {
		return &Container_GCCD_R{
			PacketContainerClose: packet.ContainerClose{},
			ReturnValueIsNil:     true,
		}
	}
	return &Container_GCCD_R{
		PacketContainerClose: *resp,
		ReturnValueIsNil:     false,
	}
}

type Container_GCOD struct{}
type Container_GCOD_R struct {
	PacketContainerOpen packet.ContainerOpen `json:"packet_container_open"`
	ReturnValueIsNil    bool                 `json:"return_value_is_nil"`
}

func (c *Container_GCOD) AutoMarshal(io interfaces.IO) {}

func (c *Container_GCOD_R) AutoMarshal(io interfaces.IO) {
	io.Uint8(&c.PacketContainerOpen.WindowID)
	io.Uint8(&c.PacketContainerOpen.ContainerType)
	io.Int32(&c.PacketContainerOpen.ContainerPosition[0])
	io.Int32(&c.PacketContainerOpen.ContainerPosition[1])
	io.Int32(&c.PacketContainerOpen.ContainerPosition[2])
	io.Int64(&c.PacketContainerOpen.ContainerEntityUniqueID)
	io.Bool(&c.ReturnValueIsNil)
}

func (c *Container_GCOD) Run(env *GameInterface.GameInterface) interfaces.Return {
	resp := env.Resources.Container.GetContainerOpeningData()
	if resp == nil {
		return &Container_GCOD_R{
			PacketContainerOpen: packet.ContainerOpen{},
			ReturnValueIsNil:    true,
		}
	}
	return &Container_GCOD_R{
		PacketContainerOpen: *resp,
		ReturnValueIsNil:    false,
	}
}
