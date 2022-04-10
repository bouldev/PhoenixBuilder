//go:build !fbconn

package uqHolder

import "phoenixbuilder/minecraft"

func (uq *UQHolder) UpdateFromConn(conn *minecraft.Conn) {
	gd := conn.GameData()
	uq.BotUniqueID = gd.EntityUniqueID
	uq.ConnectTime = gd.ConnectTime
	uq.WorldName = gd.WorldName
	uq.WorldGameMode = gd.WorldGameMode
	uq.WorldDifficulty = uint32(gd.Difficulty)
	uq.OnConnectWoldSpawnPosition = gd.WorldSpawn
	cd := conn.ClientData()
	uq.BotRandomID = cd.ClientRandomID
}
