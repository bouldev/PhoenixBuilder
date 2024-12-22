package environment

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

// This package imports only external packages to avoid import cycle.
import (
	"phoenixbuilder/fastbuilder/environment/interfaces"
	fbauth "phoenixbuilder/fastbuilder/mv4"
)

type LoginInfo struct {
	Token          string
	Username       string
	Password       string
	ServerCode     string
	ServerPasscode string
}

type PBEnvironment struct {
	LoginInfo
	IsDebug               bool
	FunctionHolder        interfaces.FunctionHolder
	FBAuthClient          interface{}
	GlobalFullConfig      interface{}
	RespondTo             string
	Connection            interface{}
	GetCheckNumEverPassed bool
	CachedPacket          interface{}
	Resources             interface{}
	ResourcesUpdater      interface{}
	GameInterface         interfaces.GameInterface
	TaskHolder            interface{}
	LRUMemoryChunkCacher  interface{}
	ChunkFeeder           interface{}
	ClientOptions         *fbauth.ClientOptions
}
