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
	// FieldMaxSize contains the maximum size of the field (both width and height).
	FieldMaxSize = 80
	// HolesEachStep holds after how many steps a hole might occur (if the preconditions are met).
	HolesEachStep = 6
	// HoleSpeed contains the minimum speed needed for a hole.
	HoleSpeed = 3
	// MaxSpeed holds the maximum speed.
	MaxSpeed = 10
)

// Game represents a game of speed. See https://github.com/informatiCup/InformatiCup2021/ for a description of the game.
// This struct is a modified from the server version to fit sl_ow.
type Game struct {
	Width             int             `json:"width"`
	Height            int             `json:"height"`
	Cells             [][]int8        `json:"cells"`
	Players           map[int]*Player `json:"players"`
	You               int             `json:"you"` // only needed for protocol, ignored everywhere else
	Running           bool            `json:"running"`
	Deadline          string          `json:"deadline,omitempty"` // RFC3339
	playerAnswer      []string
	freeCountingSlice []bool

	internalCellsFlat []int8
}

// SimulateGame simulates a full run of the game and sends the result to the provided channel.
// It has some early cut-offs for impossible games.
func (g *Game) SimulateGame(next string, result chan<- struct {
	action            string
	win               bool
	survived          int
	survivdedOpponent int
	round             int
}) {
	// Check speed
	if next == ActionSlower && g.Players[g.You].Speed == 1 {
		result <- struct {
			action            string
			win               bool
			survived          int
			survivdedOpponent int
			round             int
		}{next, false, 0, 0, 0}
		return
	}
	if next == ActionFaster && g.Players[g.You].Speed == MaxSpeed {
		result <- struct {
			action            string
			win               bool
			survived          int
			survivdedOpponent int
			round             int
		}{next, false, 0, 0, 0}
		return
	}

	for k := range g.Players {
		g.Players[k].ai = &SuperRandomAI{}
	}

	first := true
	round := 0
	survived := -1
	survivedOpponent := -1
	winner := -1
mainGame:
	for { // Loop used for rounds
		round++
		if g.Players[g.You].Active {
			survived++
		}
		for i := range g.Players {
			if i == g.You {
				continue
			}
			if g.Players[g.You].Active {
				survivedOpponent++
				break
			}
		}
		g.playerAnswer = make([]string, len(g.Players))
		for i := range g.playerAnswer {
			if i+1 == g.You && first {
				g.playerAnswer[i] = next
				continue
			}
			if g.Players[i+1].Active {
				if first {
					// First round take a possible turn even if suicide is possible - that's why RandomAI can't be used here
					answer := make(chan string, 1)
					ng := g.PublicCopy()
					ng.You = i + 1
					ai := BadRandomAI{}
					ai.GetChannel(answer)
					ai.GetState(ng)
					g.playerAnswer[i] = <-answer
					continue
				}

				// Use AI
				answer := make(chan string, 1)

				ng := g.PublicCopy()
				ng.You = i + 1
				g.Players[i+1].ai.GetChannel(answer)
				g.Players[i+1].ai.GetState(ng)
				g.playerAnswer[i] = <-answer
			}
		}

		first = false

		// Process Actions
		for i := range g.Players {
			switch g.playerAnswer[i-1] {
			case "":
				g.invalidatePlayer(i)
			case ActionTurnLeft:
				switch g.Players[i].Direction {
				case DirectionLeft:
					g.Players[i].Direction = DirectionDown
				case DirectionRight:
					g.Players[i].Direction = DirectionUp
				case DirectionUp:
					g.Players[i].Direction = DirectionLeft
				case DirectionDown:
					g.Players[i].Direction = DirectionRight
				}
			case ActionTurnRight:
				switch g.Players[i].Direction {
				case DirectionLeft:
					g.Players[i].Direction = DirectionUp
				case DirectionRight:
					g.Players[i].Direction = DirectionDown
				case DirectionUp:
					g.Players[i].Direction = DirectionRight
				case DirectionDown:
					g.Players[i].Direction = DirectionLeft
				}
			case ActionFaster:
				g.Players[i].Speed++
				if g.Players[i].Speed > MaxSpeed {
					g.invalidatePlayer(i)
				}
			case ActionSlower:
				g.Players[i].Speed--
				if g.Players[i].Speed < 1 {
					g.invalidatePlayer(i)
				}
			case ActionNOOP:
				// Do nothing
			default:
				g.invalidatePlayer(i)
			}
		}

		// Do Movement
		for i := range g.Players {
			if !g.Players[i].Active {
				continue
			}
			var dostep func(x, y int) (int, int)
			switch g.Players[i].Direction {
			case DirectionUp:
				dostep = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionDown:
				dostep = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionLeft:
				dostep = func(x, y int) (int, int) { return x - 1, y }
			case DirectionRight:
				dostep = func(x, y int) (int, int) { return x + 1, y }
			}

			g.Players[i].stepCounter++

			for s := 0; s < g.Players[i].Speed; s++ {
				g.Players[i].X, g.Players[i].Y = dostep(g.Players[i].X, g.Players[i].Y)
				if g.Players[i].X < 0 || g.Players[i].X >= g.Width || g.Players[i].Y < 0 || g.Players[i].Y >= g.Height {
					g.invalidatePlayer(i)
					break
				}
				if g.Players[i].Speed >= HoleSpeed && g.Players[i].stepCounter%HolesEachStep == 0 && s != 0 && s != g.Players[i].Speed-1 {
					continue
				}
				if g.Cells[g.Players[i].Y][g.Players[i].X] != 0 {
					g.Cells[g.Players[i].Y][g.Players[i].X] = -1
				} else {
					g.Cells[g.Players[i].Y][g.Players[i].X] = int8(i)
				}
			}
		}

		// Check crash
		for i := range g.Players {
			if !g.Players[i].Active {
				continue
			}
			var dostepback func(x, y int) (int, int)
			switch g.Players[i].Direction {
			case DirectionUp:
				dostepback = func(x, y int) (int, int) { return x, y + 1 }
			case DirectionDown:
				dostepback = func(x, y int) (int, int) { return x, y - 1 }
			case DirectionLeft:
				dostepback = func(x, y int) (int, int) { return x + 1, y }
			case DirectionRight:
				dostepback = func(x, y int) (int, int) { return x - 1, y }
			}

			backX := g.Players[i].X
			backY := g.Players[i].Y
			for s := 0; s < g.Players[i].Speed; s++ {
				if g.Cells[backY][backX] == -1 {
					// Crash - check hole
					if g.Players[i].Speed >= HoleSpeed && g.Players[i].stepCounter%HolesEachStep == 0 && s != 0 && s != g.Players[i].Speed-1 {
						// No crash - is hole
					} else {
						g.invalidatePlayer(i)
						break
					}
				}
				backX, backY = dostepback(backX, backY)
			}
		}

		if winner == -1 {
			for i := range g.Players {
				if g.Players[i].Active {
					if winner == -1 {
						winner = i
					} else {
						// Game hasn't finished - at least two alive
						winner = -1
						break
					}
				}
			}
		}

		playerAlive := false
		for i := range g.Players {
			if g.Players[i].Active {
				playerAlive = playerAlive || g.Players[i].Active
			}
		}

		if !playerAlive {
			break mainGame
		}
	}
	// Finish game
	g.Running = false

	if winner == g.You {
		survived = round
	}
	result <- struct {
		action            string
		win               bool
		survived          int
		survivdedOpponent int
		round             int
	}{next, winner == g.You, survived, survivedOpponent, round}
}

