package preset

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

import "fmt"

// Used on initialize the
// Minecraft connection
type GetLoadedInstances struct {
	PlayerRuntimeID string `json:"playerId"`
}

// Return the event name of g
func (g *GetLoadedInstances) EventName() string {
	return "GetLoadedInstances"
}

// Convert g to go object which only contains go-built-in types
func (g *GetLoadedInstances) MakeGo() (res any) {
	return map[string]any{"playerId": g.PlayerRuntimeID}
}

// Sync data to g from obj
func (g *GetLoadedInstances) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	player_runtime_id, success := object["playerId"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["playerId"] to string; object["playerId"] = %#v`, object["playerId"])
	}
	g.PlayerRuntimeID = player_runtime_id
	// get and sync data
	return nil
	// return
}
