/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"flag"

	"github.com/kokaq/repl/cmd"
)

func main() {
	addr := flag.String("address", "", "Server address")
	flag.Parse()

	// Fallback to env vars if flags are not set
	add := *addr
	if add == "" {
		add = ":9000" // default fallback
	}

	cmd.NewReplClient(add).Start()
}
