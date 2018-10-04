package main

import (
	"github.com/platinasystems/go/internal/fdt"

	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var debug bool

// GPIO Test Code ++
type gpioAliasMap map[string]string

var gpioAlias gpioAliasMap

const (
	// Lower 16 bits gives index.
	PinIndexMask uint32 = 0xffff
	// High bits are flags: direction input/output
	IsInput    uint32 = 1 << 31
	IsOutputLo uint32 = 1 << 30
	IsOutputHi uint32 = 1 << 29
)

type PinMap map[string]uint32

var gpios PinMap
var gpioBankToBase = map[string]uint32{
	"gpio0": 0,
	"gpio1": 32,
	"gpio2": 64,
	"gpio3": 96,
	"gpio4": 128,
	"gpio5": 160,
	"gpio6": 192,
}
var gpioPinMode = map[string]uint32{
	"output-high": IsOutputHi,
	"output-low":  IsOutputLo,
	"input":       IsInput,
}

// Build map of gpio pins for this gpio controller
func gatherGpioAliases(n *fdt.Node) {
	for p, pn := range n.Properties {
		if strings.Contains(p, "gpio") {
			val := strings.Split(string(pn), "\x00")
			v := strings.Split(val[0], "/")
			gpioAlias[p] = v[len(v)-1]
		}
	}
}

func buildPinMap(name string, mode string, bank string, index string) {
	if debug {
		fmt.Println("Pinmap-entry:", name, mode, bank, index)
	}
	i, _ := strconv.Atoi(index)
	gpios[name] = gpioPinMode[mode] | gpioBankToBase[bank] | uint32(i)
}

// Build map of gpio pins for this gpio controller
func gatherGpioPins(n *fdt.Node, name string, value string) {
	var pn []string
	var mode string

	for na, al := range gpioAlias {
		if al == n.Name {
			for _, c := range n.Children {
				for p, _ := range c.Properties {
					switch p {
					case "gpio-pin-desc":
						pn = strings.Split(c.Name, "@")
					case "output-high", "output-low", "input":
						mode = p
					}
				}
				if mode != "" {
					buildPinMap(pn[0], mode, na, pn[1])
				}
				mode = ""
			}
		}
	}
}

// GPIO Test Code --

// I2C Test Code ++

type i2cAliasMap map[string]string

var i2cAlias i2cAliasMap

func gatherI2cAliases(n *fdt.Node) {
	for p, pn := range n.Properties {
		if strings.Contains(p, "i2c") {
			val := strings.Split(string(pn), "\x00")
			v := strings.Split(val[0], "/")
			i2cAlias[v[len(v)-1]] = p
		}
	}
}

func printI2CTree(n *fdt.Node, name string, value string) {

	for na, _ := range i2cAlias {
		if na == n.Name {
			for _, c := range n.Children {
				fmt.Printf("%*s%s", 2*n.Depth, " ", c.Name)
			}
		}
	}
}

// I2C Test Code --

func main() {
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	t := &fdt.Tree{Debug: false, IsLittleEndian: false}
	t.Parse(b)

	if false {
		fmt.Printf("%v\n", t)
	}
	if true {
		gpioAlias = make(gpioAliasMap)
		gpios = make(PinMap)

		t.MatchNode("aliases", gatherGpioAliases)
		if debug {
			fmt.Println(gpioAlias)
		}

		t.EachProperty("gpio-controller", "", gatherGpioPins)
		for p, v := range gpios {
			fmt.Printf("%s: %x\n", p, v)
		}
	}
	if false {
		i2cAlias = make(i2cAliasMap)

		t.MatchNode("aliases", gatherI2cAliases)
		if debug {
			fmt.Println(i2cAlias)
		}

		// Pull out i2c busses - 0-based; sort them too
		sortedAliases := make([]string, len(i2cAlias))
		i := 0
		for k, _ := range i2cAlias {
			sortedAliases[i] = k
			i++
		}
		sort.Strings(sortedAliases)
		for j := 0; j < len(sortedAliases); j++ {
			fmt.Println("/", i2cAlias[sortedAliases[j]])
			t.EachNodeFrom(sortedAliases[j], func(n *fdt.Node) { fmt.Printf("%*s%s\n", 2*n.Depth, " ", n.Name) })
		}
	}
	if false {
		// Pull out gpio controllers i.e. banks
		t.EachProperty("gpio-controller", "",
			func(n *fdt.Node, name string, value string) { fmt.Println(n) })
	}
}
