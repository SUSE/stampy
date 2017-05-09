package main

import (
	"fmt"
	"os"

	"github.com/SUSE/stampy"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("Usage: %s csvfile origin series event\n", os.Args[0])
		return
	}

	stampy.Stamp(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
}
