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
	"fmt"
	"strings"
	"time"
)

var colours = []string{"\033[39m", "\033[31m", "\033[32m", "\033[33m", "\033[34m", "\033[35m", "\033[36m"}
var colourReset = "\033[0m"

// The UI interface allows the usage of different UIs in sl_ow.
type UI interface {
	Initialise() error
	NewRound(g *Game, round int)
	NewData(data GameData)
	Finish(won bool, survived, round int) error
	Wait()
}

func buildGameOverviewStrings(gd GameData, round, rounds int, detailled bool) []string {
	g := gd.Game
	if g == nil {
		return nil
	}
	ss := make([]string, 0)
	ss = append(ss, fmt.Sprintf("game state %d/%d", round, rounds))
	var sb strings.Builder

	sb.WriteString("alive: [ ")
	for i := 1; i <= 6; i++ {
		if g.Players[i] == nil {
			break
		} else if g.Players[i].Active {
			if i == g.You {
				sb.WriteRune('>')
			}
			sb.WriteRune(rune(48 + i))
			if i == g.You {
				sb.WriteRune('<')
			}
			sb.WriteRune(' ')
		} else {
			if i == g.You {
				sb.WriteRune('>')
			}
			sb.WriteRune('_')
			if i == g.You {
				sb.WriteRune('<')
			}
			sb.WriteRune(' ')
		}
	}
	sb.WriteRune(']')
	ss = append(ss, sb.String())
	ss = append(ss, fmt.Sprintf("alive: %t", gd.Alive))
	ss = append(ss, fmt.Sprintf("size: %d x %d", gd.Game.Width, gd.Game.Height))
	ss = append(ss, fmt.Sprintf("usage: %.2f", 1.0-gd.Game.usage(0)))
	ss = append(ss, fmt.Sprintf("runtime: %s", gd.Runtime.Truncate(1*time.Second).String()))

	if gd.Alive {
		ss = append(ss, "")
		ss = append(ss, fmt.Sprintf("speed: %d", g.Players[g.You].Speed))
		ss = append(ss, fmt.Sprintf("jumps: %d", gd.Jumps))
		ss = append(ss, "")
		for _, action := range []string{ActionNOOP, ActionSlower, ActionFaster, ActionTurnLeft, ActionTurnRight} {
			ss = append(ss, fmt.Sprintf("%s:", action))
			ss = append(ss, fmt.Sprintf("   win chance: %.2f", float64(gd.Collect[action].Won)/float64(gd.Collect[action].Run)))
			ss = append(ss, fmt.Sprintf("   run: %d", gd.Collect[action].Run))
			ss = append(ss, fmt.Sprintf("   won: %d", gd.Collect[action].Won))
			ss = append(ss, fmt.Sprintf("   survived: %d", gd.Collect[action].Survived))
			ss = append(ss, fmt.Sprintf("   round: %d", gd.Collect[action].Round))
			ss = append(ss, fmt.Sprintf("   average length: %.1f", float64(gd.Collect[action].Survived)/float64(gd.Collect[action].Run)))
			if detailled {
				ss = append(ss, fmt.Sprintf("   average length best opponent: %.1f", float64(gd.Collect[action].SurvivdedOpponent)/float64(gd.Collect[action].Run)))
				if len(gd.Collect[action].SurvivedList) == 0 {
					ss = append(ss, fmt.Sprintf("   1st quantile length not available"))
					ss = append(ss, fmt.Sprintf("   median length not available"))
					ss = append(ss, fmt.Sprintf("   3rd quantile length not available"))
				} else {
					ss = append(ss, fmt.Sprintf("   1st quantile length: %d", gd.Collect[action].SurvivedList[len(gd.Collect[action].SurvivedList)/4]))
					ss = append(ss, fmt.Sprintf("   median length: %d", gd.Collect[action].SurvivedList[len(gd.Collect[action].SurvivedList)/2]))
					ss = append(ss, fmt.Sprintf("   3rd quantile length: %d", gd.Collect[action].SurvivedList[len(gd.Collect[action].SurvivedList)/4*3]))
				}
			}
		}
		ss = append(ss, fmt.Sprintf("longest win: %d %s", gd.LongestWin, gd.LongestWinAction))
		ss = append(ss, fmt.Sprintf("longest: %d %s", gd.Longest, gd.LongestAction))
		ss = append(ss, "")
		ss = append(ss, fmt.Sprintf("selected: %s", gd.Action))
		ss = append(ss, fmt.Sprintf("reason: %s", gd.Reason))
		ss = append(ss, "")
	} else {
		empty := 13*5 + 10
		if detailled {
			empty += 4 * 5
		}
		for i := 0; i < empty; i++ {
			ss = append(ss, "")
		}
	}
	return ss
}
