package omega

import "phoenixbuilder/minecraft/protocol/packet"

type UQInfoHolderEntry interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	UpdateFromPacket(packet packet.Packet)
}

type BotBasicInfoHolder interface {
	GetBotName() string
	GetBotRuntimeID() uint64
	GetBotUniqueID() int64
	GetBotIdentity() string
	UQInfoHolderEntry
}

type MicroUQHolder interface {
	GetBotBasicInfo() BotBasicInfoHolder
	UQInfoHolderEntry
}

// type PlayerUQsHolder interface {
// 	GetPlayerUQByName(name string) (uq PlayerUQReader, found bool)
// 	GetPlayerUQByUUID(ud uuid.UUID) (uq PlayerUQReader, found bool)
// 	GetBot() (botUQ PlayerUQReader)
// }

// type PlayerUQReader interface {
// 	IsBot() bool
// 	GetPlayerName() string
// }

// type PlayerUQ interface {
// 	PlayerUQReader
// }
