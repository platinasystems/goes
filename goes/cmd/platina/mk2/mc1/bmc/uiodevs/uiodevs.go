// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//
package uiodevs

import (
	"os"
	"fmt"
	"syscall"

	"github.com/platinasystems/log"
)

const (
	MAX_IRQS = 32
)



func getUioName(x string) string {
	return map[string]string{
		"qsfpeventsd":		"g7_5",	// MAIN_GPIO0_INT_L
                "lceventsd":		"g7_6",	// MAIN_GPIO1_INT_L
                "psueventsd": 		"g7_7", // MAIN_GPIO2_INT_L
                "faneventsd": 		"g7_8", // MAIN_GPIO6_INT_L
                "psupwreventsd": 	"g7_9", // MAIN_GPIO7_INT_L
	}[x]
}
	

func GetIndex(n string) (int, error) {
	var name string = ""
	var index int = int(99)

	uname := getUioName(n)
	for i := 0; (i < MAX_IRQS); i++ {
		dir := fmt.Sprintf("/sys/class/uio/uio%d/name", i)
                file, err := os.Open(dir)
                if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
                        log.Print("open file: ", err)
			return int(0), err
                } else {
			 _, err = fmt.Fscan(file, &name)
	                file.Close()
         	       	if err != nil {
				log.Print("name scan: ", err)
                	}
			if uname == name {
				index = i
				break
			}
		}
	}	
	return index, nil
}
