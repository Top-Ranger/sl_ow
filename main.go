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

// sl_ow is a solution for the InformatiCup2021 by Marcus Soll
// It works by randomly simulating games and choosing the best action.
// It is based on a heavily modified server.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// GameData holds the metadata of a round.
type GameData struct {
	Alive   bool
	Collect map[string]struct {
		Run               int
		Won               int
		Survived          int
		SurvivdedOpponent int
		Round             int
		SurvivedList      []uint16
		LongestOpponent   int
	}

	LongestWin       int
	LongestWinAction string
	Longest          int
	LongestAction    string

	Action string
	Reason string

	Game    *Game
	Round   int
	Jumps   int
	Runtime time.Duration
}

func main() {
	rand.Seed(time.Now().UnixNano())

	endpoint := flag.String("api", "wss://msoll.de/spe_ed", "API Endpoint")
	key := flag.String("key", "KEY", "API key")
	quiet := flag.Bool("quiet", false, "Only print result")
	maxDurationString := flag.String("max", "", "Max computation time in seconds. 0 or empty string disables max time. Must be parseable as time.Duration")
	profile := flag.String("profile", "", "Profile program to file")
	print := flag.String("print", "", "Prints output into file")
	showui := flag.Bool("ui", false, "Enables cmd ui")
	dump := flag.String("dump", "", "Dumps game data as gob to file")
	printWin := flag.String("printwin", "", "Prints outcome of the game as a simple \"Win/Loss\" into file")
	flag.Parse()

	// Replace flags
	{
		env := os.Getenv("URL")
		if env != "" {
			fmt.Println("Using URL from env:", env)
			*endpoint = env
		}

		env = os.Getenv("KEY")
		if env != "" {
			fmt.Println("Using KEY from env:", env)
			*key = env
		}
	}

	// Max Duration
	var maxDuration time.Duration
	if *maxDurationString != "" {
		var err error
		maxDuration, err = time.ParseDuration(*maxDurationString)
		if err != nil {
			panic(err)
		}
		if maxDuration != 0 && maxDuration < 500*time.Millisecond {
			panic(fmt.Errorf("max must be at least 500ms (is %s) or else simulations can not run", maxDuration.String()))
		}
	}

	var UI UI
	if *quiet {
		UI = quietUI{}
	} else if *showui {
		UI = new(terminalUI)
	} else {
		UI = cmdUI{}
	}

	defer func() {
		err := recover()
		if err != nil {
			// Clearly close UI
			if UI != nil {
				UI.Finish(false, -1, -1)
				UI.Wait()
			}
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Panicln(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Panicln(err)
		}
		defer pprof.StopCPUProfile()
	}

	if *print != "" {
		UI = &teeUI{File: *print, UI: UI}
	}

	if *dump != "" {
		UI = &dumpUI{File: *dump, UI: UI}
	}

	if *printWin != "" {
		UI = &printWinUI{File: *printWin, UI: UI}
	}

	url := fmt.Sprintf("%s?key=%s", *endpoint, url.QueryEscape(*key))

	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		log.Panicln(err)
	}

	err = UI.Initialise()
	if err != nil {
		panic(err)
	}

	var mastergame *Game
	round := 0
	lastAlive := 0
	numberWorker := runtime.NumCPU()
	jumpsObserved := 0
	var start time.Time

	for {
		round++
		_, b, err := conn.ReadMessage()
		if err != nil {
			log.Panicln(err)
		}
		if start.IsZero() {
			start = time.Now()
		}
		mastergame = new(Game)
		err = json.Unmarshal(b, mastergame)
		if err != nil {
			log.Panicln(err)
		}
		// Add round since that is not transmitted
		for k := range mastergame.Players {
			mastergame.Players[k].stepCounter = round - 1
		}

		if mastergame.Running == false || !mastergame.Players[mastergame.You].Active {
			UI.NewData(GameData{
				Alive: mastergame.Players[mastergame.You].Active,
				Collect: make(map[string]struct {
					Run               int
					Won               int
					Survived          int
					SurvivdedOpponent int
					Round             int
					SurvivedList      []uint16
					LongestOpponent   int
				}),
				LongestWin:       0,
				LongestWinAction: "",
				Longest:          0,
				LongestAction:    "",
				Action:           "nothing",
				Reason:           "",
				Round:            round,
				Game:             mastergame,
				Jumps:            jumpsObserved,
				Runtime:          time.Now().Sub(start),
			})
		}

		if mastergame.Running == false {
			break
		}

		mastergame.PopulateInternalCellsFlat()

		UI.NewRound(mastergame.PublicCopy(), round)

		if !mastergame.Players[mastergame.You].Active {
			continue
		}

		lastAlive = round

		deadline, err := time.Parse(time.RFC3339, mastergame.Deadline)
		if err != nil {
			log.Panicln(err)
		}
		if maxDuration > 0 {
			test := time.Now().Add(maxDuration)
			if test.Before(deadline) {
				deadline = test
			}
		}
		ctxWorker, ctxWorkerCancel := context.WithDeadline(context.Background(), deadline.Add(-500*time.Millisecond))
		ctxMain, ctxMainCancel := context.WithDeadline(context.Background(), deadline.Add(-250*time.Millisecond))

		results := make(chan struct {
			action            string
			win               bool
			survived          int
			survivdedOpponent int
			round             int
		}, numberWorker)

		data := GameData{
			Alive: true,
			Collect: make(map[string]struct {
				Run               int
				Won               int
				Survived          int
				SurvivdedOpponent int
				Round             int
				SurvivedList      []uint16
				LongestOpponent   int
			}),
			LongestWin:       0,
			LongestWinAction: "",
			Longest:          0,
			LongestAction:    "",
			Action:           "nothing",
			Reason:           "",
			Round:            round,
			Game:             mastergame,
		}

		for i := 0; i < numberWorker; i++ {
			go func() {
				for {
					select {
					case <-ctxWorker.Done():
						return
					default:
						g := mastergame.PublicCopy()
						test := []string{ActionTurnLeft, ActionTurnRight, ActionSlower, ActionFaster, ActionNOOP}[rand.Intn(5)]
						g.SimulateGame(test, results)
					}
				}
			}()
		}

	collectorWorker:
		for {
			select {
			case r := <-results:
				d := data.Collect[r.action]
				d.Run++
				if r.win {
					d.Won++
					if r.survived > data.LongestWin {
						data.LongestWin = r.survived
						data.LongestWinAction = r.action
					}
				}
				if r.survived > data.Longest {
					data.Longest = r.survived
					data.LongestAction = r.action
				}
				if r.survivdedOpponent > d.LongestOpponent {
					d.LongestOpponent = r.survivdedOpponent
				}
				d.Survived += r.survived
				d.SurvivdedOpponent += r.survivdedOpponent
				d.Round += r.round
				d.SurvivedList = append(d.SurvivedList, uint16(r.survived))
				data.Collect[r.action] = d
			case <-ctxMain.Done():
				break collectorWorker
			}
		}

		ctxWorkerCancel()
		ctxMainCancel()

		// Sort survivedList
		for k := range data.Collect {
			d := data.Collect[k]
			sort.Slice(d.SurvivedList, func(i, j int) bool { return i < j })
			data.Collect[k] = d
		}

		best := 0.0

		if data.Action == "nothing" {
			for k := range data.Collect {
				d := data.Collect[k]
				if d.Run == 0 {
					continue
				}
				winchance := float64(d.Won) / float64(d.Run)
				if winchance > 0.85 && winchance > best {
					data.Reason = "win > 85%"
					data.Action = k
					best = winchance
				}
			}
		}

		best = 0.0

		if data.Action == "nothing" {
			for k := range data.Collect {
				d := data.Collect[k]
				if d.Run == 0 {
					continue
				}
				averageLength := float64(d.Survived) / float64(d.Run)
				if averageLength > best && float64(d.Won)/float64(d.Run) > 0.1 {
					best = averageLength
					data.Action = k
					data.Reason = "average length"
				}
			}
		}

		if data.Action == "nothing" {
			// In case no win path is found
			if data.LongestAction != "" {
				data.Action = data.LongestAction
				data.Reason = "longest path"
			}
		}

		if data.Action == "nothing" {
			data.Action = ActionNOOP
			data.Reason = "fallback"
		}

		answer, err := json.Marshal(Action{data.Action})
		if err != nil {
			log.Panicln(err)
		}
		err = conn.WriteMessage(websocket.TextMessage, answer)
		if err != nil {
			log.Panicln(err)
		}

		if isJump(mastergame.PublicCopy(), data.Action) {
			jumpsObserved++
		}

		data.Jumps = jumpsObserved
		data.Runtime = time.Now().Sub(start)

		UI.NewData(data)
	}

	UI.Finish(mastergame.Players[mastergame.You].Active, lastAlive, round)
	UI.Wait()
}

