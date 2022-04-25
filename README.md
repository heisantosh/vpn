# vpn

⛱️

`vpn` is a convenient CLI wrapper for the MacOS VPN application `Viscosity`.

It allows to 
* connect a given VPN profile
* disconnect
* list out the available profiles
* find the currently connected profile

It is based on the Applescript API for Viscosity https://www.sparklabs.com/support/kb/article/controlling-viscosity-with-applescript-mac/.

## Prerequisites
* OS : `MacOS`
* Application: `Viscosity`
* Profiles are set up in `Viscosity`

## Installation
### Using `Go`
```bash
$ go install github.com/heisantosh/vpn@latest
```

### Download the binary
Prebuilt binary can be downloaded from here https://github.com/heisantosh/vpn/releases.

## Usage
```bash
$ vpn help
NAME:
   vpn - A CLI wrapper for the Viscosity application in MacOS based on the Applescript API.

USAGE:
   vpn command [argument]

COMMANDS:
   list     list the available VPN profiles and see the corresponding states
   which    find the currently connectd VPN profile
   off      disconnect the VPN connection
   on       connect VPN for the given profile
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```