// Test double that echoes passed arguments and flags.
// This is handy in validating passed arguments to target binary
package main

import "os"
import "fmt"

func main() {
	fmt.Println(os.Args)
}
