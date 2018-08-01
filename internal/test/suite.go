// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"fmt"
	"testing"
)

type Tester interface {
	String() string
	Test(*testing.T)
}

type Tests []Tester

// Named Tests use testing.T.Run() but unnamed Tests ("") are run directly.
func (tests Tests) Test(t *testing.T) {
	for _, x := range tests {
		if t.Failed() {
			break
		}
		if len(x.String()) > 0 {
			t.Run(x.String(), x.Test)
		} else {
			x.Test(t)
		}
	}
}

type Suite struct {
	Name string
	Init func(*testing.T)
	Exit func(*testing.T)
	Tests
}

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

func (suite Suite) Test(t *testing.T) {
	if *DryRun {
		fmt.Println(t.Name())
	} else {
		defer suite.exit(t)
		suite.init(t)
	}
	suite.Tests.Test(t)
}

type Unit struct {
	Name string
	Func func(*testing.T)
}

func (u *Unit) String() string { return u.Name }

func (u *Unit) Test(t *testing.T) {
	if !*DryRun {
		u.Func(t)
	} else if len(u.Name) > 0 {
		fmt.Println(t.Name())
	}
}
