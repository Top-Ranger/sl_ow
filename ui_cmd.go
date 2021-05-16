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
)

type cmdUI struct {
}

func (c cmdUI) Initialise() error {
	fmt.Printf("Waiting for game\n")
	return nil
}

func (c cmdUI) NewRound(g *Game, round int) {
	fmt.Println()
	fmt.Println(g.PrintGame(true))
	fmt.Println()

	fmt.Printf("Round %d - You: %s%d%s (alive: [ ", round, colours[g.You], g.You, colourReset)
	for i := 1; i <= 6; i++ {
		if g.Players[i] == nil {
			break
		} else if g.Players[i].Active {
			fmt.Print(colours[i])
			if i == g.You {
				fmt.Print("\033[4m")
			}
			fmt.Print(i)
			fmt.Print(colourReset)
			fmt.Print(" ")
		} else {
			fmt.Print(colours[0])
			fmt.Print("⬜ ")
			fmt.Print(colourReset)
		}
	}
	fmt.Print("] )\r")
}

func (c cmdUI) NewData(data GameData) {
	// Don't show survived array
	ss := buildGameOverviewStrings(data, data.Round, -1, false)
	for i := range ss {
		if strings.TrimSpace(ss[i]) != "" {
			if i == 0 {
				ss[i] = fmt.Sprintf("%-50s", ss[i])
			}
			fmt.Println(ss[i])
		}
	}
	fmt.Println()
}

func (c cmdUI) Finish(won bool, survived, round int) error {
	if won {
		fmt.Printf("\nWin!\n\n")
	} else {
		fmt.Printf("\nLoss! (%d / %d)\n\n", survived, round)
	}
	return nil
}

func (c cmdUI) Wait() {
}