func (g *Game) checkEndGame() bool {
	numberActive := 0
	for i := range g.Players {
		if g.Players[i].Active {
			numberActive++
		}
	}
	return numberActive <= 1
}

func (g *Game) invalidatePlayer(p int) {
	_, ok := g.Players[p]
	if !ok {
		return
	}
	g.Players[p].Active = false
}

// PublicCopy returns a copy of the game with all private fields set to zero.
// As an exception for AIs, Player.stepCounter is also copied.
func (g Game) PublicCopy() *Game {
	newG := Game{
		Width:    g.Width,
		Height:   g.Height,
		Cells:    make([][]int8, len(g.Cells)),
		Players:  make(map[int]*Player, len(g.Players)),
		You:      g.You,
		Running:  g.Running,
		Deadline: g.Deadline,
	}

	if g.internalCellsFlat == nil {
		for i := range g.Cells {
			newG.Cells[i] = make([]int8, len(g.Cells[i]))
			copy(newG.Cells[i], g.Cells[i])
		}
	} else {
		newG.internalCellsFlat = make([]int8, len(g.internalCellsFlat))
		copy(newG.internalCellsFlat, g.internalCellsFlat)
		for y := range g.Cells {
			newG.Cells[y] = newG.internalCellsFlat[y*newG.Width : (y+1)*newG.Width]
		}
	}

	for k := range g.Players {
		newG.Players[k] = &Player{
			X:           g.Players[k].X,
			Y:           g.Players[k].Y,
			Direction:   g.Players[k].Direction,
			Speed:       g.Players[k].Speed,
			Active:      g.Players[k].Active,
			Name:        g.Players[k].Name,
			stepCounter: g.Players[k].stepCounter,
		}
	}
	return &newG
}

