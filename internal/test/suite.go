// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type Tester interface {
	String() string
	Test(*testing.T)
}

type Tests []Tester

func (tests Tests) Test(t *testing.T) {
	for _, x := range tests {
		if t.Failed() {
			break
		}
		t.Run(x.String(), x.Test)
	}
}

type Suite struct {
	Name string
	Init func(*testing.T)
	Exit func(*testing.T)
	Tests
}

// Log args if -test.vv
func (suite *Suite) Comment(t *testing.T, args ...interface{}) {
	t.Helper()
	if *VV {
		t.Log(args...)
	}
}

// Format args if -test.vv
func (suite *Suite) Commentf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	if *VV {
		t.Logf(format, args...)
	}
}

func (suite *Suite) String() string {
	return suite.Name
}

func (suite *Suite) init(t *testing.T) {
	t.Helper()
	begin := time.Now()
	suite.Init(t)
	suite.Commentf(t, "%s.Init (%v)", t.Name(), time.Now().Sub(begin))
}

func (suite *Suite) exit(t *testing.T) {
	t.Helper()
	begin := time.Now()
	suite.Exit(t)
	suite.Commentf(t, "%s.Exit (%v)", t.Name(), time.Now().Sub(begin))
}

func (suite Suite) Test(t *testing.T) {
	if *DryRun {
		fmt.Println(t.Name())
	} else {
		if suite.Exit != nil {
			defer suite.exit(t)
		}
		if suite.Init != nil {
			suite.init(t)
		}
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
	} else if len(u.Name) > 0 && !strings.Contains(u.Name, " ") {
		fmt.Println(t.Name())
	}
}