func (g Game) String() string {
	return g.PrintGame(false)
}

// PrintGame returns a string representation of the game cells.
func (g Game) PrintGame(colour bool) string {
	var s strings.Builder
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
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
			if colour && g.Cells[y][x] != -1 {
				s.WriteString(colours[g.Cells[y][x]])
			}
			s.WriteRune(r)
			if colour {
				s.WriteString(colourReset)
			}
		}
		if y < g.Height-1 {
			s.WriteRune('\n')
		}
	}

	return s.String()
}

func isJump(g *Game, action string) bool {
	// Not a jump round? Too slow?
	if (g.Players[g.You].stepCounter+1)%HolesEachStep != 0 {
		return false
	}

	// Process action
	switch action {
	case ActionTurnLeft:
		switch g.Players[g.You].Direction {
		case DirectionLeft:
			g.Players[g.You].Direction = DirectionDown
		case DirectionRight:
			g.Players[g.You].Direction = DirectionUp
		case DirectionUp:
			g.Players[g.You].Direction = DirectionLeft
		case DirectionDown:
			g.Players[g.You].Direction = DirectionRight
		}
	case ActionTurnRight:
		switch g.Players[g.You].Direction {
		case DirectionLeft:
			g.Players[g.You].Direction = DirectionUp
		case DirectionRight:
			g.Players[g.You].Direction = DirectionDown
		case DirectionUp:
			g.Players[g.You].Direction = DirectionRight
		case DirectionDown:
			g.Players[g.You].Direction = DirectionLeft
		}
	case ActionFaster:
		g.Players[g.You].Speed++
		if g.Players[g.You].Speed > MaxSpeed {
			return false
		}
	case ActionSlower:
		g.Players[g.You].Speed--
		if g.Players[g.You].Speed < 1 {
			return false
		}
	case ActionNOOP:
		// Do nothing
	default:
		log.Println("isJump:", "unknown action", action)
		return false
	}

	if g.Players[g.You].Speed < HoleSpeed {
		return false
	}

	// Check Jump
	jumpFound := false

	var dostep func(x, y int) (int, int)
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

	for s := 0; s < g.Players[g.You].Speed; s++ {
		g.Players[g.You].X, g.Players[g.You].Y = dostep(g.Players[g.You].X, g.Players[g.You].Y)
		if g.Players[g.You].X < 0 || g.Players[g.You].X >= g.Width || g.Players[g.You].Y < 0 || g.Players[g.You].Y >= g.Height {
			return false
		}
		if s != 0 && s != g.Players[g.You].Speed-1 {
			if g.Cells[g.Players[g.You].Y][g.Players[g.You].X] != 0 {
				jumpFound = true
			}
			continue
		}
		if g.Cells[g.Players[g.You].Y][g.Players[g.You].X] != 0 {
			return false
		}
	}
	return jumpFound
}
