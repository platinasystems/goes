// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

import (
	"fmt"
	"strconv"
	"time"

	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
)

const (
	pen = 0x0000BC65

	chassisTypeT = 0x50
	boardTypeT   = 0x51
	ppnT         = 0x53
	subTypeT     = 0x52

	chassisTypeL = 1
	boardTypeL   = 1
	ppnL         = 14
	subTypeL     = 1

	chassisTypeTor1  = 0x0
	chassisType4Ch1  = 0x1
	chassisType8Ch1  = 0x2
	chassisType16Ch1 = 0x3
	chassisTypeNone  = 0xFF

	boardTypeTor1   = 0x0
	boardTypeBde2c  = 0x1
	boardTypeBde4c  = 0x2
	boardTypeBde8c  = 0x3
	boardTypeCh1Mc1 = 0x4
	boardTypeCh1Lc1 = 0x5 // 32x100
	boardTypeCh1Mcb = 0x6
	boardTypeCh1Fan = 0x7

	ppnTor1   = "900-000000-000"
	ppnBde2c  = "900-000001-000"
	ppnBde4c  = "900-000002-000"
	ppnBde8c  = "900-000003-000"
	ppnCh1Lc1 = "900-000004-000"
	ppnCh1Mc1 = "900-000005-000"
	ppn16sCh1 = "900-000006-000"
	ppn8sCh1  = "900-000007-000"
	ppn4sCh1  = "900-000008-000"

	subTypeGa    = 0xA
	subTypeProto = 0x1
)

