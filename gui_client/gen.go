//+build generate

package main

import (
	"fmt"
	"os"

	"github.com/zserge/lorca"
)

func main() {
	if err := lorca.Embed("main", "assets.go", "www"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
