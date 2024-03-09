package py_rpc_content

// Return a pool/map that contains
// all the content of PyRpc packet
func Pool() map[string]PyRpcContent {
	return map[string]PyRpcContent{
		"arenaGamePlayerFinishLoad":       &ArenaGamePlayerFinishLoad{},
		"ClientLoadAddonsFinishedFromGac": &ClientLoadAddonsFinishedFromGac{},
		"GetMCPCheckNum":                  &GetMCPCheckNum{},
		"S2CHeartBeat":                    &HeartBeat{Type: ServerToClientHeartBeat},
		"C2SHeartBeat":                    &HeartBeat{Type: ClientToServerHeartBeat},
		"ModEventS2C":                     &ModEvent{Type: ModEventServerToClient},
		"ModEventC2S":                     &ModEvent{Type: ModEventClientToServer},
		"SetMCPCheckNum":                  &SetMCPCheckNum{},
		"SetOwnerId":                      &SetOwnerId{},
		"GetStartType":                    &StartType{Type: StartTypeRequest},
		"SetStartType":                    &StartType{Type: StartTypeResponse},
		"SyncUsingMod":                    &SyncUsingMod{},
		"SyncVipSkinUuid":                 &SyncVipSkinUUID{},
	}
}
