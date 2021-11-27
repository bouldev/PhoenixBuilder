package packet

import (
	"bytes"
	"phoenixbuilder/minecraft/protocol"
)


type NeteaseJson struct {
	Content string
}

// ID ...
func (*NeteaseJson) ID() uint32 {
	return IDNeteaseJson
}

// Marshal ...
func (pk *NeteaseJson) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.Content)
}

// Unmarshal ...
func (pk *NeteaseJson) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.Content),
	)
}
