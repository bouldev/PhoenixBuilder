package py_rpc

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

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
