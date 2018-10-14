// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"
)

//after vlan interface removed, containers deleted, there should be no adjacency rewrites leftover
func NoAdjacency(t *testing.T) (err error) {
	fail := false
	time.Sleep(1 * time.Second)

	// Check leftover adjacencies:
	// Should be no rewrites after interfaces are admin down
	cmd := exec.Command("goes", "vnet", "show", "fe1", "adj")
	out, _ := cmd.Output()
	out_string := fmt.Sprintf("%s\n", out)
	re := regexp.MustCompile("hard.*l3_unicast.*true.*")
	rewrites := re.FindAllStringSubmatch(out_string, -1)
	num := len(rewrites)
	if num > 0 {
		fmt.Printf("number of rewrites not zero\n")
		fail = true
		if *VV {
			for i := range rewrites {
				fmt.Printf("%v\n", rewrites[i])
			}
		}
	}

	// Check leftover l3 IIF rules:
	// For vlan tests, the vlans are admin down, but the underlying xeth may be up.
	// Even though there is no interface addr for them, show fe1 l3 may still have class_id:nil reference to them
	// Check only that all vlan interfaces class_id_nil are removed from show fe1 l3
	for pipe := 0; pipe < 4; pipe++ {
		re := regexp.MustCompile("xeth.*\\.....")
		r, w, _ := os.Pipe()
		p := strconv.Itoa(pipe)
		cmd := exec.Command("goes", "vnet", "show", "fe1", "l3", "pipe", p)
		grep := exec.Command("grep", "nil")
		grep.Stdin, _ = cmd.StdoutPipe()
		grep.Stdout = w
		grep.Start()
		cmd.Run()
		grep.Wait()
		w.Close()
		out, _ := ioutil.ReadAll(r)
		out_string := fmt.Sprintf("%s", out)
		match := re.MatchString(out_string)
		if match {
			fmt.Printf("pipe %v shouldn't have: %s\n", p, re.FindAllStringSubmatch(out_string, -1))
			fmt.Printf("%v", out_string)
			fail = match
		}
	}
	if fail {
		err = fmt.Errorf("unexpected entries in adjacency or l3 table")
		Pause("No adjacency test Failed\n")
	}
	return err
}
