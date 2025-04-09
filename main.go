package main

import (
	"fmt"
	"os"

	"github.com/ezra-sullivan/net-sniff/pkg/cmd"
)

func main() {
	if err := cmd.NewNetSniffCommand().Execute(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
