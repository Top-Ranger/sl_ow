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
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell"
)

type terminalUI struct {
	screen          tcell.Screen
	gameStates      []GameData
	gameStateIndex  int
	colors          map[int]tcell.Color
	ctx             context.Context
	done            context.CancelFunc
	firstGame       chan bool
	newData         chan GameData
	running         chan bool
	positionRunning int
	once            *sync.Once
}

func (tui *terminalUI) Initialise() error {
	var err error

	tui.firstGame = make(chan bool, 5)
	tui.newData = make(chan GameData, 5)
	tui.running = make(chan bool, 5)
	tui.ctx, tui.done = context.WithCancel(context.Background())
	tui.once = new(sync.Once)

	tui.screen, err = tcell.NewScreen()
	if err != nil {
		return err
	}

	err = tui.screen.Init()
	if err != nil {
		return err
	}

	tui.colors = map[int]tcell.Color{
		-1: tcell.ColorBlack,
		0:  tcell.ColorBlack,
		1:  tcell.NewRGBColor(178, 24, 24),
		2:  tcell.NewRGBColor(24, 178, 24),
		3:  tcell.NewRGBColor(178, 104, 24),
		4:  tcell.NewRGBColor(24, 24, 178),
		5:  tcell.NewRGBColor(178, 24, 178),
		6:  tcell.NewRGBColor(24, 178, 178),
	}

	tui.drawString(0, 0, "Waiting for game")
	tui.positionRunning = FieldMaxSize + 2 + 30
	tui.drawString(tui.positionRunning, 0, "ready")
	tui.screen.Show()

	go tui.mainLoop()

	return nil
}

func (tui *terminalUI) NewRound(g *Game, round int) {
	tui.once.Do(func() {
		select {
		case tui.firstGame <- true:
		default:
		}
	})
}

func (tui *terminalUI) NewData(data GameData) {
	select {
	case tui.newData <- data:
	default:
	}
}

func (tui *terminalUI) Finish(won bool, survived, round int) error {
	select {
	case tui.running <- false:
	default:
	}
	return nil
}

func (tui *terminalUI) Wait() {
	<-tui.ctx.Done()
}

func (tui *terminalUI) drawString(x, y int, v string) {
	for i, r := range v {
		tui.screen.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
}

func (tui *terminalUI) drawGameState() {
	gd := tui.gameStates[tui.gameStateIndex]

	g := gd.Game

	if g == nil {
		tui.screen.Clear()
		return
	}
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			r := g.runeAt(y, x)
			v := int(g.Cells[y][x])
			tui.screen.SetContent(x, y, r, nil, tcell.StyleDefault.Foreground(tui.colors[v]))
		}
	}
	ss := buildGameOverviewStrings(gd, tui.gameStateIndex+1, len(tui.gameStates), false)
	ox := g.Width + 2
	for i, s := range ss {
		if i == 0 {
			tui.drawString(ox, i, fmt.Sprintf("%-30s", s))
		} else {
			tui.drawString(ox, i, fmt.Sprintf("%-100s", s))
		}
	}
	tui.screen.Show()
}

func (g *Game) runeAt(y, x int) rune {
	v := int(g.Cells[y][x])
	r := rune(48 + v)
	switch {
	case v == 0:
		r = '·'
	case v > 0:
		if v == g.You {
			r = '●'
		}
		p := g.Players[v]
		if p.X == x && p.Y == y {
			switch p.Direction {
			case DirectionUp:
				r = '⮝'
				if v == g.You {
					r = '⮉'
				}
			case DirectionRight:
				r = '⮞'
				if v == g.You {
					r = '⮊'
				}
			case DirectionDown:
				r = '⮟'
				if v == g.You {
					r = '⮋'
				}
			case DirectionLeft:
				r = '⮜'
				if v == g.You {
					r = '⮈'
				}
			}
		}
	case v == -1:
		r = '×'
	}
	return r
}

func (tui *terminalUI) mainLoop() {
	running := true
	ec := make(chan tcell.Event, 0)
	go func() {
		for {
			e := tui.screen.PollEvent()
			ec <- e
		}
	}()

	for {
		select {
		case e := <-ec:
			switch ev := e.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyHome:
					tui.gameStateIndex = 0
					tui.drawGameState()
				case tcell.KeyEnd:
					tui.gameStateIndex = len(tui.gameStates) - 1
					tui.drawGameState()
				case tcell.KeyLeft:
					if tui.gameStateIndex > 0 {
						tui.gameStateIndex--
						tui.drawGameState()
					}
				case tcell.KeyRight:
					if tui.gameStateIndex < len(tui.gameStates)-1 {
						tui.gameStateIndex++
						tui.drawGameState()
					}
				case tcell.KeyRune:
					if ev.Rune() != 'q' {
						continue
					}
					fallthrough
				case tcell.KeyEscape, tcell.KeyCtrlC:
					tui.screen.Fini()
					tui.done()
					if running {
						fmt.Println("terminal ui closed")
					}
					return
				}
			}
		case gs, ok := <-tui.newData:
			if ok && gs.Game != nil {
				tui.gameStates = append(tui.gameStates, gs)
				if tui.gameStateIndex == len(tui.gameStates)-2 {
					tui.gameStateIndex = len(tui.gameStates) - 1
				}
				tui.drawGameState()
			}
		case _ = <-tui.running:
			running = false
			tui.drawString(tui.positionRunning, 0, "finished")
			tui.screen.Show()
			tui.running = nil
		case _ = <-tui.firstGame:
			tui.drawString(tui.positionRunning, 0, "running")
			tui.screen.Show()
			tui.firstGame = nil
		}
	}
}
