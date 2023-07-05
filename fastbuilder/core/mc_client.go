package core

import (
	"context"
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/py_rpc"
)

func InitializeMinecraftConnection(ctx context.Context, authentication minecraft.Authenticator, options ...Option) (conn *minecraft.Conn, err error) {
	if checkOption(options, OptionDebug) {
		conn = &minecraft.Conn{
			DebugMode: true,
		}
	} else {
		dialer := minecraft.Dialer{
			Authenticator: authentication,
		}
		conn, err = dialer.DialContext(ctx, "raknet")
	}
	if err != nil {
		return
	}
	conn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	runtimeid := fmt.Sprintf("%d", conn.GameData().EntityUniqueID)
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"SyncUsingMod",
			[]interface{} {},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"SyncVipSkinUuid",
			[]interface{} {nil},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"ClientLoadAddonsFinishedFromGac",
			[]interface{} {},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"ModEventC2S",
			[]interface{} {
				"Minecraft",
				"preset",
				"GetLoadedInstances",
				map[string]interface{} {
					"playerId": runtimeid,
				},
			},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"arenaGamePlayerFinishLoad",
			[]interface{} {},
			nil,
		}),
	})
	conn.WritePacket(&packet.PyRpc{
		Value: py_rpc.FromGo([]interface{} {
			"ModEventC2S",
			[]interface{} {
				"Minecraft",
				"vipEventSystem",
				"PlayerUiInit",
				runtimeid,
			},
			nil,
		}),
	})
	return
}
