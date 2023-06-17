package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 请求 request 代表的结构请求并获取与之对应的响应体。
// 当且仅当租赁服响应结构请求时本函数才会返回值。
//
// 请确保在使用此函数前占用了结构资源，否则这将导致程序惊慌
func (g *GlobalAPI) SendStructureRequestWithResponce(
	request *packet.StructureTemplateDataRequest,
) (packet.StructureTemplateDataResponse, error) {
	g.Resources.Structure.WriteRequest()
	// prepare
	err := g.WritePacket(request)
	if err != nil {
		return packet.StructureTemplateDataResponse{}, fmt.Errorf("SendStructureRequestWithResponce: %v", err)
	}
	// send packet
	return g.Resources.Structure.LoadResponce(), nil
	// load responce and return
}
