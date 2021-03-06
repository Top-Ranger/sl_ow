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

// Action represents an Action send to the server. It is mainly used to unmarshal JSON answers.
type Action struct {
	Action string `json:"action"`
}

const (
	// ActionTurnLeft represent the action "turn_left".
	ActionTurnLeft = "turn_left"
	// ActionTurnRight represent the action "turn_right".
	ActionTurnRight = "turn_right"
	// ActionSlower represent the action "slow_down".
	ActionSlower = "slow_down"
	// ActionFaster represent the action "speed_up".
	ActionFaster = "speed_up"
	// ActionNOOP represent the action "change_nothing".
	ActionNOOP = "change_nothing"
)

// IsValidAction returns whether a string is a valid action.
func IsValidAction(a string) bool {
	return a == ActionTurnLeft || a == ActionTurnRight || a == ActionSlower || a == ActionFaster || a == ActionNOOP
}
