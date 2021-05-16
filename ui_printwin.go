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

type printWinUI struct {
	File        string
	UI          UI
	initialised bool
}

func (p *printWinUI) Initialise() error {
	p.initialised = true
	return p.UI.Initialise()
}

func (p *printWinUI) NewRound(g *Game, round int) {
	if p.UI != nil {
		p.UI.NewRound(g, round)
	}
}

func (p *printWinUI) NewData(data GameData) {
	if p.UI != nil {
		p.UI.NewData(data)
	}
}

func (p *printWinUI) Finish(won bool, survived, round int) error {
	var err error
	if p.UI != nil {
		err = p.UI.Finish(won, survived, round)
	}

	if p.initialised {
		f, newErr := os.Create(p.File)
		if newErr != nil {
			return newErr
		}
		defer f.Close()

		if won {
			_, newErr = f.WriteString(fmt.Sprintf("\nWin!"))
			if newErr != nil {
				return newErr
			}

		} else {
			_, newErr = f.WriteString(fmt.Sprintf("\nLoss! (%d / %d)", survived, round))
			if newErr != nil {
				return newErr
			}
		}

		if newErr != nil {
			return newErr
		}
	}

	return err
}

func (p *printWinUI) Wait() {
	if p.UI != nil {
		p.UI.Wait()
	}
}