func (g *Game) usage(id int) float64 {
	u := 0
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			if g.Cells[y][x] == int8(id) {
				u++
			}
		}
	}

	return float64(u) / float64(g.Height*g.Width)
}
func (g *Game) freeSpaceConnected(x, y, cutoff int) int {
	// cutoff -1 == no cutoff
	// Not concurrent safe

	if g.freeCountingSlice == nil {
		g.freeCountingSlice = make([]bool, g.Height*g.Width)
	} else {
		for i := range g.freeCountingSlice {
			g.freeCountingSlice[i] = false
		}
	}

	current := 0

	current = g.freeSpaceConnectedInternal(x, y, cutoff, current)
	current = g.freeSpaceConnectedInternal(x-1, y, cutoff, current)
	current = g.freeSpaceConnectedInternal(x+1, y, cutoff, current)
	current = g.freeSpaceConnectedInternal(x, y-1, cutoff, current)
	current = g.freeSpaceConnectedInternal(x, y+1, cutoff, current)

	return current
}

func (g *Game) freeSpaceConnectedInternal(x, y, cutoff, current int) int {
	if cutoff != -1 && current > cutoff {
		return current
	}

	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return current
	}

	cell := y*g.Width + x

	if g.freeCountingSlice[cell] {
		return current
	}
	g.freeCountingSlice[cell] = true

	if g.Cells[y][x] != 0 {
		return current
	}
	current++

	current = g.freeSpaceConnectedInternal(x-1, y, cutoff, current)
	current = g.freeSpaceConnectedInternal(x+1, y, cutoff, current)
	current = g.freeSpaceConnectedInternal(x, y-1, cutoff, current)
	current = g.freeSpaceConnectedInternal(x, y+1, cutoff, current)

	return current
}

// PopulateInternalCellsFlat populates the internal flat cells, thus providing a speed boost to PublicCopy after being called.
// The game object is not safe to use while this function is running.
func (g *Game) PopulateInternalCellsFlat() {
	g.internalCellsFlat = make([]int8, g.Height*g.Width)

	for y := range g.Cells {
		for x := range g.Cells[y] {
			g.internalCellsFlat[y*g.Width+x] = g.Cells[y][x]
		}
	}

	g.Cells = make([][]int8, len(g.Cells))
	for y := range g.Cells {
		g.Cells[y] = g.internalCellsFlat[y*g.Width : (y+1)*g.Width]
	}
}
