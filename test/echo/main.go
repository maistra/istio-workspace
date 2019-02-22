// Test double that echoes passed arguments and flags.
// This is handy in validating passed arguments to target binary
package main

import (
	"os"
	"strconv"
	"time"
)
import "fmt"

var SleepMs string

func main() {
	fmt.Println(os.Args)
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
