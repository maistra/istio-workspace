// Test double that echoes passed arguments and flags.
// This is handy in validating passed arguments to target binary
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// SleepMs indicates for how long the program should sleep before he gets terminated
// The value is in milliseconds or "infinite" which will result in endless loop.
var SleepMs string

// Echo if specified as build flag will be the output of the command.
var Echo string

func main() {
	if Echo != "" {
		fmt.Println(Echo)
	} else {
		fmt.Println(os.Args)
	}

	if SleepMs != "" {
		if SleepMs == "infinite" {
			infiniteLoop()
		} else {
			sleep()
		}
	}
}

func infiniteLoop() {
	println("...infinite loop")
	select {}
}

func sleep() {
	sleepInt, err := strconv.ParseInt(SleepMs, 0, 64)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("Sleep for: " + SleepMs)
	time.Sleep(time.Duration(sleepInt) * time.Millisecond)
}
