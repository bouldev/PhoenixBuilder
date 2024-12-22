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

import "fmt"

// 机器人进服后的挑战之一
type GetMCPCheckNum struct {
	FirstArg  string
	SecondArg GetMCPCheckNumSecondArg
}

type GetMCPCheckNumSecondArg struct {
	Arg             string // e.g. "uBic92T8NS"
	FirstExtraData  []any  // e.g. []interface {}{}
	SecondExtraData []any  // e.g. []interface {}{}
}

// Return the name of g
func (g *GetMCPCheckNum) Name() string {
	return "GetMCPCheckNum"
}

// Convert g to go object which only contains go-built-in types
func (g *GetMCPCheckNum) MakeGo() (res any) {
	if g == nil {
		return nil
	}
	// check
	return []any{
		g.FirstArg,
		[]any{
			g.SecondArg.Arg,
			g.SecondArg.FirstExtraData,
			g.SecondArg.SecondExtraData,
		},
	}
	// return
}

// Sync data to g from obj
func (g *GetMCPCheckNum) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	if len(object) != 2 {
		return fmt.Errorf("FromGo: The length of object is not equal to 2; object = %#v", object)
	}
	// convert and check data
	first_arg, success := object[0].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[0] to string; object[0] = %#v`, object[0])
	}
	// get first arg
	second_arg, success := object[1].([]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[1] to []interface{}; object[1] = %#v`, object[1])
	}
	if len(second_arg) != 3 {
		return fmt.Errorf("FromGo: The length of second_arg is not equal to 3; second_arg = %#v", second_arg)
	}
	sa_arg, success := second_arg[0].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert second_arg[0] to string; second_arg[0] = %#v`, second_arg[0])
	}
	sa_first_extra_data, success := second_arg[1].([]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert second_arg[1] to []interface{}; second_arg[1] = %#v`, second_arg[1])
	}
	sa_second_extra_data, success := second_arg[2].([]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert second_arg[2] to []interface{}; second_arg[2] = %#v`, second_arg[2])
	}
	// get second arg
	*g = GetMCPCheckNum{
		FirstArg: first_arg,
		SecondArg: GetMCPCheckNumSecondArg{
			Arg:             sa_arg,
			FirstExtraData:  sa_first_extra_data,
			SecondExtraData: sa_second_extra_data,
		},
	}
	// sync data
	return nil
	// return
}
