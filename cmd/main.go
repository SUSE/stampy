package main

import (
	"fmt"
	"os"

	"github.com/hpcloud/stampy"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s csvfile series event\n", os.Args[0])
		return
	}

	stampy.Stamp(os.Args[1], os.Args[2], os.Args[3])
}
