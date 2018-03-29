// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot' requestor, this will be run in parallel on muliple client devices
// this will be run automatically from the kernel+initrd(goes) image
// kernel + initrd will be loaded by Coreboot or PXE boot

// DISCLAIMER
// this is a work in progress, this will change significantly before release
// this package must be manually added to the mk1 goes.go to be included ATM

/* DESIGN NOTES
    STATE MACHINE ON MASTER FOR TOR-X86, TOR-BMC, and non-TOR
    ALL MESSAGING ORIGINATES ON CLIENT
    IF KEEPALIVES STOP, MASTER CAN POSSIBLY CHOOSE TO RESET CLIENT (how? bmc?)
    KEEPALIVES GET AN ACK or INSTRUCTIONS A A REPLY
    POSSIBLY TUNNEL CONSOLE THROUGH BMC LIKE INTEL
    DATABASE IS PRE-PROGRAMMED FOR ZERO TOUCH INSTALL
    FOR NON-PLATINA EQUIP: k&i install one-of-us borgify (or--always pxe boot k&i)
    AT INSTALL: k&i install/partition
    AT INSTALL: deb install w/preseed
    MASTER CAN FORCE A REBOOT AND RE-INSTALL

TO DO
    convert array to struct
    load structs from local database file
    register and manage state machine index, with timestamps
    define state machine states
    reply via json tftp
    multiple replies on server
    maintain state list for each client (100 max)
    progress dashboard showing state per unit
    pass down a goes or linux script, i.e. JSON and exec
    add real test infra
    add test case of 100 units simultaneously registering
    Installing apt-gets support

    CLIENT                                     MASTER
      |                                          |
      v                                          v
                                              FUTURE: PRIME MASTER FROM INTERNET
					      ASSUME PRE-PRIMED MASTER FOR NOW

   POWERON                                    POWERON
   BOOT K&I FROM FLASH (OR PXE BOOT K&I)      BOOT K&I (MASTER --> so boot SDA2)
   DETERMINE OUR MAC, IP, CERT   	      DHCP ON
   DETERMINE LIST OF POSS. MASTER IPs	      PXE SERVER K&I ON
					      VERIFY DEBIAN ISO
					      READ DATABASE (from local or cloud)
					      START HTTP SERVER (SERVES DASHBOARD TOO)
					      INIT CLIENT ARRAY OF STRUCTS
                                              SET ALL CLIENT STATES TO (0)

  CLIENT HTTP contact master           --->   MASTER message rec'd (A) STATE
           MESSAGE TYPE: REGISTER             DATABASE LOOKUP
	   IP                                 VERIFY CERT
	   MAC                                DB==INSTALLED?, RTN: NAME SCRIPT
	                                      script -> boot sda2 (B) STATE
	   MASTER IP                          ELSE NEEDS INSTALLED, (C) STATE
	   CERT                               REPLY WITH NAME, SCRIPT
	   MACHINE TYPE                       script -> install debian
	   CONTEXT: K&I or REAL LINUX         IF BMC, DIFFERENT STATE MACHINE

	                                      DATABASE, time of last good boot
					      installed or not
					      time since last keep alive

					      DEB INSTALL GOOD (D) STATE (REBOOT)
					      DEB INSTALL FAILS (0) STATE -REBOOT
                                       <---
           DISPLAY NAME
	   EXECUTE SCRIPT
	   AFTER NORMAL BOOT ->               KEEP TRACK OF LAST 10 KEEPALIVE TIMESTAMPS
	    SEND KEEPALIVE MSG PERIODICALLY   SAVE IN DB LAST BOOT TIME, INSTALL OK


    LIST OF KEY ELEMENTS
    (a) boot(/init) to contact server and run script, boot sda2
    (b) kernel+initrd+boot(/init) payload
    (c) web based dashboard
    (d) configuration database indexed by mac/cert (stored on local or cloud)
    (e) boot-controller(webserver) on master tor
    (f) debian isos (etc.) on master tor
    (g) preseed file to answer debian install questions

    (h) NEAR FUTURE: hand off to ansible and follow on steps (pre to post container)
    (i) FUTURE: x509 cert support
    (j) FUTURE: modify debian installer to install Coreboot (ToR only?)

boot sequence
    1. CB boots kernel+initrd with new goes "boot logic" as /init
    2. boot logic loops through the list of possible server addresses
        the top candidate is our DHCP address with .1 as lowest octet
    3. register with server.  Server will supply our instructions:
           a. hey, normal boot off of sda2
           or
           b. run this script
               (typically: format sda2, install debian, install goes, reboot)

    if registration fails ==>  fall through to normal boot from SDA2
    if normal boot from SDA2 fails ==> try PXE boot

units of work
  1. CB to boot goes payload (this is done I think)
  2. ability to run goes scripts in goes (this is done I think)
  3. ability for initial goes to format SDA2 and install debian (does this work??)
  4. PXE boot from CB (I think this is not done)
*/

package bootc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
)

// /*
type Command struct{}

func (Command) String() string { return "bootc" }

func (Command) Usage() string { return "bootc" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "boot client hook to communicate with tor master",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
description
	the bootc command is for debugging bootc client.`,
	}
}

func (Command) Main(args ...string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("args: missing")
	}

	c, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	switch c {
	case 0: // pre-registration
		if err = preRegistration(); err != nil {
			return err
		}
	case 1: // registration, display json reply
		if err = registration(); err != nil {
			return err
		}
	case 2: // show server variobles
		if err = dumpServerVars(); err != nil {
			return err
		}
	case 3: // show dashboard
		if err = showDashboard(); err != nil {
			return err
		}
	default:
		fmt.Println("no command...")
	}
	return nil
}

// */

//
// "/init" in the CB kernel+initrd+goes payload
//
func bootSequence() (err error) {

	// determine possible server addresses
	// loop through server addresses to register with master ToR
	//   if registration fails ==> try normal SDA2 boot
	// server replied with script
	//   either boot sda2
	//   or do format, debian install, etc.
	//   if debian install fails ==> try again, then try PXE boot

}

// pre-registration - get mac, ip, etc.
func preRegister() error {
	return nil
}

// get our name, and our script
func register(mip, ip, mac string) (name string, script string, err error) {
	return "", "", nil
}

// for debugging
func dumpServerVars() error { //any browser can see this
	return nil
}

// for debugging
func showDashboard() error { //any browser can see this
	return nil
}

// send request to master ToR webserver //FIXME
func sendreq(c int, s string) {
	if c == 1 {
		resp, err := http.Get("http://192.168.101.142:9090/" + s)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}

	if c == 2 {
		resp, err := http.Get("http://192.168.101.142:9090/" + "dump")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}

	if 1 == 3 { //post case
		resp, err := http.PostForm("http://duckduckgo.com",
			url.Values{"q": {"github"}})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("post:\n", minLines(string(body), 3))
	}
}

func minLines(s string, n int) string {
	result := strings.Join(strings.Split(s, "\n")[:n], "\n")
	return strings.Replace(result, "\r", "", -1)
}

func func1() error {
	mac := getMAC()
	ip := getIP()
	masterIP := getMasterIP(ip)
	ourName, ourState, err := register(masterIP, ip, mac)
	if err != nil {
		fmt.Println(ourName, ourState)
		return err
	}
	return nil
}

func getMasterIP(ip string) string {
	return "192.168.101.142" //hardcode as ip for testing for now //use .1 or DNS, or WWW.PRIME.COM, or HARDCODE IP Or all of the above
}

func getIP() string {
	return "192.168.101.142" //hardcode for now
}

func getMAC() string {
	return "00:00:00:00:00:00" //hardcode for now
}
