package command

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

import (
	"encoding/binary"
	"io"
	"phoenixbuilder/fastbuilder/types"
)

type PlaceRuntimeBlockWithChestDataAndUint32RuntimeID struct {
	BlockRuntimeID uint32
	ChestSlots     []types.ChestSlot
}

func (_ *PlaceRuntimeBlockWithChestDataAndUint32RuntimeID) ID() uint16 {
	return 38
}

func (_ *PlaceRuntimeBlockWithChestDataAndUint32RuntimeID) Name() string {
	return "PlaceRuntimeBlockWithChestDataAndUint32RuntimeIDCommand"
}

func (cmd *PlaceRuntimeBlockWithChestDataAndUint32RuntimeID) Marshal(writer io.Writer) error {
	uint16_buf := make([]byte, 2)
	uint32_buf := make([]byte, 4)
	binary.BigEndian.PutUint32(uint32_buf, cmd.BlockRuntimeID)
	_, err := writer.Write(append(uint32_buf, uint8(len(cmd.ChestSlots))))
	// They are different parts, but wrote together for convenient
	if err != nil {
		return err
	}
	for _, slot := range cmd.ChestSlots {
		binary.BigEndian.PutUint16(uint16_buf, slot.Damage)
		tmpbuf := append([]byte(slot.Name), []byte{0, slot.Count}...)
		tmpbuf = append(tmpbuf, append(uint16_buf, slot.Slot)...)
		_, err = writer.Write(tmpbuf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd *PlaceRuntimeBlockWithChestDataAndUint32RuntimeID) Unmarshal(reader io.Reader) error {
	uint32_buf := make([]byte, 4)
	_, err := io.ReadAtLeast(reader, uint32_buf, 4)
	if err != nil {
		return err
	}
	cmd.BlockRuntimeID = binary.BigEndian.Uint32(uint32_buf)
	uint8_buf := make([]byte, 1)
	_, err = io.ReadAtLeast(reader, uint8_buf, 1)
	if err != nil {
		return err
	}
	cmd.ChestSlots = make([]types.ChestSlot, int(uint8_buf[0]))
	for i := 0; i < int(uint8_buf[0]); i++ {
		item_name, err := readString(reader)
		if err != nil {
			return err
		}
		cmd.ChestSlots[i].Name = item_name
		count_damage_slot_buf := make([]byte, 4)
		_, err = io.ReadAtLeast(reader, count_damage_slot_buf, 4)
		if err != nil {
			return err
		}
		cmd.ChestSlots[i].Count = count_damage_slot_buf[0]
		cmd.ChestSlots[i].Damage = binary.BigEndian.Uint16(count_damage_slot_buf[1:3])
		cmd.ChestSlots[i].Slot = count_damage_slot_buf[3]
	}
	return nil
}
