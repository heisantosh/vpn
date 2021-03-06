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
	// States of a VPN connection
	_stateConnected    = "Connected"
	_stateDisconnected = "Disconnected"
	_stateNotFound     = "NotFound"
)

// VPN has a list of profiles.
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

// getProfiles finds the connection states of the VPN profiles.
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

// getProfileStates finds the state of the connection of the given profile.
func (vpn VPN) getProfileState(profile string) string {
	states := vpn.getProfileStates()
	for k, v := range states {
		if profile == vpn.profiles[k] {
			return v
		}
	}
	return _stateNotFound
}

func (vpn VPN) which() {
	states := vpn.getProfileStates()
	for k, v := range states {
		if v == _stateConnected {
			fmt.Println(vpn.profiles[k])
			return
		}
	}
}

// disconnect disconnects all VPN connections.
func (vpn VPN) disconnect() {
	err := sh.Command("osascript", "-e", "tell application \"Viscosity\" to disconnectall").Run()
	if err != nil {
		log.Fatal(err)
	}
}

// connect connects the VPN using the given profile.
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

	// Disconnect any existing VPN to connect with new profile.
	vpn.disconnect()
	for strings.Contains(strings.Join(vpn.getProfileStates(), ","), _stateConnected) {
		time.Sleep(1 * time.Second)
	}

	if err := sh.Command("osascript", "-e", "tell application \"Viscosity\" to connect \""+profile+"\"").Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}

// listProfiles prints out the list of available VPN profiles and the connection states.
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

	// Command to list out the VPN profiles available and the connection states.
	listCmd := &cli.Command{
		Name:  "list",
		Usage: "list the available VPN profiles and see the corresponding states",
		UsageText: "vpn list",
		Action: func(ctx *cli.Context) error {
			vpn.listProfiles()
			return nil
		},
	}

	// Command to get currently connected VPN profile.
	whichCmd := &cli.Command{
		Name:  "which",
		Usage: "find the currently connectd VPN profile",
		UsageText: "vpn which",
		Action: func(ctx *cli.Context) error {
			vpn.which()
			return nil
		},
	}

	// Command to disconnect the VPN connection.
	start := time.Now()
	offCmd := &cli.Command{
		Name:  "off",
		Usage: "disconnect the VPN connection",
		UsageText: "vpn off",
		Action: func(ctx *cli.Context) error {
			offSpin := spin{
				// Nerdfont symbol nf-mdi-lan_disconnect
				symbols:      []string{"???", "???", "???"},
				symbolColors: []func(string) string{red, black, yellow},
				statusFunc:   func() string { return "Disonnecting" },
				durationFunc: func() bool {
					// At the least wait for 2 seconds to complete the disconnect command.
					for time.Since(start) < 2 * time.Second || strings.Contains(strings.Join(vpn.getProfileStates(), ","), _stateConnected) {
						return true
					}
					return false
				},
			}
			vpn.disconnect()
			offSpin.do()
			fmt.Printf("%s  VPN is disconnected\n", red("???"))
			return nil
		},
	}

	// Command to connect the VPN using the given profile.
	onCmd := &cli.Command{
		Name:      "on",
		Usage:     "connect VPN for the given profile",
		UsageText: "vpn on <profile>",
		Action: func(ctx *cli.Context) error {
			profile := ctx.Args().Get(0)
			onSpin := spin{
				// Nerdfont symbol nf-mdi-lan_connect
				symbols:      []string{"???", "???", "???"},
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
			fmt.Printf("%s  VPN is connected to %s\n", green("???"), profile)
			return nil
		},
	}

	app := &cli.App{
		Usage:     "A CLI wrapper for the Viscosity application in MacOS based on the Applescript API.",
		UsageText: "vpn command [argument]",
		Commands: []*cli.Command{
			listCmd,
			whichCmd,
			offCmd,
			onCmd,
		},
	}

	app.Run(os.Args)
}
