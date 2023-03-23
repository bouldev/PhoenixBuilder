package omega

import "github.com/google/uuid"

type BotBasicInfoHolder interface {
	GetBotName() string
	GetBotRuntimeID() uint64
	GetBotUniqueID() int64
}

type PlayerUQsHolder interface {
	GetPlayerUQByName(name string) (uq PlayerUQReader, found bool)
	GetPlayerUQByUUID(ud uuid.UUID) (uq PlayerUQReader, found bool)
	GetBot() (botUQ PlayerUQReader)
}

type PlayerUQReader interface {
	IsBot() bool
	GetPlayerName() string
}

type PlayerUQ interface {
	PlayerUQReader
}