func diagProm() error {
	var c, v string

	//ONIE vendor exention fields
	tor1Vedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeTor1, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	bde2cVedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeBde2c, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	bde4cVedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeBde4c, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	tor1Vedga := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeTor1, subTypeT, subTypeL, subTypeGa, ppnT, ppnL}
	bde2cVedga := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeBde2c, subTypeT, subTypeL, subTypeGa, ppnT, ppnL}
	bde4cVedga := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisTypeTor1, boardTypeT, boardTypeL, boardTypeBde4c, subTypeT, subTypeL, subTypeGa, ppnT, ppnL}

	mc14sCh1Vedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisType4Ch1, boardTypeT, boardTypeL, boardTypeCh1Mc1, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	lc14sCh1Vedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisType4Ch1, boardTypeT, boardTypeL, boardTypeCh1Lc1, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	bde2c4sCh1Vedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisType4Ch1, boardTypeT, boardTypeL, boardTypeBde2c, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}
	bde4c4sCh1Vedp := []byte{byte(pen >> 24), byte(pen >> 16), byte(pen >> 8), byte(pen & 0xff), chassisTypeT, chassisTypeL, chassisType4Ch1, boardTypeT, boardTypeL, boardTypeBde4c, subTypeT, subTypeL, subTypeProto, ppnT, ppnL}

	var vByte []byte

	vf := uint(0)
	vl := uint(0)

	s, err := redis.Hget(redis.DefaultHash, "machine")
	if err != nil {
		log.Print("redis: ", err)
		return err
	} else {
		if s == "platina-mk1-bmc" {
			diagI2cWrite1Byte(0x00, 0x76, 0x00)
			diagI2cWrite1Byte(0x01, 0x72, 0x00)
		}
		if s == "platina-mk2-mc1-bmc" {
			diagI2cWrite1Byte(0x00, 0x71, 0x00)
			diagI2cWrite1Byte(0x00, 0x72, 0x00)
			diagI2cWrite1Byte(0x01, 0x70, 0x00)
		}
		if s == "platina-mk2-lc1-bmc" {
			diagI2cWrite1Byte(0x00, 0x74, 0x00)
		}
	}

	for i := 0; i < ppnL; i++ {
		tor1Vedp = append(tor1Vedp, ppnTor1[i])
		bde2cVedp = append(bde2cVedp, ppnBde2c[i])
		bde4cVedp = append(bde4cVedp, ppnBde4c[i])
		tor1Vedga = append(tor1Vedga, ppnTor1[i])
		bde2cVedga = append(bde2cVedga, ppnBde2c[i])
		bde4cVedga = append(bde4cVedga, ppnBde4c[i])

		mc14sCh1Vedp = append(mc14sCh1Vedp, ppnCh1Mc1[i])
		lc14sCh1Vedp = append(lc14sCh1Vedp, ppnCh1Lc1[i])
		bde2c4sCh1Vedp = append(bde2c4sCh1Vedp, ppnBde2c[i])
		bde4c4sCh1Vedp = append(bde4c4sCh1Vedp, ppnBde4c[i])
	}

	d := eeprom.Device{
		BusIndex: 0,
	}
	//select host or bmc prom
	if !flagF.ByName["-x86"] {
		d.BusAddress = 0x55
	} else {
		d.BusAddress = 0x51
		gpioSet("CPU_TO_MAIN_I2C_EN", true)
		time.Sleep(10 * time.Millisecond)
	}

	//dump prom
	if len(argF) == 1 {
		result, rawData := d.DumpProm()
		if !result {
			fmt.Printf("invalid: eeprom is not in onie format\n")
			return nil
		}
		fmt.Printf("raw data: %x\n\n", rawData)
		fmt.Printf("id: %s, rev: 0x%x, length: %d\n", string(rawData[0:7]), rawData[8], (uint(rawData[9])<<8)|uint(rawData[10]))
		fmt.Printf("Type | Length | Value \n")
		fmt.Printf("-----------------------------------------\n")
		for i := uint(0 + 11); i < uint(len(rawData)); {
			tlv, tlen := rawData[i], uint(rawData[i+1])
			v := rawData[i+2 : i+2+tlen]
			switch tlv {
			case 0x24, 0x26, 0x2a, 0xfe:
				fmt.Printf("0x%x |  %3d   |              %12x\n", tlv, tlen, v)
			case 0xfd:
				vf = i + 2
				vl = tlen
				fmt.Printf("0x%x |  %3d   |              output below\n", tlv, tlen)
			default:
				fmt.Printf("0x%x |  %3d   | %25s\n", tlv, tlen, string(v))
			}
			i += 2 + tlen
		}

		if vf != 0 {
			ved := rawData[vf : vf+vl]
			fmt.Printf("\npen: 0x%x\n", ved[0:4])
			if ved[0] != 0x00 || ved[1] != 0x00 || ved[2] != 0xbc || ved[3] != 0x65 {
				fmt.Print("Invalid vendor extension PEN\n")
			} else {
				fmt.Printf("Type | Length | Value \n")
				fmt.Printf("-----------------------------------------\n")
				for j := uint(4); j < uint(len(ved)); {
					tlv, tlen := ved[j], uint(ved[j+1])
					v := ved[j+2 : j+2+tlen]
					switch tlv {
					case 0x50, 0x51, 0x52:
						fmt.Printf("0x%x |  %3d   |              %12x\n", tlv, tlen, v)
					case 0x53, 0x54:
						fmt.Printf("0x%x |  %3d   | %25s\n", tlv, tlen, string(v))
					default:
					}
					j += 2 + tlen
				}
			}
		}
	} else if writeField && len(argF) == 3 {
		c = argF[1]
		v = argF[2]
		vByte = []byte(v)

		switch c {
		case "24":
			//write mac base address
			h, _ := strconv.ParseUint(v, 16, 64)
			b := []byte{byte((h >> 40) & 0xff), byte((h >> 32) & 0xff), byte((h >> 24) & 0xff), byte((h >> 16) & 0xff), byte((h >> 8) & 0xff), byte(h & 0xff)}
			fmt.Printf("%s\n", d.WriteField(c, b))
		case "26":
			//write version
			h, _ := strconv.ParseUint(v, 16, 64)
			b := []byte{byte(h & 0xff)}
			fmt.Printf("%s\n", d.WriteField(c, b))
		case "2a":
			//write number of macs
			h, _ := strconv.ParseUint(v, 16, 64)
			b := []byte{byte((h >> 8) & 0xff), byte(h & 0xff)}
			fmt.Printf("%s\n", d.WriteField(c, b))
		case "length":
			//write onie length field (debug tool to fix invalid format)
			n, _ := strconv.ParseUint(v, 10, 64)
			l := []byte{byte(n >> 8), byte(n & 0xff)}
			fmt.Printf("%s\n", d.WriteField(c, l))
		case "fd":
			//write vendor extension fields
			switch v {
			case "tor1p":
				fmt.Printf("%s\n", d.WriteField(c, tor1Vedp))
			case "bde4cp":
				fmt.Printf("%s\n", d.WriteField(c, bde4cVedp))
			case "bde2cp":
				fmt.Printf("%s\n", d.WriteField(c, bde2cVedp))
			case "tor1ga":
				fmt.Printf("%s\n", d.WriteField(c, tor1Vedga))
			case "bde4cga":
				fmt.Printf("%s\n", d.WriteField(c, bde4cVedga))
			case "bde2cga":
				fmt.Printf("%s\n", d.WriteField(c, bde2cVedga))

			case "mc14sch1p":
				fmt.Printf("%s\n", d.WriteField(c, mc14sCh1Vedp))
			case "lc14sch1p":
				fmt.Printf("%s\n", d.WriteField(c, lc14sCh1Vedp))
			case "bde2c4sch1p":
				fmt.Printf("%s\n", d.WriteField(c, bde2c4sCh1Vedp))
			case "bde4c4sch1p":
				fmt.Printf("%s\n", d.WriteField(c, bde4c4sCh1Vedp))
			default:
			}
		case "vsn":
			//write serial number to vendor extension field
			_, rawData := d.DumpProm()
			for i := uint(0 + 11); i < uint(len(rawData)); {
				tlv, tlen := rawData[i], uint(rawData[i+1])
				w := rawData[i+2 : i+2+tlen]
				if tlv == 0xfd {
					w = append(w, 0x54)
					w = append(w, uint8(len(v)))
					for j := uint(0); j < uint(len(v)); j++ {
						w = append(w, v[j])
					}
					d.DeleteField("fd")
					d.WriteField("fd", w)
					fmt.Printf("added serial number to vendor extension\n")
					break
				}
				i += 2 + tlen
			}
		case "pen4":
			//change pen from 2-bytes to 4-bytes
			_, rawData := d.DumpProm()
			for i := uint(0 + 11); i < uint(len(rawData)); {
				tlv, tlen := rawData[i], uint(rawData[i+1])
				w := rawData[i+2 : i+2+tlen]
				if tlv == 0xfd {
					data := []byte{0, 0}
					w = append(data, w...)
					d.DeleteField("fd")
					d.WriteField("fd", w)
					fmt.Printf("added 0x0000 to pen\n")
					break
				}
				i += 2 + tlen
			}
		case "fdrev":
			//overwrite hw revision field
			_, rawData := d.DumpProm()
			for i := uint(0 + 11); i < uint(len(rawData)); {
				tlv, tlen := rawData[i], uint(rawData[i+1])
				w := rawData[i+2 : i+2+tlen]
				if tlv == 0xfd {
					for x := uint(4); x < uint(len(w)); {
						wtlv, wtlen := w[x], uint(w[x+1])
						if wtlv == 0x52 && wtlen == 0x1 {
							h, _ := strconv.ParseUint(v, 16, 64)
							b := []byte{byte(h & 0xff)}
							w[x+2] = b[0]
							log.Print(b[0])
							break
						}
						x += 2 + wtlen
					}
					d.DeleteField("fd")
					d.WriteField("fd", w)
					fmt.Printf("changed vendor extension device revision\n")
					break
				}
				i += 2 + tlen
			}
		case "fdpn":
			//overwrite part number field
			_, rawData := d.DumpProm()
			for i := uint(0 + 11); i < uint(len(rawData)); {
				tlv, tlen := rawData[i], uint(rawData[i+1])
				w := rawData[i+2 : i+2+tlen]
				if tlv == 0xfd {
					for x := uint(4); x < uint(len(w)); {
						wtlv, wtlen := w[x], uint(w[x+1])
						if wtlv == 0x53 && wtlen == 0xe {
							if uint(len(v)) == wtlen {
								for y := uint(0); y < wtlen; y++ {
									w[x+2+y] = v[y]
								}
							}
							break
						}
						x += 2 + wtlen
					}
					d.DeleteField("fd")
					d.WriteField("fd", w)
					fmt.Printf("changed vendor extension device part number\n")
					break
				}
				i += 2 + tlen
			}
		default:
			//write any field with value
			fmt.Printf("%s\n", d.WriteField(c, vByte))
		}
	} else if writeField && len(argF) == 2 {
		c = argF[1]

		switch c {
		case "onie":
			//write onie header
			//fmt.Printf("onie: c: %v, v: %v\n", c, vByte)
			fmt.Printf("%s\n", d.WriteField(c, vByte))
		case "crc":
			//recalculate crc32 and update crc field
			fmt.Printf("%s\n", d.CalcCrc())
		case "addcrc":
			//add crc field (debug tool to fix a invalid format)
			fmt.Printf("%s\n", d.AddCrc())
		case "copy":
			//copy host prom to bmc prom and update vendor extension field
			var rawData []byte
			if !x86 {
				d.BusAddress = 0x51
				gpioSet("CPU_TO_MAIN_I2C_EN", true)
				time.Sleep(10 * time.Millisecond)
				_, rawData = d.DumpProm()
				gpioSet("CPU_TO_MAIN_I2C_EN", false)
				time.Sleep(10 * time.Millisecond)
				d.BusAddress = 0x55
			} else {
				d.BusAddress = 0x55
				_, rawData = d.DumpProm()
				d.BusAddress = 0x51
				gpioSet("CPU_TO_MAIN_I2C_EN", true)
				time.Sleep(10 * time.Millisecond)
			}

			for i := uint(0 + 11); i < uint(len(rawData)); {
				tlv, tlen := rawData[i], uint(rawData[i+1])
				v := rawData[i+2 : i+2+tlen]
				if tlv == 0xfd {
					for j := uint(4); j < uint(len(v)); {
						tlv, tlen := v[j], uint(v[j+1])
						if tlv == 0x53 {
							var ppnNew string
							if !x86 {
								ppnNew = ppnTor1
							} else {
								ppnNew = ppnBde4c
							}
							for k := uint(0); k < uint(len(ppnNew)); k++ {
								rawData[i+2+j+2+k] = ppnNew[k]
							}
							j += 2 + tlen
							continue
						} else if tlv == 0x51 {
							var boardTypeNew uint8
							if !x86 {
								boardTypeNew = boardTypeTor1
							} else {
								boardTypeNew = boardTypeBde4c
							}
							rawData[i+2+j+2] = boardTypeNew
							j += 2 + tlen
							continue
						}
						j += 2 + tlen
					}
					break
				}
				i += 2 + tlen
			}
			d.CopyAll(rawData)
			fmt.Printf("data copied\n")

			d.CalcCrc()
			fmt.Printf("crc updated\n")

			if x86 {
				gpioSet("CPU_TO_MAIN_I2C_EN", false)
			}

		default:
		}
	} else if delField && len(argF) == 2 {
		//delete a field
		c = argF[1]
		fmt.Printf("%s\n", d.DeleteField(c))
	} else {
		fmt.Printf("Invalid or insufficient arguments\n")
	}

	if flagF.ByName["-x86"] {
		gpioSet("CPU_TO_MAIN_I2C_EN", false)
	}
	return nil
}
