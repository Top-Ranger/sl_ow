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
	"encoding/gob"
	"os"
)

type dumpUI struct {
	File       string
	UI         UI
	gameStates []GameData
}

func (d *dumpUI) Initialise() error {
	return d.UI.Initialise()
}

func (d *dumpUI) NewRound(g *Game, round int) {
	if d.UI != nil {
		d.UI.NewRound(g, round)
	}
}

func (d *dumpUI) NewData(data GameData) {
	d.gameStates = append(d.gameStates, data)

	if d.UI != nil {
		d.UI.NewData(data)
	}
}

func (d *dumpUI) Finish(won bool, survived, round int) error {
	var err error
	if d.UI != nil {
		err = d.UI.Finish(won, survived, round)
	}

	if len(d.gameStates) == 0 {
		return err
	}

	f, newErr := os.Create(d.File)
	if newErr != nil {
		return newErr
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	newErr = enc.Encode(d.gameStates)

	if newErr != nil {
		return newErr
	}

	return err
}

func (d *dumpUI) Wait() {
	if d.UI != nil {
		d.UI.Wait()
	}
}
