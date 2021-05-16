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

const (
	// DirectionUp contains the string value representing "up"
	DirectionUp = "up"
	// DirectionDown contains the string value representing "down"
	DirectionDown = "down"
	// DirectionLeft contains the string value representing "left"
	DirectionLeft = "left"
	// DirectionRight contains the string value representing "right"
	DirectionRight = "right"
)

// Player represents a player of the game.
// Compared to the server, this struct only holds the values and the current AI.
type Player struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Direction string `json:"direction"`
	Speed     int    `json:"speed"`
	Active    bool   `json:"active"`
	Name      string `json:"name,omitempty"`

	// To know where wholes need to be
	stepCounter int

	ai AI
}
