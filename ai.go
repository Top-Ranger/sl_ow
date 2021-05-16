// SPDX-License-Identifier: Apache-2.0
// Copyright 2020,2021 Marcus Soll
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"math/rand"
)

// The AI interface provides the interface for different AIs.
//
// All channel actions must be non-blocking.
//
// In GetState, AIs can only access public fields (and change them) plus stepCounter, privat fields are set to zero.
// Modification of the game is allowed. The Caller has to make sure that modifications to the provided game can be done without side effects (e.g. by using Game.PublicCopy )
type AI interface {
	GetChannel(c chan string)
	GetState(g *Game)
	Name() string
}

// GetAI returns a singe AI out of the current rotation.
func GetAI() AI {
	var AIArray = []func() AI{
		func() AI { return new(EndRound) },
		func() AI { return new(HeartAI) },
		func() AI { return new(ChristmasAI) },
		func() AI { return new(StupidAI) },
		func() AI { return new(StupidAI) },
		func() AI { return new(StupidAI) },
		func() AI { return new(StupidAI) },
		func() AI { return new(StupidAI) },
		func() AI { return new(SnailAI) },
		func() AI { return new(SnailAI) },
		func() AI { return new(SuperSnailAI) },
		func() AI { return new(SuperSnailAI) },
		func() AI { return new(SuperSnailAI) },
		func() AI { return new(JumpingSnailAI) },
		func() AI { return new(JumpingSnailAI) },
		func() AI { return new(JumpingSnailAI) },
		func() AI { return new(JumpingSnailAI) },
		func() AI { return new(JumpingSnailAI) },
		func() AI { return new(LargestFreeAI) },
		func() AI { return new(LargestFreeAI) },
		func() AI { return new(LargestFreeAI) },
		func() AI { return new(LargestFreeAI) },
		func() AI { return new(LargestFreeAI) },
		func() AI { return new(JumpingLargestFreeAI) },
		func() AI { return new(JumpingLargestFreeAI) },
		func() AI { return new(JumpingLargestFreeAI) },
		func() AI { return new(JumpingLargestFreeAI) },
		func() AI { return new(JumpingLargestFreeAI) },
		func() AI { return new(RandomAI) },
		func() AI { return new(RandomAI) },
		func() AI { return new(BadRandomAI) },
		func() AI { return new(RandomAISlow) },
		func() AI { return new(RandomAISlow) },
		func() AI { return new(SuperRandomAI) },
		func() AI { return new(SuperRandomAI) },
		func() AI { return new(SuperRandomAI) },
		func() AI { return new(SuperRandomAI) },
		func() AI { return new(SuperRandomAI) },
		func() AI { return new(MirrorAI) },
		func() AI { return new(MirrorAI) },
		func() AI { return new(MirrorAI) },
		func() AI { return new(MirrorAI) },
		func() AI { return new(JumpAI) },
		func() AI { return new(JumpAI) },
		func() AI { return new(JumpAI) },
		func() AI { return new(JumpAI) },
		func() AI { return new(JumpAI) },
		func() AI { return new(MetaAI) },
		func() AI { return new(MetaAI) },
		func() AI { return new(MetaAI) },
		func() AI { return new(MetaAI) },
		func() AI { return new(MetaAI) },
	}
	return AIArray[rand.Intn(len(AIArray))]()
}
