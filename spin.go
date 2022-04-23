// Copyright 2022 Santosh Heigrujam. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

// Helper functions to print a string in a particular color
// using ASCII code.

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func black(s string) string {
	return "\033[30m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

// spin represents the spinner  displayed while the
// processing is going on. It comprises of a symbol and a status.
// e.g. @ connecting...
// Multiple symbols and colors can be use to achieve a dynamic effect.
type spin struct {
	symbols      []string
	symbolColors []func(string) string
	statusFunc   func() string
	durationFunc func() bool
}

// do prints a symbol and a status. The symbol can be printed with different colors
// at each update to give a dynamic effect.
func (s spin) do() {
	for i := 0; s.durationFunc(); i++ {
		j, k := i%len(s.symbolColors), i%len(s.symbols)
		fmt.Printf("\r%s  %s", s.symbolColors[j](s.symbols[k]), s.statusFunc())
		time.Sleep(300 * time.Millisecond)
	}
	fmt.Print("\r\033[K")
}
