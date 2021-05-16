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
	"os"
)

type teeUI struct {
	File string
	UI   UI
	f    *os.File
}

func (t *teeUI) Initialise() error {
	if t.f != nil {
		return fmt.Errorf("file already opened")
	}
	var err error
	t.f, err = os.Create(t.File)
	if err != nil {
		t.f = nil
		return err
	}
	if t.UI != nil {
		return t.UI.Initialise()
	}
	return nil
}

func (t *teeUI) NewRound(g *Game, round int) {
	if t.UI != nil {
		t.UI.NewRound(g, round)
	}
}

func (t *teeUI) NewData(data GameData) {
	if t.f != nil {
		t.f.WriteString("\n")
		t.f.WriteString(data.Game.PrintGame(false))
		t.f.WriteString("\n\n")

		t.f.WriteString(fmt.Sprintf("Round %d - You: %d\n", data.Round, data.Game.You))

		ss := buildGameOverviewStrings(data, data.Round, -1, false)
		for i := range ss {
			t.f.WriteString(ss[i])
			t.f.WriteString("\n")
		}
		t.f.WriteString("\n")
	}

	if t.UI != nil {
		t.UI.NewData(data)
	}
}

func (t *teeUI) Finish(won bool, survived, round int) error {
	var err error
	if t.f != nil {
		if won {
			t.f.WriteString(fmt.Sprintf("\nWin!"))

		} else {
			t.f.WriteString(fmt.Sprintf("\nLoss! (%d / %d)", survived, round))
		}

		err = t.f.Close()
	}
	if t.UI != nil {
		newErr := t.UI.Finish(won, survived, round)
		if newErr != nil && err != nil {
			return fmt.Errorf("two errors: %s, %s", err.Error(), newErr.Error())
		} else if newErr != nil {
			err = newErr
		}
	}
	return err
}

func (t *teeUI) Wait() {
	if t.UI != nil {
		t.UI.Wait()
	}
}
