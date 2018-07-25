// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"os"
	"testing"
)

func (suite *Suite) String() string {
	return suite.Name
}

func (suite *Suite) init(t *testing.T) {
	if suite.Init != nil {
		suite.Init(t)
	}
}

func (suite *Suite) exit(t *testing.T) {
	if suite.Exit != nil {
		suite.Exit(t)
	}
}

func (suite Suite) Run(t *testing.T) {
	if !*DryRun {
		defer suite.exit(t)
		suite.init(t)
	}
	for _, x := range suite.Tests {
		if t.Failed() {
			break
		}
		t.Run(x.String(), x.Run)
	}
}

type Tester interface {
	String() string
	Run(*testing.T)
}

type Tests []Tester

type Suite struct {
	Name string
	Init func(*testing.T)
	Exit func(*testing.T)
	Tests
}

type UnitTest struct {
	Name string
	Func func(*testing.T)
}

func (ut *UnitTest) String() string { return ut.Name }

func (ut *UnitTest) Run(t *testing.T) {
	if *DryRun {
		os.Stdout.WriteString(t.Name())
		os.Stdout.WriteString("\n")
	} else {
		ut.Func(t)
	}
}
