package py_rpc

// Return a pool/map that contains
// all the type of PyRpc packet
func Pool() map[string]PyRpc {
	return map[string]PyRpc{
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
