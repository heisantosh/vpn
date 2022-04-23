// Copyright 2022 Santosh Heigrujam. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

// vpn is a convenient CLI wrapper around the Viscosity Applescript API.
// It allows to connect/disconnect the VPN and listing the available profiles.
// Viscosity Applescript API: https://www.sparklabs.com/support/kb/article/controlling-viscosity-with-applescript-mac/

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/urfave/cli/v2"
)

const (
	_stateConnected    = "Connected"
	_stateDisconnected = "Disconnected"
	_stateNotFound     = "NotFound"
)

type VPN struct {
	profiles []string
}

func NewVPN() *VPN {
	vpn := VPN{}
	out, err := sh.Command("osascript", "-e", "tell application \"Viscosity\"", "-e", "name of (every connection)", "-e", "end tell").Output()
	if err != nil {
		log.Fatal("err")
	}
	// Ouput format of above script is - profile1, profile2, profile3
	v := strings.Trim(string(out), "\n")
	vpn.profiles = strings.Split(strings.ReplaceAll(v, ",", ""), " ")
	return &vpn
}

func (vpn VPN) getProfileStates() []string {
	// We take advantage of the fact that the order of the profile names returned earlier corresponds
	// to the order of the state of the profiles returned.
	out, err := sh.Command("osascript", "-e", "tell application \"Viscosity\"", "-e", "state of (every connection)", "-e", "end tell").Output()
	if err != nil {
		log.Fatal("err")
	}
	// Ouput format of above script is - Disconnected, Connected, Disconnected
	v := strings.Trim(string(out), "\n")
	states := strings.Split(strings.ReplaceAll(v, ",", ""), " ")
	return states
}

func (vpn VPN) getProfileState(profile string) string {
	states := vpn.getProfileStates()
	for k, v := range states {
		if profile == vpn.profiles[k] {
			return v
		}
	}
	return _stateNotFound
}

func (vpn VPN) disconnect() {
	err := sh.Command("osascript", "-e", "tell application \"Viscosity\" to disconnectall").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (vpn VPN) connect(profile string) error {
	var valid bool
	for _, v := range vpn.profiles {
		if v == profile {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Printf("%s is an invalid VPN profile\n", profile)
		fmt.Println("Run `vpn list` to see the list of available profiles")
		return errors.New("invalid VPN profile: " + profile)
	}

	if err := sh.Command("osascript", "-e", "tell application \"Viscosity\" to connect \""+profile+"\"").Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (vpn VPN) listProfiles() {
	states := vpn.getProfileStates()
	for k, v := range vpn.profiles {
		s := red(states[k])
		if states[k] == _stateConnected {
			s = green(states[k])
		}
		fmt.Printf("%-30s  %s\n", v, s)
	}
}

func main() {
	vpn := NewVPN()

	listCmd := &cli.Command{
		Name:  "list",
		Usage: "list the available VPN profiles and see the corresponding states",
		Action: func(ctx *cli.Context) error {
			vpn.listProfiles()
			return nil
		},
	}

	start := time.Now()
	offCmd := &cli.Command{
		Name:  "off",
		Usage: "disconnect the VPN connection",
		Action: func(ctx *cli.Context) error {
			offSpin := spin{
				// Nerdfont symbol nf-mdi-lan_disconnect
				symbols:      []string{"", "", ""},
				symbolColors: []func(string) string{red, black, yellow},
				statusFunc:   func() string { return "Disonnecting" },
				// Disconnecting happens almost instantly but we want to
				// wait a bit to spin for dramatic effect ;) 😜
				durationFunc: func() bool {
					return time.Since(start) < 4*time.Second
				},
			}
			vpn.disconnect()
			offSpin.do()
			fmt.Printf("%s  VPN is disconnected\n", red(""))
			return nil
		},
	}

	onCmd := &cli.Command{
		Name:  "on",
		Usage: "connect VPN for the given profile",
		UsageText: "vpn on <profile>",
		Action: func(ctx *cli.Context) error {
			profile := ctx.Args().Get(0)
			onSpin := spin{
				// Nerdfont symbol nf-mdi-lan_connect
				symbols:      []string{"", "", ""},
				symbolColors: []func(string) string{green, black, yellow},
				statusFunc:   func() string { return vpn.getProfileState(profile) },
				durationFunc: func() bool {
					state := vpn.getProfileState(profile)
					v := state != _stateConnected
					return v
				},
			}
			if err := vpn.connect(profile); err != nil {
				return err
			}
			onSpin.do()
			fmt.Printf("%s  VPN is connected to %s\n", green(""), profile)
			return nil
		},
	}

	app := &cli.App{
		Usage: "A CLI wrapper for the Viscosity application in MacOS based on the Applescript API.",
		Commands: []*cli.Command{
			listCmd,
			offCmd,
			onCmd,
		},
	}

	app.Run(os.Args)
}
