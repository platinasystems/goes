// This file contains Data Structures and Methods required for parsing msconfig.cfg file.
// msconfig.cfg file persists mirror and sflow related configuration in form of destination profiles, mirror profiles and sflow profiles.

package vnet

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type MirrorConfigData struct {
	Name string
	Src string
	Dst string
	Type string
	Active string
}

type DestConfigData struct {
	Name string
	Encap string
	Agent_IP string
	Binded_to string
	Used_by string
}

type SflowConfigData struct {
	Name	string
	Src		string
	Active	string
	Cpu		string
	Mirror  string
	Rate 	string
	Dst 	string
}

type MirrorConfigFileData struct {
	Mirror []MirrorConfigData
}

type DestConfigFileData struct {
	Dest []DestConfigData
}

type SflowConfigFileData struct {
	Sflow []SflowConfigData
}

type ConfigFileData struct{
	MirrorCfgFileData MirrorConfigFileData
	DestCfgFileData DestConfigFileData
	SflowCfgFileData SflowConfigFileData
}

func (cfd *ConfigFileData) ReadConfigFile() (err error){
	var fileData []byte

	// Open file
	file, err := os.Open("/etc/goes/msconfig.cfg")
	if err != nil {
		return err
	}

	// Close file at the end of function
	defer file.Close()

	// Read config file
	fileData, err = ioutil.ReadFile(file.Name())
	if err != nil {
		return err
	}

	// Parse mirror config from file fileData
	yaml.Unmarshal(fileData, &cfd.MirrorCfgFileData)
	if err != nil {
		err = errors.New("yaml unmarshal failed while reading mirror config.")
		return err
	}

	// Parse destination config from file fileData
	yaml.Unmarshal(fileData, &cfd.DestCfgFileData)
	if err != nil {
		err = errors.New("yaml unmarshal failed while reading destination config.")
		return err
	}

	// Parse sflow config from file fileData
	yaml.Unmarshal(fileData, &cfd.SflowCfgFileData)
	if err != nil {
		err = errors.New("yaml unmarshal failed while reading sflow config.")
		return err
	}

	return err
}

func (cfd *ConfigFileData) WriteConfigFile() (err error){

	// Create file
	file, err := os.Create("/etc/goes/msconfig.cfg")
	if err != nil {
		fmt.Println(err)
	}

	// Close file at the end of function
	defer file.Close()

	// Write to file in yaml format

	mirrorYamlData, _ := yaml.Marshal(&cfd.MirrorCfgFileData.Mirror)
	destYamlData, _ :=   yaml.Marshal(&cfd.DestCfgFileData.Dest)
	sflowYamlData,_ :=  yaml.Marshal(&cfd.SflowCfgFileData.Sflow)

	fileData := ""

	if len(cfd.DestCfgFileData.Dest) > 0 {
		fileData += "dest:\n" + (string)(destYamlData)
	}

	if len(cfd.MirrorCfgFileData.Mirror) > 0 {
		fileData += "\n\n" + "mirror:\n" + (string)(mirrorYamlData)
	}

	if len(cfd.SflowCfgFileData.Sflow) > 0 {
		fileData += "\n\n" + "sflow:\n" + (string)(sflowYamlData)
	}

	file.Write(([]byte)(fileData))

	return err
}