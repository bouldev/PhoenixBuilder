package core

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
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	fbauth "phoenixbuilder/fastbuilder/mv4"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/mirror/io/global"
	"phoenixbuilder/mirror/io/lru"
)

func create_environment() *environment.PBEnvironment {
	clientOption := fbauth.MakeDefaultClientOptions()
	clientOption.AuthServer = args.AuthServer
	clientOption.RespondUserOverride = args.CustomGameName
	env := &environment.PBEnvironment{
		ClientOptions: clientOption,
	}
	env.Resources = nil
	env.TaskHolder = fbtask.NewTaskHolder()
	functionHolder := function.NewFunctionHolder(env)
	env.FunctionHolder = functionHolder
	env.LRUMemoryChunkCacher = lru.NewLRUMemoryChunkCacher(12, false)
	env.ChunkFeeder = global.NewChunkFeeder()
	return env
}

// Shouldn't be called when running a debug client
func ConfigRealEnvironment(token string, server_code string, server_password string, username string, password string) (*environment.PBEnvironment, error) {
	env := create_environment()
	env.LoginInfo = environment.LoginInfo{
		Token:          token,
		ServerCode:     server_code,
		ServerPasscode: server_password,
		Username:       username,
		Password:       password,
	}
	authClient, err := fbauth.CreateClient(env.ClientOptions)
	if err != nil {
		return nil, err
	}
	env.FBAuthClient = authClient
	return env, nil
}

func ConfigDebugEnvironment() *environment.PBEnvironment {
	env := create_environment()
	env.IsDebug = true
	env.LoginInfo = environment.LoginInfo{
		ServerCode: "[DEBUG]",
	}
	return env
}

func DestroyEnvironment(env *environment.PBEnvironment) {
	env.Connection.(*minecraft.Conn).Close()
}
