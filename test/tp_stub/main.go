package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// SleepMs indicates for how long the program should sleep before he gets terminated
// The value is in milliseconds.
var SleepMs string

// Return if specified as build flag will be the output of the command.
var Return string

// Version indicates which version of binary should be used.
var Version string

const versionFlag = "version"

func main() {
	if Version == "v2" {
		if os.Args[1] == versionFlag {
			os.Exit(0)
		}
	} else {
		if os.Args[1] == versionFlag {
			os.Exit(123)
		}
	}

	if Return != "" {
		fmt.Println(Return)
	} else {
		fmt.Println(os.Args[1:])
	}

	if SleepMs != "" {
		sleep()
	}
}

func sleep() {
	sleepInt, err := strconv.ParseInt(SleepMs, 0, 64)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("Sleep for: " + SleepMs + "ms")
	time.Sleep(time.Duration(sleepInt) * time.Millisecond)
}
