package ResourcesControlCenter

import "phoenixbuilder/minecraft/protocol/packet"

// 提交结构请求
func (m *mcstructure) WriteRequest() {
	m.resp = make(chan packet.StructureTemplateDataResponse, 1)
}

// 向结构请求写入返回值 resp 。
// 属于私有实现。
func (m *mcstructure) writeResponce(data packet.StructureTemplateDataResponse) {
	m.resp <- data
	close(m.resp)
}

// 从管道读取结构请求的返回值
func (m *mcstructure) LoadResponce() packet.StructureTemplateDataResponse {
	return <-m.resp
}
