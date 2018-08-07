// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	homeURLa string = "http://172.17.2.33:8081/register/bootstatus/12345678"
	homeURLb string = "http://172.17.2.33:8081/register/invadercfg/12345678"
	homeURLc string = "http://172.17.2.33:8081/register/invader/12345678"
	homeURLd string = "http://172.17.2.33:8081/register/unregister/12345678"
)

func postita(status string) {
	fmt.Println("**** register/bootstatus ****\n")
	v := url.Values{}
	v.Set("msg", status)
	//v.Set("b", "beta")
	//v.Set("g", "10")

	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", homeURLa, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}
	fmt.Printf("read resp.Body successfully:\n%v\n", string(data))
}

func postitb() {
	fmt.Println("**** register/invadercfg ****\n")
	v := url.Values{}
	v.Set("msg", "/register/invadercfg/12345678")
	//v.Set("b", "beta")
	//v.Set("g", "10")

	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", homeURLb, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}
	fmt.Printf("read resp.Body successfully:\n%v\n", string(data))
}

func postitc() {
	fmt.Println("**** register/invader ****\n")
	v := url.Values{}
	v.Set("msg", "/register/invader/12345678")
	//v.Set("b", "beta")
	//v.Set("g", "10")

	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", homeURLc, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}
	fmt.Printf("read resp.Body successfully:\n%v\n", string(data))
}

func postitd() {
	fmt.Println("**** register/unregister ****\n")
	v := url.Values{}
	v.Set("msg", "/register/unregister/12345678")
	//v.Set("b", "beta")
	//v.Set("g", "10")

	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", homeURLd, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}
	fmt.Printf("read resp.Body successfully:\n%v\n", string(data))
}
