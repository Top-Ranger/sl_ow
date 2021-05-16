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
	"sync"
)

// JumpingSnailAIJumpAtLessThanFree is the number of free cells connected at which the AI tries to jump.
const JumpingSnailAIJumpAtLessThanFree = 50

// JumpingSnailAI behaves like SnailAI but tries to jump out of small areas.
type JumpingSnailAI struct {
	l sync.Mutex

	i                 chan string
	snail             AI
	jump              AI
	freeCountingSlice []bool
}

// GetChannel receives the answer channel.
func (js *JumpingSnailAI) GetChannel(c chan string) {
	js.l.Lock()
	defer js.l.Unlock()

	js.i = c

	if js.snail != nil {
		js.snail.GetChannel(c)
	}

	if js.jump != nil {
		js.jump.GetChannel(c)
	}
}

// GetState gets the game state and computes an answer.
func (js *JumpingSnailAI) GetState(g *Game) {
	js.l.Lock()
	defer js.l.Unlock()

	if js.i == nil {
		return
	}

	if js.snail == nil {
		js.snail = new(SnailAI)
		js.snail.GetChannel(js.i)
	}

	if g.Running && g.Players[g.You].Active {
		if js.freeSpaceConnected(g.Players[g.You].X, g.Players[g.You].Y, JumpingSnailAIJumpAtLessThanFree+1, g) < JumpingSnailAIJumpAtLessThanFree {
			if js.jump == nil {
				js.jump = new(JumpAI)
				js.jump.GetChannel(js.i)
			}
			js.jump.GetState(g)
		} else if g.Players[g.You].Speed > 1 {

			// Check slow_down
			var dostep func(x, y int) (int, int)
			possible := true
			x, y := g.Players[g.You].X, g.Players[g.You].Y

			switch g.Players[g.You].Direction {
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			for i := 0; i < g.Players[g.You].Speed-1; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case js.i <- ActionSlower:
				default:
				}
				return
			}

			// Check turn_left
			possible = true
			x, y = g.Players[g.You].X, g.Players[g.You].Y

			switch g.Players[g.You].Direction {
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			for i := 0; i < g.Players[g.You].Speed; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case js.i <- ActionTurnLeft:
				default:
				}
				return
			}

			// Check turn_right
			possible = true
			x, y = g.Players[g.You].X, g.Players[g.You].Y

			switch g.Players[g.You].Direction {
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			for i := 0; i < g.Players[g.You].Speed; i++ {
				x, y = dostep(x, y)
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
					possible = false
					break
				}
			}

			if possible {
				select {
				case js.i <- ActionTurnRight:
				default:
				}
				return
			}

			// nothing to survive, so....
			select {
			case js.i <- ActionSlower:
			default:
			}
		} else {
			js.jump = nil
			js.snail.GetState(g)
		}
	}
}

// Name returns the name of the AI.
func (js *JumpingSnailAI) Name() string {
	return "JumpingSnailAI"
}

// freeSpaceConnected calculates the number of free space connected to given area.
// It is not safe for concurrent usage on the same instance of JumpingSnailAI.
func (js *JumpingSnailAI) freeSpaceConnected(x, y, cutoff int, g *Game) int {
	// cutoff -1 == no cutoff
	// Not concurrent safe

	if len(js.freeCountingSlice) != g.Height*g.Width {
		js.freeCountingSlice = make([]bool, g.Height*g.Width)
	} else {
		for i := range js.freeCountingSlice {
			js.freeCountingSlice[i] = false
		}
	}

	current := 0

	current = js.freeSpaceConnectedInternal(x, y, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x-1, y, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x+1, y, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x, y-1, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x, y+1, cutoff, current, g)

	return current
}

func (js *JumpingSnailAI) freeSpaceConnectedInternal(x, y, cutoff, current int, g *Game) int {
	if cutoff != -1 && current > cutoff {
		return current
	}

	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return current
	}

	cell := y*g.Width + x

	if js.freeCountingSlice[cell] {
		return current
	}
	js.freeCountingSlice[cell] = true

	if g.Cells[y][x] != 0 {
		return current
	}
	current++

	current = js.freeSpaceConnectedInternal(x-1, y, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x+1, y, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x, y-1, cutoff, current, g)
	current = js.freeSpaceConnectedInternal(x, y+1, cutoff, current, g)

	return current
}
