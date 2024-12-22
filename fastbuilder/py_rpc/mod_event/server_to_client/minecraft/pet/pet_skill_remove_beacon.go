package pet

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

// 用于表示宠物的临时信标的消失
type PetSkillRemoveBeacon struct {
	PetID string `json:"petId"` // e.g. "-455266532512"
}

// Return the event name of p
func (p *PetSkillRemoveBeacon) EventName() string {
	return "pet_skill_remove_beacon"
}

// Convert p to go object which only contains go-built-in types
func (p *PetSkillRemoveBeacon) MakeGo() (res any) {
	return map[string]any{"petId": p.PetID}
}

// Sync data to p from obj
func (p *PetSkillRemoveBeacon) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	pet_id, success := object["petId"].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["petId"] to string; object["petId"] = %#v`, object["petId"])
	}
	*p = PetSkillRemoveBeacon{PetID: pet_id}
	// get and sync data
	return nil
	// return
}
