// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot' client for auto-install - runs on each client device
// loaded as a Coreboot payload - kernel+initrd+goes
// kernel + initrd will be loaded by Coreboot or PXE boot

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
    MASTER TO KNOW HE IS MASTER WHEN STARTING WEBSERVER (how?)

TO DO
    register and manage state machine index, with timestamps
    define state machine states
    maintain state list for each client (100 max)
    pass down a goes or linux script, i.e. JSON and exec
    add real test infra
    add test case of 100 units simultaneously registering
    Installing apt-gets support
    console
    port to accept reset

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
  5. RUN INSTALL/PRESEED
*/

//TODO remove globals

package bootc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/platinasystems/go/goes/lang"
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
	s := ""
	mip := getMasterIP()

	switch c {
	case 1:
		mac := getMAC()
		ip := getIP()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 2:
		mac := getMAC2()
		ip := getIP2()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 3:
		mac := getMAC3()
		ip := getIP3()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 4:
		if err = dumpVars(mip); err != nil {
			return err
		}
	case 5:
		if err = dashboard(mip); err != nil {
			return err
		}
	case 6:
		if err = test404(mip); err != nil {
			return err
		}
	default:
		fmt.Println("no command...")
	}
	return nil
}

// */

func bootSequencer() (err error) { // "init" function for Coreboot
	mip := getMasterIP()
	mac := getMAC()
	ip := getIP()
	// possibly try register 5 times with delay in between
	if _, err = register(mip, mac, ip); err != nil {
		// register failed: possibly try just booting sda2, forget registering
		return err
	}
	// unpack script

	// run script (format, install debian, etc. OR just boot)

	// if debian install fails ==> try again, then try PXE boot

	return nil
}

func register(mip string, mac string, ip string) (s string, err error) {
	if s, err = sendReq(mip, "register "+mac+" "+ip); err != nil { // FIXME string body JSON
		return "", err // FIXME ERROR FMT MESSAGE
	}
	return s, nil //FIXME name, script, err
}

func sendReq(mip string, s string) (res string, err error) {
	resp, err := http.Get("http://" + mip + ":9090/" + s)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getMasterIP() string {
	// FIXME try .1, try DNS, try URL, try hardcode list
	return "192.168.101.142" //hardcode for now
}

func getIP() string {
	return "192.168.101.142" // FIXME hardcode for now
}

func getMAC() string {
	return "01:02:03:04:05:06" // FIXME hardcode for now
}

//
// debugging functions
//

func dumpVars(mip string) (err error) { // debugging
	s := ""
	if s, err = sendReq(mip, "dumpvars"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func dashboard(mip string) (err error) { // debugging
	s := ""
	if s, err = sendReq(mip, "dashboard"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func test404(mip string) (err error) { // debugging
	s := ""
	if s, err = sendReq(mip, "xxx"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func getIP2() string {
	return "192.168.101.143" //hardcode for now
}

func getIP3() string {
	return "192.168.101.144" //hardcode for now
}

func getMAC2() string {
	return "01:02:03:04:05:07" //hardcode for now
}

func getMAC3() string {
	return "01:02:03:04:05:08" //hardcode for now
}
